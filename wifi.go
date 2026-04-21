package main

import "strconv"

//0 or 1
func switchWiFiModule(SwitchOption int) {
	msg, err := getCMD("RD", "wa_inner_version", "cr_version")
	if err != nil {
		print(err)
		return
	}
	LD, ok := msg["RD"].(string)
	if !ok {
		return
	}
	wa_inner_version, ok := msg["wa_inner_version"].(string)
	if !ok {
		return
	}
	cr_version, ok := msg["cr_version"].(string)
	if !ok {
		return
	}
	AD := SHA256(SHA256(wa_inner_version+cr_version) + LD) //Wa@9w+YWRtaW4=
	data := "goformId=switchWiFiModule&isTest=false&SwitchOption=" + strconv.Itoa(SwitchOption) + "&AD=" + AD

	body, err := setCMD(data)
	if err != nil {
		print(err)
		return
	}
	if string(body) != `{"result":0}` {
		if SwitchOption == 0 {
			print("关闭WiFi失败", string(body))
		} else if SwitchOption == 1 {
			print("开启WiFi失败", string(body))
		}
		return
	}
}
