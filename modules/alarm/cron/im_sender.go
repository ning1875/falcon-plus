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
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/utils"
	"github.com/open-falcon/falcon-plus/modules/alarm/api"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/alarm/model"
	"github.com/open-falcon/falcon-plus/modules/alarm/redi"
)

type BotResp struct {
	Data interface{} `json:"data"`
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
}

func ConsumeIM() {
	for {
		//rpop出所有的报警信息到一个slice中
		L := redi.PopAllIM(redi.IM_QUEUE_NAME)
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendIMList(L)
	}
}

func ConsumeFailedIM() {
	for {
		time.Sleep(time.Minute * 10)
		//rpop出所有的报警信息到一个slice中
		L := redi.PopAllIM(redi.FAILED_IM_QUEUE_NAME)
		if len(L) == 0 {
			continue
		}
		SendFailedIMList(L)
	}
}

func SendIMList(L []*model.IM) {
	for _, im := range L {
		/*
			1.IMWorkerChan是带缓冲的chan,chan的长度意思就是同时可以多少个send作业
			2.向im发送workerchan中写入1说明可以发送一条
			3.如果队列没满,是不会阻塞在这里的,否则会阻塞
		*/
		IMWorkerChan <- 1
		go SendIM(im, redi.FAILED_IM_QUEUE_NAME, IMWorkerChan)
	}
}

func SendFailedIMList(L []*model.IM) {
	for _, im := range L {
		/*
			1.IMWorkerChan是带缓冲的chan,chan的长度意思就是同时可以多少个send作业
			2.向im发送workerchan中写入1说明可以发送一条
			3.如果队列没满,是不会阻塞在这里的,否则会阻塞
		*/
		FailedIMWorkerChan <- 1
		go SendIM(im, redi.FINALLY_FAILED_IM_QUEUE_NAME, FailedIMWorkerChan)
	}
}

func LarkBotText(url, Content, Tos, token string) error {

	data := make(map[string]interface{})
	content := make(map[string]string)
	content["text"] = Content
	isGroup := g.JudgeIsLongInt(Tos)
	if isGroup {
		data["chat_id"] = strings.Split(Tos, "@")[0]
	} else {
		data["email"] = Tos + g.BYTEMAIL
	}
	data["msg_type"] = "text"
	data["content"] = content
	//log.Debugf("LarkBot:token:%+v,Tos:%+v,data:%+v,", token, Tos, data)
	bytesData, err := json.Marshal(data)
	//log.Infof("[LarkBotText] %s", string(bytesData))
	if err != nil {
		log.Errorf("[LarkBotText]LarkBot_json_marshal_error:%+v,Tos:%+v,token:%+v,content:%+v,", err.Error(), Tos, token, Content)
		return err
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Errorf("[LarkBotText]LarkBot_http_post_1_error:%+v,Tos:%+v,token:%+v,content:%+v,", err.Error(), Tos, token, Content)
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("Authorization", "Bearer "+token)
	client := http.Client{}
	client.Timeout = time.Duration(time.Second * 2)
	resp, err := client.Do(request)
	if err != nil {
		log.Errorf("[LarkBotText]LarkBot_http_post_2_error:%+v,Tos:%+v,token:%+v,content:%+v,", err.Error(), Tos, token, Content)
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[LarkBotText]LarkBot_http_post_3_error:%+v,Tos:%+v,token:%+v,content:%+v,", err.Error(), Tos, token, Content)
		return err
	}
	var BotResp *BotResp
	err = json.Unmarshal(respBytes, &BotResp)
	if err != nil {
		log.Errorf("[LarkBotText]LarkBot_json_unmarshal_error:%+v,Tos:%+v,token:%+v,content:%+v,", err.Error(), Tos, token, Content)
		return err
	}

	if BotResp.Code != 0 {
		log.Errorf("[LarkBotText]req_lark_bot_error:%s,code:%d,Msg:%+v,Tos:%+v,token:%+v,content:%+v", BotResp.Data, BotResp.Code, BotResp.Msg, Tos, token, Content)
		return errors.New(fmt.Sprintf("req_lark_bot_error:%s,code:%d", BotResp.Data, BotResp.Code))
	} else {
		log.Infof("[LarkBotText]send_lark_success:  tos:%s, content:%s", Tos, Content)
		return nil
	}
}

func LarkBotCard(url, bodyData, Tos, token string) error {

	var dd []byte
	dd = []byte(bodyData)

	reader := bytes.NewReader(dd)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Errorf("[LarkBotCard]LarkBot_http_post_1_error:%+v,Tos:%+v,token:%+v,content:%+v,", err.Error(), Tos, token, bodyData)
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("Authorization", "Bearer "+token)
	client := http.Client{}
	client.Timeout = time.Duration(time.Second * 2)
	resp, err := client.Do(request)
	if err != nil {
		log.Errorf("[LarkBotCard]LarkBot_http_post_2_error:%+v,Tos:%+v,token:%+v,content:%+v,", err.Error(), Tos, token, bodyData)
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[LarkBotCard]LarkBot_http_post_3_error:%+v,Tos:%+v,token:%+v,content:%+v,", err.Error(), Tos, token, bodyData)
		return err
	}

	var BotResp *BotResp
	err = json.Unmarshal(respBytes, &BotResp)
	if err != nil {
		log.Errorf("[LarkBotCard]LarkBot_json_unmarshal_error:%+v,Tos:%+v,token:%+v,content:%+v,", err.Error(), Tos, token, bodyData)
		return err
	}

	if BotResp.Code != 0 {
		log.Errorf("[LarkBotCard]req_lark_bot_error:code:%d,Msg:%+v,Tos:%+v,token:%+v,content:%+v", BotResp.Code, BotResp.Msg, Tos, token, bodyData)
		return errors.New(fmt.Sprintf("req_lark_bot_error:%s,code:%d", BotResp.Code))
	} else {
		log.Infof("[LarkBotCard]send_lark_success:  tos:%s, content:%s", Tos, bodyData)
		return nil
	}
}

func GetTokenFromCache(token string) string {

	if res, found := api.OtherCache.Get(token); found {
		return res.(string)
	}
	app_id, app_secret := strings.Split(token, ",")[0], strings.Split(token, ",")[1]
	res := GetRebotTokenByIdAndSecret(app_id, app_secret, g.Config().Api.LarkTenantAccessTokenUrl)
	return res

}

func GetRebotTokenByIdAndSecret(app_id, app_secret, url string) string {
	type tokenRes struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		expire            int    `json:"expire"`
	}
	data := make(map[string]interface{})
	data["app_id"] = app_id
	data["app_secret"] = app_secret
	bytesData, err := json.Marshal(data)
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	// http短连接
	request.Close = true
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Error("[GetRebotTokenByIdAndSecret] http req error [2]", "url", url, "err", err.Error())
		return ""
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[GetRebotTokenByIdAndSecret] ioutil.ReadAll error :%+v", err)
		return ""
	}
	var BotResp *tokenRes
	err = json.Unmarshal(respBytes, &BotResp)
	if err != nil {
		log.Errorf("[GetRebotTokenByIdAndSecret] json.Unmarshal error :%+v", err)
		return ""
	}
	tokenKey := fmt.Sprintf("%s,%s", app_id, app_secret)

	api.OtherCache.Set(tokenKey, BotResp.TenantAccessToken, time.Duration(BotResp.expire))
	return BotResp.TenantAccessToken
}

func SendIM(im *model.IM, queue string, worker chan int) {
	/*
		1.这里使用defer的逻辑是先发送后读取chan
		2.因为如果先读取意味着又有一个work可以开始和逻辑相反
		3.下面就是自己定制的发送方式了
	*/
	defer func() {
		<-worker
	}()
	if im.Tos == "" {
		log.Errorf("SendIM_content_tos_empty_error %+v ", im)
		return
	}
	url := g.Config().Api.IM
	var larkTokens []string
	for i := 0; i < g.Config().AlarmRetryTimes; i++ {
		larkTokens = append(larkTokens, g.Config().LarkBotTokens...)
	}
	//token := "b-2a6a03a1-0bde-47de-8081-f80d32eee41d"
	//token_bk := "b-d35c5c0c-5e1b-4d9d-a78d-2fe42ff19a7e"

	for _, user := range strings.Split(im.Tos, ",") {
		thisSucc := false
		for _, token := range larkTokens {
			realToken := GetTokenFromCache(token)
			//var larkTo string
			//if len(strings.Split(user, "@")) <= 1 {
			//	larkTo = fmt.Sprintf("%s%s", user, g.BYTEMAIL)
			//} else {
			//	larkTo = user
			//}
			var sendErr error
			if im.IsCard {
				sendErr = LarkBotCard(url, im.Content, utils.CompressStr(user), realToken)
			} else {
				sendErr = LarkBotText(url, im.Content, utils.CompressStr(user), realToken)
			}
			if sendErr == nil {
				thisSucc = true
				//log.Infof("SendIM_success_from_token:%s,queue:%s_to_user:%s,content:%s", realToken, queue, user, im.Content)
				break
			}
			time.Sleep(1 * time.Second)

		}

		if thisSucc == false {
			//log.Errorf("SendIM_failed_to_user:%s, content:%s", user, im.Content)
			var newIm model.IM
			newIm.Tos = user
			newIm.Content = im.Content
			redi.WriteIMModel(&newIm, queue)
			log.Errorf("[SendIM]SendIM_failed_to_user_push_to_queue:%s,to_user:%s, content:%s", queue, user, im.Content)
		}

	}

}
