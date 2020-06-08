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
	"fmt"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/common/utils"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"

	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/alarm/api"
)

//func BuildCommonSMSContent(event *model.Event) string {
func BuildCommonSMSContent(event *model.Event, uic string, amstag string, grps string) string {
	return fmt.Sprintf(
		//"[P%d][%s][%s][][%s %s %s %s %s%s%s][O%d %s]",
		"[%s][P%d][Id:%s][状态%s][Endpoint:%s][报警分组:%s][接收组:%s][组tag列表:%s][metric:%s_tag:%s 注意:%s  表达式:%s ][第%d/%d次 ,时间%s]",
		AlarmIsUnionString(event),
		event.Priority(),
		event.Id,
		event.Status,
		event.Endpoint,
		grps,
		uic,
		amstag,
		event.Metric(),
		FormatMetricTags(event),
		event.Note(),
		UnionAlarmFormatFuncValue(event),
		event.CurrentStep,
		event.MaxStep(),
		event.FormattedTime(),
	)
}

func BuildCommonPhoneContent(event *model.Event) string {
	return fmt.Sprintf(
		"%s%s%s",
		AlarmIsUnionString(event),
		event.Endpoint,
		event.Metric(),
	)
}

func FormatOneFuncValue(event *model.Event) string {
	// all(#1): 90.77289!=21313
	return fmt.Sprintf("%s: %s%s%s", event.Func(), utils.ReadableFloat(event.LeftValue), event.Operator(), utils.ReadableFloat(event.RightValue()))

}

func UnionAlarmFormatFuncValue(event *model.Event) string {
	// all(#1): 90.77289!=21313 ,
	one := FormatOneFuncValue(event)

	for _, e := range event.EventChain {
		one = one + "," + FormatOneFuncValue(e)
	}
	return one
}

func FormatMetricTags(event *model.Event) string {
	one := utils.SortedTags(event.PushedTags)
	for _, e := range event.EventChain {
		one = one + "," + utils.SortedTags(e.PushedTags)
	}
	return one
}

func AlarmIsUnionString(event *model.Event) string {
	unionAlarm := "单条策略报警"
	if event.Strategy.UnionStrategyId > 0 {
		//log.Printf("收到组合报警:%+v", event)
		nums := len(event.EventChain) + 1
		unionAlarm = fmt.Sprintf("[%d]条组合策略报警", nums)
	}
	return unionAlarm
}

//func BuildLarkCardImContent(mails []string, event *model.Event, amstag string, grps string) map[string]string {
//	CardMap := make(map[string]string)
//	actionId := event.ActionId()
//	if actionId <= 0 {
//		log.Println("action id error")
//	}
//	action := api.GetAction(actionId)
//	if action == nil {
//		return nil
//	}
//	uic := action.Uic
//
//	link := g.Link(event)
//	modLink := g.ModLink(event)
//	modUicLink := g.ModUicLink(uic)
//	blockUrl := fmt.Sprintf("%s/blockmonitor", g.Config().AlarmApi)
//	unBlockUrl := fmt.Sprintf("%s/unblockmonitor", g.Config().AlarmApi)
//	counter := fmt.Sprintf("%s_%s", event.Endpoint, event.Metric())
//	unionAlarm := AlarmIsUnionString(event)
//
//	for _, userMail := range mails {
//		userName := strings.Split(userMail, "@")[0]
//		cardMsg := fmt.Sprintf("<card><title style='color:blue'><i18n local='zh_CN'>falcon-lark报警</i18n></title><p><text>报警触发类型:%s</text></p><p><text>状态:%s</text></p><p><text>级别:%d</text></p><p><text>Endpoint:%s</text></p><p><text>主机组列表:%s</text></p><p><text>接收组:%s</text><a href='%s'> 修改接收人</a></p><p><text>tag列表:%s</text></p><p><text>Metric:%s</text></p><p><text>Tags:%s</text></p><p><text>表达式:%s</text></p><p><text>Note:%s</text></p><p><text>最大报警数:%d, 当前第几次:%d</text></p><p><text>%s</text></p><p><a href='%s'>查看模板</a><a href='%s'> 修改模板</a></p><action changeable='true'><button><text><i18n local='zh_CN'>屏蔽1小时</i18n></text><request method='post' url='%s' parameter='{\"counter\":\"%s\",\"time\":3600,\"user_name\":\"%s\"}' need_user_info='true' need_message_info='true'></request></button><button><text><i18n local='zh_CN'>屏蔽6小时</i18n></text><request method='post' url='%s' parameter='{\"counter\":\"%s\",\"time\":21600,\"user_name\":\"%s\"}' need_user_info='true' need_message_info='true'></request></button><button><text><i18n local='zh_CN'>屏蔽24小时</i18n></text><request method='post' url='%s' parameter='{\"counter\":\"%s\",\"time\":86400,\"user_name\":\"%s\"}' need_user_info='true' need_message_info='true'></request></button><button><text><i18n local='zh_CN'>取消屏蔽</i18n></text><request method='post' url='%s' parameter='{\"counter\":\"%s\",\"user_name\":\"%s\"}' need_user_info='true' need_message_info='true'></request></button></action></card>",
//			unionAlarm,
//			event.Status,
//			event.Priority(),
//			event.Endpoint,
//			grps,
//			uic,
//			modUicLink,
//			amstag,
//			event.Metric(),
//			FormatMetricTags(event),
//			UnionAlarmFormatFuncValue(event),
//			event.Note(),
//			event.MaxStep(),
//			event.CurrentStep,
//			event.FormattedTime(),
//			link,
//			modLink,
//			blockUrl,
//			counter,
//			userName,
//			blockUrl,
//			counter,
//			userName,
//			blockUrl,
//			counter,
//			userName,
//			unBlockUrl,
//			counter,
//			userName)
//		CardMap[userMail] = cardMsg
//
//	}
//
//	return CardMap
//}

func LarkCardButton(name, url string) map[string]interface{} {
	action := make(map[string]interface{})
	action["tag"] = "button"
	Ctext := make(map[string]interface{})
	Ctext["tag"] = "lark_md"
	Ctext["content"] = name
	action["text"] = Ctext
	action["url"] = url
	action["type"] = "primary"
	return action
}

func LarkCardCallBackButton(name string, data interface{}) map[string]interface{} {
	action := make(map[string]interface{})
	action["tag"] = "button"
	Ctext := make(map[string]interface{})
	Ctext["tag"] = "lark_md"
	Ctext["content"] = name
	action["text"] = Ctext
	action["value"] = data
	action["type"] = "primary"
	return action
}

func BuildLarkCardImContent(userNames []string, event *model.Event, amstag string, grps string) map[string]string {
	CardMap := make(map[string]string)
	actionId := event.ActionId()
	if actionId <= 0 {
		log.Errorf("[BuildLarkCardImContent]actionId_lt_0 event_id:%s,event:%+v,actionId:%+v", event.Id, event, actionId)
	}
	action := api.GetAction(actionId)
	if action == nil {
		log.Errorf("[BuildLarkCardImContent]actionId_null event_id:%s,event:%+v,actionId:%+v", event.Id, event, actionId)

		return nil
	}
	uic := action.Uic

	link := g.Link(event)
	modLink := g.ModLink(event)
	modUicLink := g.ModUicLink(uic)

	ec := fmt.Sprintf("%s_%s", event.Endpoint, event.Metric())
	mc := event.Metric()
	unionAlarm := AlarmIsUnionString(event)

	for _, userName := range userNames {
		isGroup := g.JudgeIsLongInt(userName)

		data := make(map[string]interface{})
		cardData := make(map[string]interface{})

		if isGroup {
			data["chat_id"] = userName
		} else {
			data["email"] = userName + g.BYTEMAIL
		}
		data["msg_type"] = "interactive"

		config := make(map[string]bool)
		config["wide_screen_mode"] = true
		cardData["config"] = config

		header := make(map[string]interface{})
		title := make(map[string]interface{})
		title["tag"] = "plain_text"
		title["content"] = "falcon-lark报警"
		header["title"] = title
		header["template"] = "red"

		cardData["header"] = header

		// element
		element := make([]interface{}, 0)
		elementText := make(map[string]interface{})
		elementText["tag"] = "div"
		elementTextInner := make(map[string]interface{})

		elementTextInner["tag"] = "plain_text"

		contentText := fmt.Sprintf(
			"报警触发类型:%s\n"+
				"状态:%s\n"+
				"级别:%d\n"+
				"Id:%s\n"+
				"Endpoint:%s\n"+
				"主机组列表:%s\n"+
				"接收组:%s\n"+
				"tag列表:%s\n"+
				"Metric:%s\n"+
				"Tags:%s\n"+
				"表达式:%s\n"+
				"Note:%s\n"+
				"最大报警数:%d\n"+
				"当前第几次:%d\n"+
				"触发时间:%s",
			unionAlarm,
			event.Status,
			event.Priority(),
			event.Id,
			event.Endpoint,
			grps,
			uic,
			amstag,
			event.Metric(),
			FormatMetricTags(event),
			UnionAlarmFormatFuncValue(event),
			event.Note(),
			event.MaxStep(),
			event.CurrentStep,
			event.FormattedTime(),
		)
		elementTextInner["content"] = contentText
		elementText["text"] = elementTextInner

		element = append(element, elementText)
		// ele hr
		elementHr := make(map[string]interface{})
		elementHr["tag"] = "hr"
		element = append(element, elementHr)

		// ele action
		elementAction1 := make(map[string]interface{})
		elementAction1["tag"] = "action"
		elementActionInner1 := make([]interface{}, 0)

		elementAction2 := make(map[string]interface{})
		elementAction2["tag"] = "action"
		elementActionInner2 := make([]interface{}, 0)

		bottonData1 := make(map[string]interface{})
		bottonData2 := make(map[string]interface{})
		bottonData1[`修改接收人`] = modUicLink
		bottonData1[`查看模板`] = link
		bottonData1[`修改模板`] = modLink

		blockEcData := make(map[string]interface{})
		unblockEcData := make(map[string]interface{})
		blockMcData := make(map[string]interface{})
		unblockMcData := make(map[string]interface{})

		blockEcData[`user_name`] = userName
		blockEcData[`counter`] = ec
		blockEcData[`time`] = 3 * 3600
		blockEcData[`type`] = `block`

		unblockEcData[`user_name`] = userName
		unblockEcData[`counter`] = ec
		unblockEcData[`time`] = 3 * 3600
		unblockEcData[`type`] = `unblock`

		blockMcData[`user_name`] = userName
		blockMcData[`counter`] = mc
		blockMcData[`time`] = 3 * 3600
		blockMcData[`type`] = `block`

		unblockMcData[`user_name`] = userName
		unblockMcData[`counter`] = mc
		unblockMcData[`time`] = 3 * 3600
		unblockMcData[`type`] = `unblock`

		bottonData2[`屏蔽endpoint+metric 3小时`] = blockEcData
		bottonData2[`屏蔽metric 3小时`] = blockMcData

		bottonData2[`取消end+metric屏蔽`] = unblockEcData
		bottonData2[`取消metric屏蔽`] = unblockMcData

		for k, v := range bottonData1 {
			nv := v.(string)
			t := LarkCardButton(k, nv)
			elementActionInner1 = append(elementActionInner1, t)
		}
		elementAction1["actions"] = elementActionInner1
		element = append(element, elementAction1)
		element = append(element, elementHr)

		for k, v := range bottonData2 {
			//nv := v.(string)
			t := LarkCardCallBackButton(k, v)
			elementActionInner2 = append(elementActionInner2, t)
		}

		elementAction2["actions"] = elementActionInner2
		element = append(element, elementAction2)
		element = append(element, elementHr)

		cardData["elements"] = element

		data["card"] = cardData

		bytesData, _ := json.Marshal(data)

		fContent := string(bytesData)

		CardMap[userName] = fContent

	}

	return CardMap
}

func BuildCommonIMContent(event *model.Event) string {
	return fmt.Sprintf(
		"[P%d][%s][%s][][%s %s %s %s %s%s%s][O%d %s]",
		event.Priority(),
		event.Status,
		event.Endpoint,
		event.Note(),
		event.Func(),
		event.Metric(),
		utils.SortedTags(event.PushedTags),
		utils.ReadableFloat(event.LeftValue),
		event.Operator(),
		utils.ReadableFloat(event.RightValue()),
		event.CurrentStep,
		event.FormattedTime(),
	)
}

func BuildCommonMailContent(event *model.Event, uic string, amstag string, grps string) string {
	link := g.Link(event)
	return fmt.Sprintf(
		//"%s\r\nP%d\r\nEndpoint:%s\r\nMetric:%s\r\nTags:%s\r\n%s: %s%s%s\r\nNote:%s\r\nMax:%d, Current:%d\r\nTimestamp:%s\r\n%s\r\n",
		"%s\r%s\r\nP%d\r\nId:%s\r\nEndpoint:%s\r\n主机组列表:%s\r\nUic:%s\r\nAmstag:%s\r\nMetric:%s\r\nTags:%s\r\n%s\r\nNote:%s\r\nMax:%d, Current:%d\r\nTimestamp:%s\r\n%s\r\n<br>",
		AlarmIsUnionString(event),
		event.Status,
		event.Priority(),
		event.Id,
		event.Endpoint,
		grps,
		uic,
		amstag,
		event.Metric(),
		FormatMetricTags(event),
		UnionAlarmFormatFuncValue(event),
		event.Note(),
		event.MaxStep(),
		event.CurrentStep,
		event.FormattedTime(),
		link,
	)
}

//func GenerateSmsContent(event *model.Event) string {
//	return BuildCommonSMSContent(event)
//}
//
//func GenerateMailContent(event *model.Event) string {
//	return BuildCommonMailContent(event)
//}

func GenerateSmsContent(event *model.Event, amstag string, grps string) string {
	actionId := event.ActionId()
	if actionId <= 0 {
		log.Errorf("[GenerateSmsContent]consume_actionId_let_0 event_id:%s,event:%+v", event.Id, event)
	}
	action := api.GetAction(actionId)
	if action == nil {
		return ""
	}
	uic := action.Uic
	return BuildCommonSMSContent(event, uic, amstag, grps)
}

func GenerateMailContent(event *model.Event, amstag string, grps string) string {

	actionId := event.ActionId()
	if actionId <= 0 {
		log.Errorf("[GenerateMailContent]consume_actionId_let_0 event_id:%s,event:%+v", event.Id, event)
	}
	action := api.GetAction(actionId)
	if action == nil {
		return ""
	}
	uic := action.Uic
	return BuildCommonMailContent(event, uic, amstag, grps)
}

func GenerateIMContent(event *model.Event) string {
	return BuildCommonIMContent(event)
}

func GeneratePhoneContent(event *model.Event) string {
	return BuildCommonPhoneContent(event)
}
