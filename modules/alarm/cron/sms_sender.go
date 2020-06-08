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
	"time"

	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/alarm/model"
	"github.com/open-falcon/falcon-plus/modules/alarm/redi"
	"github.com/toolkits/net/httplib"
)

func ConsumeSms() {
	for {
		L := redi.PopAllSms()
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendSmsList(L)
	}
}

func SendSmsList(L []*model.Sms) {
	for _, sms := range L {
		SmsWorkerChan <- 1
		go SendSms(sms)
	}
}

func SendSms(sms *model.Sms) {
	defer func() {
		<-SmsWorkerChan
	}()
	retryS := 3

	for now := 0; now < retryS; now++ {
		err := HttpSendSms(sms)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

}

func HttpSendSms(sms *model.Sms) error {
	url := g.Config().Api.Sms
	r := httplib.Post(url).SetTimeout(5*time.Second, 30*time.Second)
	if sms.Tos == "" {
		return errors.New("empty_sms_tos")
	}

	//短信接口支持 多个用户名逗号分隔
	r.Param("username", sms.Tos)
	r.Param("msg", sms.Content)
	r.Param("sign", "6be72e1e6527834d72cc7a05370e3849")
	r.Param("app_id", "205768798499864")

	type resp_info struct {
		Response interface{} `json:"response"`
		Ret      int         `json:"ret"`
	}
	var Resp_info *resp_info
	resp_byte, _ := r.Bytes()
	err := json.Unmarshal(resp_byte, &Resp_info)
	if err != nil {
		log.Errorf("[HttpSendSms]error decoding sakura response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Errorf("[HttpSendSms]syntax error at byte offset %d", e.Offset)
		}
		log.Errorf("[HttpSendSms]sakura response: %q", resp_byte)
		return errors.New("json.Unmarshal_sakura response")
	}

	if Resp_info.Ret != 0 {
		log.Errorf("[HttpSendSms]send_sms_fail, tos:%s, cotent:%s, error:%v, resp.res:%v ,resp.ret:%v", sms.Tos, sms.Content, err, Resp_info.Response, Resp_info.Ret)
		return errors.New("Resp_info.Ret_not_equal_0")
	} else {
		log.Infof("[HttpSendSms]send_sms_success, tos:%s, cotent:%s", sms.Tos, sms.Content)
		return nil
	}

}
