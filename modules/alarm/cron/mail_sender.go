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

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/alarm/model"
	"github.com/open-falcon/falcon-plus/modules/alarm/redi"
	"github.com/toolkits/net/httplib"
)

func ConsumeMail() {
	for {
		L := redi.PopAllMail()
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendMailList(L)
	}
}

func SendMailList(L []*model.Mail) {
	for _, mail := range L {
		MailWorkerChan <- 1
		go SendMail(mail)
	}
}

func SendMail(mail *model.Mail) {
	defer func() {
		<-MailWorkerChan
	}()
	if mail.Tos == "" {
		return
	}
	url := g.Config().Api.Mail
	r := httplib.Post(url).SetTimeout(5*time.Second, 30*time.Second)
	r.Param("recipient", mail.Tos)
	r.Param("subject", mail.Subject)
	r.Param("message", mail.Content)
	r.Param("app_id", "205768798499864")
	r.Param("sign", "6be72e1e6527834d72cc7a05370e3849")
	type resp_info struct {
		Response string `json:"response"`
		Ret      int    `json:"ret"`
	}
	var Resp_info *resp_info
	resp_byte, _ := r.Bytes()
	err := json.Unmarshal(resp_byte, &Resp_info)
	if err != nil {
		log.Printf("error decoding sakura response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Errorf("[SendMail]syntax error at byte offset %d", e.Offset)
		}
		log.Errorf("[SendMail]sakura response: %q", resp_byte)
		return
	}

	if Resp_info.Ret != 0 {
		log.Errorf("[SendMail]send_mail_fail, recipient:%s, subject:%s, message:%s, error:%v ,resp:%v", mail.Tos, mail.Subject, mail.Content, err, Resp_info.Response)
	} else {
		log.Infof("[SendMail]send_mail_success, recipient:%s, subject:%s, message:%s", mail.Tos, mail.Subject, mail.Content)
	}

}
