package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

var HOST string = "127.0.0.1"

func reqGet(uri string) (body []byte, err error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	req.Header.Set("Referer", "http://"+HOST+":8080")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}
func reqPost(uri, data string) (body []byte, err error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(data))
	if err != nil {
		return
	}
	req.Header.Set("Referer", "http://"+HOST+":8080")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

func getCMD(commands ...string) (map[string]interface{}, error) {
	cmds := strings.Join(commands, ",")
	uri := "http://" + HOST + ":8080/goform/goform_get_cmd_process?isTest=false&cmd=" + cmds
	if len(commands) > 1 {
		uri += "&multi_data"
	}
	body, err := reqGet(uri)
	if err != nil {
		return nil, err
	}
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// func getCMD(cmds string) (body []byte, err error) {
// 	body, err = reqGet("http://" + HOST + ":8080/goform/goform_get_cmd_process?isTest=false&cmd=" + cmds + "&multi_data=1")
// 	fmt.Println(string(body), err)
// 	return
// }

func setCMD(data string) (body []byte, err error) {
	body, err = reqPost("http://"+HOST+":8080/goform/goform_set_cmd_process", data)
	return
}
func SHA256(txt string) string {
	hashed := sha256.Sum256([]byte(txt))
	hashedHex := hex.EncodeToString(hashed[:])
	return strings.ToUpper(hashedHex)
}
func gsmEncode(text string) string {
	var encoded []byte
	for _, r := range text {
		if r <= '\uFFFF' {
			encoded = append(encoded, byte(r>>8), byte(r))
		} else {
			codePoint := int(r)
			highSurrogate := 0xD800 + ((codePoint - 0x10000) >> 10)
			lowSurrogate := 0xDC00 + ((codePoint - 0x10000) & 0x3FF)
			encoded = append(encoded, byte(highSurrogate>>8), byte(highSurrogate), byte(lowSurrogate>>8), byte(lowSurrogate))
		}
	}
	return hex.EncodeToString(encoded)
}
