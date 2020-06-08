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

	log "github.com/Sirupsen/logrus"

	"fmt"

	cmodel "github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/alarm/api"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/alarm/redi"
	"github.com/open-falcon/falcon-plus/modules/api/app/model/uic"
)

func consume(event *cmodel.Event, isHigh bool) {
	actionId := event.ActionId()
	if actionId <= 0 {
		log.Errorf("[consume]consume_actionId_let_0 event_id:%s,event:%+v", event.Id, event)
		return
	}
	/*这里通过 event中的actionid 拿到 action
	  就是拿到这个 报警组的名字 是否有回调等信息
	*/
	action := api.GetAction(actionId)
	if action == nil {
		log.Errorf("[consume]consume_GetAction_failed_event event_id:%s,event:%+v", event.Id, event)
		return
	}
	// 通过api 获取 endpoint + tpl_id 对应的主机组列表
	groups := api.GetEndpointTplGroups(event.TplId(), event.Endpoint)
	// 通过api 获取 endpoint + tpl_id 对应的tag列表
	amstag := api.GetAmsTag(event.Endpoint)
	log.Infof("[consume]consume_action_successfully  event_id:%s,event:%+v action:%+v", event.Id, event, action)
	//有回调的话处理下http get调用相应的回调函数,会把报警的信息作为参数带上
	if action.Callback == 1 {
		HandleCallback(event, action, amstag, groups)
	}
	if isHigh {
		consumeHighEvents(event, action, amstag, groups)
	} else {
		consumeLowEvents(event, action, amstag, groups)
	}
}

// 高优先级的不做报警合并
func consumeHighEvents(event *cmodel.Event, action *api.Action, amstag string, grps string) {
	//如果报警没有接收组,那么直接返回了
	if action.Uic == "" && action.LarkGroupId == "" {
		log.Errorf("[consumeHighEvents]action_uic_or_larkGid_null  event_id:%s,event:%+v,action:%+v", event.Id, event, action)
		return
	}
	// 给lark 群组报警用的
	//userNames, mails := api.ParseTeamsForLarkGroup(action.Uic, action.LarkGroupId)

	userMap := GenerateUserMap(event, action)

	var filterNames []string
	var filterMails []string
	for k, _ := range userMap {
		filterNames = append(filterNames, k)
		filterMails = append(filterMails, fmt.Sprintf("%s%s", k, g.BYTEMAIL))
	}

	//生成报警内容,这里可以为不同通道的报警做定制
	smsContent := GenerateSmsContent(event, amstag, grps)
	mailContent := GenerateMailContent(event, amstag, grps)
	//imContent := GenerateIMContent(event)
	imCardContentMap := BuildLarkCardImContent(filterNames, event, amstag, grps)

	phoneContent := GeneratePhoneContent(event)

	/* 这里根据报警的级别可以做通道的定制
	如<=P2 才发送短信 =p9 电话报警等等
	下面的redi.wirtesms等方法就是将报警内容lpush到不通通道的发送队列中
	*/

	if event.Priority() < 3 {
		redi.WriteSms(filterNames, smsContent)
	}
	//p9 电话报警
	if event.Priority() == 9 {
		redi.WriteSms(filterNames, smsContent)
		// 去掉P9 OK状态的报警
		if event.Status == "PROBLEM" {
			redi.WritePhone(filterNames, phoneContent)
		}
	}
	if imCardContentMap != nil {
		redi.WriteImCard(imCardContentMap)
	}

	redi.WriteMail(filterMails, smsContent, mailContent)
	log.Infof("[consumeHighEvents]successfully_generate_content event_id:%s,event:%+v,uic:%s,lark_gid:%s", event.Id, event, action.Uic, action.LarkGroupId)

}

// 低优先级的做报警合并
func consumeLowEvents(event *cmodel.Event, action *api.Action, amstag string, grps string) {
	if action.Uic == "" && action.LarkGroupId == "" {
		log.Errorf("[consumeLowEvents]action_uic_or_larkGid_null  event_id:%s,event:%+v,action:%+v", event.Id, event, action)
		return
	}

	// <=P2 才发送短信
	if event.Priority() < 3 {
		ParseUserSms(event, action, amstag, grps)
	}

	ParseUserIm(event, action, amstag, grps)
	ParseUserMail(event, action, amstag, grps)
	log.Infof("[consumeLowEvents]successfully_generate_content event_id:%s,event:%+v,uic:%s,lark_gid:%s", event.Id, event, action.Uic, action.LarkGroupId)

}

func LowLevelFilterBlock(event *cmodel.Event, userMap map[string]*uic.User) map[string]*uic.User {
	NewMap := make(map[string]*uic.User)

	for userName, user := range userMap {
		counter := fmt.Sprintf("%s_%s", event.Endpoint, event.Metric())
		blockKey := fmt.Sprintf("%s%s_%s", g.BLOCK_MONITOR_KEY_PREFIX, userName, counter)
		if _, exist := BlockMonitorCounter.Load(blockKey); exist == false {
			NewMap[userName] = user
		}
	}
	return NewMap
}

func ParseUserSms(event *cmodel.Event, action *api.Action, amstag string, grps string) {
	userMap := GenerateUserMap(event, action)
	content := GenerateSmsContent(event, amstag, grps)
	metric := event.Metric()
	status := event.Status
	priority := event.Priority()

	queue := g.Config().Redis.UserSmsQueue

	//rc := g.RedisConnPool.Get()
	rc := redi.RedisCluster
	//defer rc.Close()

	for _, user := range userMap {
		dto := SmsDto{
			Priority: priority,
			Metric:   metric,
			Content:  content,
			Phone:    user.Name,
			Status:   status,
		}
		bs, err := json.Marshal(dto)
		if err != nil {
			log.Error("ParseUserSms_json_marshal_SmsDto_fail_error:%+v,SmsDto:%+v", err, dto)
			continue
		}

		_, err = rc.Do("LPUSH", queue, string(bs))
		if err != nil {
			log.Error("ParseUserSms_LPUSH_redis", queue, "fail:", err, "dto:", string(bs))
		}
	}
}

func ParseUserMail(event *cmodel.Event, action *api.Action, amstag string, grps string) {
	//api根据报警组获取组里人
	userMap := GenerateUserMap(event, action)
	metric := event.Metric()
	subject := GenerateSmsContent(event, amstag, grps)
	content := GenerateMailContent(event, amstag, grps)
	status := event.Status
	priority := event.Priority()

	queue := g.Config().Redis.UserMailQueue

	//rc := g.RedisConnPool.Get()
	rc := redi.RedisCluster
	//defer rc.Close()
	//遍历usermap 生成报警 中间态消息并写入中间队列
	for _, user := range userMap {
		dto := MailDto{
			Priority: priority,
			Metric:   metric,
			Subject:  subject,
			Content:  content,
			Email:    user.Email,
			Status:   status,
		}
		bs, err := json.Marshal(dto)
		if err != nil {
			log.Error("ParseUserMail_json_marshal_MailDto_fail_error:%+v,MailDto:%+v", err, dto)
			continue
		}

		_, err = rc.Do("LPUSH", queue, string(bs))
		if err != nil {
			log.Error("ParseUserMail_LPUSH_redis", queue, "fail:", err, "dto:", string(bs))
		}
	}
}

func GenerateUserMap(event *cmodel.Event, action *api.Action) map[string]*uic.User {

	userMap := api.GetUsers(action.Uic)
	if action.LarkGroupId != "" {
		grpUser := &uic.User{Name: action.LarkGroupId, Phone: action.LarkGroupId}
		userMap[action.LarkGroupId] = grpUser
	}
	userMap = CommonFilterBlock(event, userMap)
	return userMap

}
func ParseUserIm(event *cmodel.Event, action *api.Action, amstag string, grps string) {
	/*
		高优先级的报警需要将uic解析成 user列表
		低有先级报警需要合并所以要单个解析user
	*/
	userMap := GenerateUserMap(event, action)
	commonContent := GenerateMailContent(event, amstag, grps)
	metric := event.Metric()
	status := event.Status
	priority := event.Priority()

	queue := g.Config().Redis.UserIMQueue

	//rc := g.RedisConnPool.Get()
	rc := redi.RedisCluster
	//defer rc.Close()
	for userName, user := range userMap {
		userNames := []string{userName}
		thisUserImContent := BuildLarkCardImContent(userNames, event, amstag, grps)[userName]
		dto := ImDto{
			Priority:        priority,
			Metric:          metric,
			Content:         commonContent,
			IM:              user.Name,
			Name:            user.Name,
			Status:          status,
			LarkCardContent: thisUserImContent,
		}
		bs, err := json.Marshal(dto)
		if err != nil {
			log.Error("ParseUserIm_json_marshal_ImDto_fail_error:%+v,ImDto:%+v", err, dto)
			continue
		}

		_, err = rc.Do("LPUSH", queue, string(bs))
		if err != nil {
			log.Error("ParseUserIm_LPUSH_redis", queue, "fail:", err, "dto:", string(bs))
		}
	}
}
