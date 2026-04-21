package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var ZTELD string = "204b95bc-6648-4da9-ac29-3bb47aab3aa7"

type SessionInfo struct {
	ExpiresAt time.Time
	ClientIP  string
}

var (
	sessionStore = make(map[string]SessionInfo)
	storeMutex   sync.RWMutex
)
var webSubFS fs.FS

func startServer() {
	go func() {
		webSubFS, _ = fs.Sub(staticFS, "web")

		http.Handle("/",
			http.FileServer(http.FS(webSubFS)),
		)

		http.HandleFunc("/goform/", goformHandler)

		http.HandleFunc("/atwat/", atwatAPI)

		http.HandleFunc("/iptablesAPI/", iptablesAPI)

		print("后台端口 :8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			print("监听端口8081失败:", err)
			return
		}
	}()
}

var CPUusage float64
var MEMusage float64
var MAXtemp float64
var Battery int
var BatteryCharging bool

func atwatAPI(w http.ResponseWriter, r *http.Request) {
	// 获取请求路径
	path := r.URL.Path
	if !isLogin(w, r) {
		json.NewEncoder(w).Encode(map[string]string{"loginfo": "no"})
		return
	}
	var ttydSock = func(sockPath string) {
		// 升级当前连接为 WebSocket
		var upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			Subprotocols: []string{"tty"},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		// 使用 UNIX 域套接字连接到目标 WebSocket 服务
		dialer := &websocket.Dialer{
			NetDial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		}

		// 创建请求头，添加子协议
		header := http.Header{}
		header.Add("Sec-WebSocket-Protocol", "tty")
		targetURL := "ws://localhost/ws" // 使用 localhost 作为虚拟目标
		targetConn, _, err := dialer.Dial(targetURL, header)
		if err != nil {
			conn.Close()
			return
		}

		// 在客户端和目标 WebSocket 服务之间转发消息
		go func() {
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					break
				}
				if err := targetConn.WriteMessage(messageType, message); err != nil {
					break
				}
			}
			targetConn.Close()
			conn.Close()
		}()

		for {
			messageType, message, err := targetConn.ReadMessage()
			if err != nil {
				break
			}
			if err := conn.WriteMessage(messageType, message); err != nil {
				break
			}
		}
		targetConn.Close()
		conn.Close()
	}
	type AdbPortInfo struct {
		AdbPort     int `json:"adbPort"`
		AdbPortSafe int `json:"adbPortSafe"`
	}
	type SMSforInfo struct {
		SmsForwardCmd string `json:"smsForwardCmd"`
	}
	switch path {
	case "/atwat/sms_forward/status":
		smsForwardCmd := cfgInfo.Section("").Key("smsForwardCmd").String()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SMSforInfo{smsForwardCmd})

	case "/atwat/sms_forward/test":
		if r.Method == http.MethodPost {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			dateTime, err := time.Parse("06,01,02,15,04,05,-0700", req["time"].(string))
			if err != nil {
				w.Write([]byte("时间格式化错误:" + err.Error()))
				return
			}
			dt := dateTime.Format("2006/01/02 15:04:05")
			msg := req["content"].(string)
			sender := req["sender"].(string)
			smsForwardCmd := cfgInfo.Section("").Key("smsForwardCmd").String()
			command := fmt.Sprintf(`
sms_time=$(echo "%s" | base64 --decode)
sms_sender=$(echo "%s" | base64 --decode)
sms_content=$(echo "%s" | base64 --decode)
sms_time_query_escape=$(echo "%s" | base64 --decode)
sms_sender_query_escape=$(echo "%s" | base64 --decode)
sms_content_query_escape=$(echo "%s" | base64 --decode)
`, encodeBase64(dt), encodeBase64(sender), encodeBase64(msg), encodeBase64(url.QueryEscape(dt)), encodeBase64(url.QueryEscape(sender)), encodeBase64(url.QueryEscape(msg))) + smsForwardCmd + "\n"
			ss := NewCommandExecutor(command)
			ss.Start()

			w.Write([]byte("ok"))
		} else {
			http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		}
	case "/atwat/sms_forward/cmd":
		if r.Method == http.MethodPost {
			var req SMSforInfo
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfgInfo.Section("").Key("smsForwardCmd").SetValue(req.SmsForwardCmd)
			cfgInfo.SaveTo(configFile)
			w.Write([]byte("ok"))
		} else {
			http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		}
	case "/atwat/adbPort":
		adbPort := cfgInfo.Section("").Key("adbPort").MustInt(5555)
		adbPortSafe := cfgInfo.Section("").Key("adbPortForward").MustInt(5556)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AdbPortInfo{AdbPort: adbPort, AdbPortSafe: adbPortSafe})
	case "/atwat/adbPortSet":
		if r.Method == http.MethodPost {
			var req AdbPortInfo
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// 这里可以添加处理逻辑
			processAdbPort(req.AdbPort, req.AdbPortSafe)

			adbPort := cfgInfo.Section("").Key("adbPort").MustInt(5555)
			adbPortSafe := cfgInfo.Section("").Key("adbPortForward").MustInt(5556)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(AdbPortInfo{AdbPort: adbPort, AdbPortSafe: adbPortSafe})
		} else {
			http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		}
	case "/atwat/atcmd":
		type ATCommandRequest struct {
			Channel string `json:"channel"`
			AT      string `json:"AT"`
		}
		type ATCommandResponse struct {
			Result string `json:"result"`
		}
		if r.Method == http.MethodPost {
			var req ATCommandRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// 这里可以添加处理AT命令的逻辑
			result := processATCommand(req.Channel, req.AT)
			// 返回结果
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ATCommandResponse{Result: result})
		} else {
			http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		}
	case "/atwat/sysinfo":
		// Create a JSON response
		response := map[string]interface{}{
			"CPUusage": CPUusage,
			"MEMusage": MEMusage,
			"MAXtemp":  MAXtemp,
			"Battery":  Battery,
		}
		jsonData, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	case "/atwat/ttyd/":
		// 读取嵌入的 HTML 文件到内存
		data, err := fs.ReadFile(webSubFS, "atwat/ttyd/index.html")
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// 转换为 ReadSeeker
		reader := bytes.NewReader(data)

		// 设置 Content-Type
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// 正确调用 ServeContent
		http.ServeContent(w, r, "index.html", time.Time{}, reader)

	case "/atwat/ttyd/log":
		w.Write([]byte(strings.Join(ttydCmd.GetLogs(), "")))
	case "/atwat/ttyd/status":
		w.Write([]byte(ttydCmd.GetStatus()))
	case "/atwat/ttyd/stop":
		err := ttydCmd.Stop()
		if err == nil {
			w.Write([]byte("ok"))
		} else {
			w.Write([]byte("err:" + err.Error()))
		}
	case "/atwat/ttyd/start":
		err := ttydCmd.Start()
		if err == nil {
			w.Write([]byte("ok"))
		} else {
			w.Write([]byte("err:" + err.Error()))
		}
	case "/atwat/ttyd/token":
		err := ttydCmd.Start()
		if err != nil && err.Error() != "command is already running" {
			print("启动ttyd失败", err)
			return
		}
		w.Write(nil)
	case "/atwat/ttyd/ws":
		ttydSock(runtimePath("ttyd-termux.sock"))

	case "/atwat/ttyd-root/":
		// 读取嵌入的 HTML 文件到内存
		data, err := fs.ReadFile(webSubFS, "atwat/ttyd/index.html")
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// 转换为 ReadSeeker
		reader := bytes.NewReader(data)

		// 设置 Content-Type
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// 正确调用 ServeContent
		http.ServeContent(w, r, "index.html", time.Time{}, reader)

	case "/atwat/ttyd-root/token":
		rootShell(`SOCK_FILE="`+runtimePath("ttyd-root.sock")+`"
TTYD_BIN="`+bundledBinary("ttyd")+`"
if [ ! -x "$TTYD_BIN" ]; then
    TTYD_BIN="$(command -v ttyd 2>/dev/null || true)"
fi
if [ -n "$TTYD_BIN" ] && [ -x "$TTYD_BIN" ]; then
    SHELL_BIN="`+shellPath()+`"
    if ! pgrep -f "$TTYD_BIN -q -W -i $SOCK_FILE $SHELL_BIN" > /dev/null; then
        PATH="`+binDir+`:$PATH" "$TTYD_BIN" -q -W -i "$SOCK_FILE" "$SHELL_BIN" &
        sleep 1s && chmod 777 "$SOCK_FILE"
    fi
fi
`, false)
		w.Write(nil)
	case "/atwat/ttyd-root/ws":
		ttydSock(runtimePath("ttyd-root.sock"))
	default:
		http.NotFound(w, r)
	}
}

func isIPLoggedIn(ip string) bool {
	storeMutex.RLock()
	defer storeMutex.RUnlock()
	now := time.Now()
	for _, info := range sessionStore {
		if info.ClientIP == ip && now.Before(info.ExpiresAt) {
			return true
		}
	}
	return false
}

// extractClientIP 提取客户端的 IP 地址
func extractClientIP(addr string) string {

	lastColonIndex := strings.LastIndex(addr, ":")
	if lastColonIndex == -1 || lastColonIndex == len(addr)-1 {

	} else {
		addr = addr[:lastColonIndex]
	}
	if strings.HasPrefix(addr, "[") && strings.HasSuffix(addr, "]") {
		addr = strings.TrimPrefix(addr, "[")
		addr = strings.TrimSuffix(addr, "]")
	}
	return addr
}
func isLogin(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("token")
	if err == nil {
		storeMutex.RLock()
		sessionInfo, exists := sessionStore[cookie.Value]
		storeMutex.RUnlock()
		if exists && !time.Now().After(sessionInfo.ExpiresAt) {
			newExpiresAt := time.Now().Add(1 * time.Hour)
			storeMutex.Lock()
			sessionStore[cookie.Value] = SessionInfo{
				ExpiresAt: newExpiresAt,
				ClientIP:  extractClientIP(r.RemoteAddr),
			}

			storeMutex.Unlock()

			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    cookie.Value,
				Expires:  newExpiresAt,
				Path:     "/",
				Secure:   false,
				HttpOnly: true,
			})
			return true
		}
	}
	return false
}

func generateSessionToken() string {
	rand.Seed(time.Now().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func goformHandler(w http.ResponseWriter, r *http.Request) {
	// 获取请求的路径
	path := r.URL.Path
	if !strings.HasPrefix(path, "/goform/") {
		//http.Error(w, "Invalid path", http.StatusNotFound)
		return
	}

	// 去掉前缀 "/goform/"
	cmd := strings.TrimPrefix(path, "/goform/")

	switch cmd {
	case "goform_get_cmd_process":
		goformGetCmdProcess(w, r)
	case "goform_set_cmd_process":
		goformSetCmdProcess(w, r)
	default:
		// 转发请求到 http://"+HOST+"/goform/
		forwardRequest(w, r, "http://"+HOST+":8080/goform/"+cmd)
	}
}

func goformGetCmdProcess(w http.ResponseWriter, r *http.Request) {
	data := r.URL.Query().Encode()
	body, err := reqGet("http://" + HOST + ":8080/goform/goform_get_cmd_process?" + data)
	if err != nil {
		//http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		//http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	lg := isLogin(w, r)
	we := map[string]bool{}
	we["LD"] = true
	we["Language"] = true
	for k := range response {
		if k == "LD" {
			ZTELD = fmt.Sprintf("%v", response[k])
		}
		if _, ok := we[k]; !lg && !ok {
			delete(response, k)
			continue
		}
		if k == "battery_charging" {
			response[k] = BatteryCharging
		}
		if k == "battery_value" {
			response[k] = Battery
		}
		if k == "battery_vol_percent" {
			response[k] = Battery
		}

	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func goformSetCmdProcess(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		//http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	logi := false
	if v, ok := r.PostForm["goformId"]; ok {
		if pw, ok := r.PostForm["password"]; ok && v[0] == "LOGIN" {
			logi = true
			pwd := ""
			if loginPwd != "" && ZTELD != "" {
				pwd = SHA256(loginPwd + ZTELD) //Wa@9w+YWRtaW4=
			}
			if pwd != "" && len(pw) > 0 && pwd != pw[0] {
				logi = false
			}
		}
	}
	if !isLogin(w, r) && !logi {
		json.NewEncoder(w).Encode(map[string]string{"loginfo": "no"})
		return
	}
	data := r.PostForm.Encode()
	body, err := reqPost("http://"+HOST+":8080/goform/goform_set_cmd_process", data)
	if err != nil {
		//http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	if logi && string(body) == `{"result":0}` {
		sessionToken := generateSessionToken()

		storeMutex.Lock()
		sessionStore[sessionToken] = SessionInfo{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		storeMutex.Unlock()

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    sessionToken,
			Expires:  time.Now().Add(1 * time.Hour),
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
		})
	}
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		//http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func forwardRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	// 创建一个新的请求，目标是 targetURL
	newReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		//http.Error(w, "Failed to create new request", http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for name, values := range r.Header {
		for _, value := range values {
			newReq.Header.Add(name, value)
		}
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(newReq)
	if err != nil {
		//http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	// 将响应体返回给客户端
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}
