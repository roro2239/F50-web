package main

import (
	"bytes"
	"encoding/hex"
	"os/exec"
	"regexp"
	"strconv"
)

func parseATCommand(input string) string {
	// 使用正则表达式提取16进制数值
	hexPattern := regexp.MustCompile(`\b[0-9a-fA-F]{8}\b`)
	matches := hexPattern.FindAllString(input, -1)

	var result bytes.Buffer

	for _, hexStr := range matches {
		// 将16进制字符串转换为字节数组
		byteVal, err := hex.DecodeString(hexStr)
		if err != nil {
			continue
		}

		// 假设字节顺序是小端（低字节在前），根据实际情况调整
		for i := len(byteVal) - 1; i >= 0; i-- {
			c := byteVal[i]
			if c >= 32 && c <= 126 { // 可打印的ASCII字符
				result.WriteByte(c)
			} else if c == 0x0d || c == 0x0a { // 处理回车和换行
				result.WriteByte(c)
			}
		}
	}

	return result.String()
}
func processATCommand(channel string, command string) string {
	quotedCmd := strconv.Quote("sendAt " + channel + " " + command)
	cmdString := `/system/bin/service call vendor.sprd.hardware.log.ILogControl/default 1 s16 "miscserver" s16 ` + quotedCmd

	// 执行命令
	cmd := exec.Command(shellPath(), "-c", cmdString)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "命令执行失败: " + err.Error()
	}
	parsed := parseATCommand(string(out))
	return parsed
}
