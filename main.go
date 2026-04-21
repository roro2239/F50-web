package main

import (
	"embed"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed web/*
var staticFS embed.FS

var loginPwd string
var onF50 bool

func init() {
	onF50 = true
}
func main() {
	defer func() {
		if r := recover(); r != nil {
			print("主进程异常退出:", fmt.Sprint(r))
		}
	}()
	print("主进程启动")
	initRuntime()
	print("运行目录:", baseDir)
	CONF()
	if cfgInfo == nil {
		print("配置初始化失败")
		return
	}
	initHost()
	startServer()

	if onF50 {
		print("F50模式初始化")
		adbPort := strconv.Itoa(cfgInfo.Section("").Key("adbPort").MustInt(5555))
		if err := rootShell(`
# 删除可能存在的旧规则
iptables -t nat -D PREROUTING -p tcp -m tcp -i br0 -d 192.168.0.1 --dport 80 -j DNAT --to-destination 192.168.0.1:8080
iptables -t nat -D PREROUTING -p tcp -m tcp -i br0 -d 192.168.0.1 --dport 80 -j DNAT --to-destination 192.168.0.1:8081
iptables -t nat -D PREROUTING -p tcp -m tcp -i br0 -d 192.168.0.1 --dport 443 -j DNAT --to-destination 192.168.0.1:8443
iptables -D INPUT -p tcp -s 127.0.0.1 --dport 8080 -j ACCEPT
ip6tables -D INPUT -p tcp --dport 8080 -j DROP 2>/dev/null
ip6tables -D INPUT -p tcp --dport `+adbPort+` -j DROP 2>/dev/null
ip6tables -D INPUT -p tcp --dport 445 -j DROP 2>/dev/null
# 添加新规则
iptables -I INPUT -p tcp -s 127.0.0.1 --dport 8080 -j ACCEPT
iptables -t nat -I PREROUTING -p tcp -m tcp -i br0 -d 192.168.0.1 --dport 80 -j DNAT --to-destination 192.168.0.1:8081
ip6tables -I INPUT -p tcp --dport 8080 -j DROP
ip6tables -I INPUT -p tcp --dport `+adbPort+` -j DROP
ip6tables -I INPUT -p tcp --dport 445 -j DROP
pkill com.zte.web
/system/bin/device_config set_sync_disabled_for_tests persistent
/system/bin/device_config put activity_manager max_phantom_processes 2147483647
settings put global settings_enable_monitor_phantom_procs false
setprop service.adb.tcp.port `+adbPort+`
stop adbd
start adbd
`, false); err != nil {
			print("root初始化失败:", err)
		}
		adbSafe()
		smsForward()
	} else {
		HOST = "192.168.0.1"
		loginPwd = "09C03C6E345EC522F36D345F43BD65C2B2D0A44DED07DC309AACF87BF376B363"
	}

	print("进入登录保活循环")
	for {
		time.Sleep(time.Second * 10)
		if loginPwd == "" {
			continue
		}
		msg, err := getCMD("LD")
		if err != nil {
			print(err)
			continue
		}
		LD, ok := msg["LD"].(string)
		if !ok {
			continue
		}
		pwd := SHA256(loginPwd + LD) //Wa@9w+YWRtaW4=
		data := "goformId=LOGIN&isTest=false&password=" + pwd

		body, err := setCMD(data)
		if err != nil {
			print(err)
			continue
		}
		if string(body) != `{"result":0}` {
			print("登录异常", string(body))
			continue
		}
		time.Sleep(time.Second * 300)
	}

}
