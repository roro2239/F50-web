package main

import (
	"os"

	ini "gopkg.in/ini.v1"
)

var cfgInfo *ini.File

func CONF() {
	initRuntime()
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// 创建一个空的配置文件
		cfg := ini.Empty()
		// 添加默认的全局配置
		section := cfg.Section("")
		section.Key("adbPort").SetValue("5555")
		section.Key("adbPortForward").SetValue("5556")
		section.Key("smsForwardCmd").SetValue("")

		// 将配置写入文件
		cfg.SaveTo(configFile)
		cfgInfo = cfg
		return
	}

	// 读取配置文件
	cfg, err := ini.Load(configFile)
	if err != nil {
		printf("获取配置文件失败: %v\n", err)
		return
	}
	// 获取配置项
	// adbPort := cfgInfo.Section("").Key("adbPort").MustInt(5555)
	// adbPortForward := cfgInfo.Section("").Key("adbPortForward").MustInt(5556)
	cfgInfo = cfg
}
