// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cron

import (
	"encoding/json"
	"strings"
	"time"

	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/alarm/model"
	"github.com/open-falcon/falcon-plus/modules/alarm/redi"
	"github.com/toolkits/net/httplib"
)

func ConsumePhone() {
	for {
		L := redi.PopAllPhone()
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendPhoneList(L)
	}
}

func SendPhoneList(L []*model.Sms) {
	for _, phone := range L {
		PhoneWorkerChan <- 1
		go SendPhone(phone)
	}
}

func SendPhone(sms *model.Sms) {
	defer func() {
		<-PhoneWorkerChan
	}()
	log.Infof("SendPhone:%+v", sms.Tos)
	for _, user := range strings.Split(sms.Tos, ",") {

		retryS := 3
		for now := 0; now < retryS; now++ {
			err := HttpSendPhone(user, sms.Content)
			if err == nil {
				break
			}
			time.Sleep(1 * time.Second)
		}

		//if success == false {
		//	log.Errorf("[SendPhone]send_phone_fail, tos:%s, cotent:%s, error:%v ", user, sms.Content)
		//}
	}
}

func HttpSendPhone(user, content string) error {
	url := g.Config().Api.Phone
	r := httplib.Post(url).SetTimeout(5*time.Second, 30*time.Second)
	r.Param("username", user)
	r.Param("msg", content)
	r.Param("sign", "e21be0048c73646a850ae29cf1aa69cd")
	r.Param("app_id", "99403998548533")

	type resp_info struct {
		Response string `json:"response"`
		Ret      int    `json:"ret"`
	}

	var Resp_info *resp_info

	resp_byte, _ := r.Bytes()
	err := json.Unmarshal(resp_byte, &Resp_info)
	if err != nil {
		log.Errorf("[HttpSendPhone]error decoding sakura response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Errorf("[HttpSendPhone]syntax error at byte offset %d", e.Offset)
		}
		log.Errorf("[HttpSendPhone]sakura response: %q", resp_byte)
		return err
	}

	//resp, err := r.String()
	if Resp_info.Ret != 0 {
		log.Errorf("[HttpSendPhone]send_phone_fail, tos:%s, cotent:%s, error:%v ,resp:%v", user, content, err, Resp_info.Response)
		return errors.New("HttpSendPhone_Resp_info.Ret_not_zero")
	} else {
		log.Infof("[HttpSendPhone]send_phone_success, tos:%s, cotent:%s", user, content)
		return nil
	}

}
