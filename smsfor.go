package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

func init() {
}
func smsForward() {
	go func() {
		time.Sleep(time.Second * 5)
		id := -1
		for {
			time.Sleep(time.Second)
			smsList(&id)
		}
	}()
}
func decodeBase64(encodedStr string) string {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		return ""
	}
	return string(decodedBytes)
}

//编码
func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
func smsList(id *int) {
	uri := "http://" + HOST + ":8080/goform/goform_get_cmd_process?isTest=false&cmd=sms_data_total&page=0&data_per_page=1000"
	body, err := reqGet(uri)
	if err != nil {
		print(err)
		return
	}

	// 定义单条消息的结构体
	type SMSMessage struct {
		Content      string `json:"content"`
		Date         string `json:"date"`
		DraftGroupID string `json:"draft_group_id"`
		ID           string `json:"id"`
		Number       string `json:"number"`
		Tag          string `json:"tag"`
	}
	// 定义响应体结构体
	type SMSResponse struct {
		Messages []SMSMessage `json:"messages"`
	}
	var response SMSResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		print(err)
		return
	}
	maxID := 0
	// 遍历消息列表
	for _, sms := range response.Messages {
		msg := decodeBase64(sms.Content)
		dateTime, err := time.Parse("06,01,02,15,04,05,-0700", sms.Date)
		if err != nil {
			print("时间格式化错误:", err)
			continue
		}
		dt := dateTime.Format("2006/01/02 15:04:05")
		iID, _ := strconv.Atoi(sms.ID)
		if iID > maxID {
			maxID = iID
		}
		if *id == -1 {
			*id = iID
		}
		if iID > *id {
			//转发短信
			if sms.Tag == "2" {
				continue
			}
			sender := sms.Number
			printf("新短信%v, %v, %v", dt, sender, msg)
			smsForwardCmd := cfgInfo.Section("").Key("smsForwardCmd").String()
			command := fmt.Sprintf(`
sms_time=$(echo "%s" | base64 --decode)
sms_sender=$(echo "%s" | base64 --decode)
sms_content=$(echo "%s" | base64 --decode)
sms_time_query_escape=$(echo "%s" | base64 --decode)
sms_sender_query_escape=$(echo "%s" | base64 --decode)
sms_content_query_escape=$(echo "%s" | base64 --decode)
`, encodeBase64(dt), encodeBase64(sender), encodeBase64(msg), encodeBase64(url.QueryEscape(dt)), encodeBase64(url.QueryEscape(sender)), encodeBase64(url.QueryEscape(msg))) + smsForwardCmd + "\n"
			//print(command)
			ss := NewCommandExecutor(command)
			ss.Start()
		}
		break
	}
	*id = maxID
}
