package main

import (
	"io"
	"net"
	"strconv"
)

func adbSafe() {
	go func() {
		// 监听5556端口
		port2 := strconv.Itoa(cfgInfo.Section("").Key("adbPortForward").MustInt(5556))

		listener, err := net.Listen("tcp", ":"+port2)
		if err != nil {
			printf("监听端口%v失败:%v", port2, err)
			return
		}
		defer listener.Close()
		printf("安全的ADB端口 :%v", port2)

		for strconv.Itoa(cfgInfo.Section("").Key("adbPortForward").MustInt(5556)) == port2 {
			// 接受客户端连接
			conn, err := listener.Accept()
			if err != nil {
				//print("Failed to accept connection:", err)
				continue
			}

			// 获取客户端IP
			clientIP := conn.RemoteAddr().(*net.TCPAddr).IP.String()

			// 检查是否在白名单中
			if !isIPLoggedIn(clientIP) {
				//printf("Client %s is not in the whitelist. Connection refused.", clientIP)
				conn.Close()
				continue
			}

			printf("用户 %s 连接到:%v", clientIP, port2)

			// 连接到目标端口5555
			targetConn, err := net.Dial("tcp", HOST+":"+strconv.Itoa(cfgInfo.Section("").Key("adbPort").MustInt(5555)))
			if err != nil {
				//print("Failed to connect to target port 5555:", err)
				conn.Close()
				continue
			}

			// 启动两个协程，双向转发数据
			go forwardData(conn, targetConn)
			go forwardData(targetConn, conn)
		}
	}()
}

// 数据转发函数
func forwardData(src net.Conn, dst net.Conn) {
	defer src.Close()
	defer dst.Close()

	// 将src的数据复制到dst
	io.Copy(dst, src)

}
func processAdbPort(adbPort, adbPortSafe int) {
	sl := `
ip6tables -D INPUT -p tcp --dport ` + strconv.Itoa(cfgInfo.Section("").Key("adbPort").MustInt(5555)) + ` -j DROP 2>/dev/null
ip6tables -I INPUT -p tcp --dport ` + strconv.Itoa(adbPort) + ` -j DROP
setprop service.adb.tcp.port ` + strconv.Itoa(adbPort) + `
stop adbd
start adbd
`
	rootShell(sl, false)
	if cfgInfo.Section("").Key("adbPortForward").MustInt(5555) != adbPortSafe {
		adbSafe()
	}
	cfgInfo.Section("").Key("adbPort").SetValue(strconv.Itoa(adbPort))
	cfgInfo.Section("").Key("adbPortForward").SetValue(strconv.Itoa(adbPortSafe))
	cfgInfo.SaveTo(configFile)
}
