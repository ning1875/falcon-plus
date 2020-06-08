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

package http

import (
	"net/http"

	"fmt"

	"time"

	"errors"

	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/open-falcon/falcon-plus/modules/alarm/cron"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/alarm/redi"
	"github.com/toolkits/file"
)

const badstatus = http.StatusBadRequest

type SingleBlockMonitorPer struct {
	CustomerParameter SingleBlockMonitor `json:"customer_parameter"`
}

type SingleBlockMonitor struct {
	Counter  string `json:"counter" form:"counter"`
	UserName string `json:"user_name" form:"user_name"`
	Time     int    `json:"time" form:"time"`
}

func Version(c *gin.Context) {
	c.String(200, g.VERSION)
}

func Health(c *gin.Context) {
	c.String(200, "ok")
}

func Workdir(c *gin.Context) {
	c.String(200, file.SelfDir())
}

func GetBlockMonitor(c *gin.Context) {
	res := make(map[string]string)
	f := func(k, v interface{}) bool {
		res[k.(string)] = v.(string)
		return true
	}
	cron.BlockMonitorCounter.Range(f)
	JSONR(c, 200, res)
}

func CreateBlockMonitorGet(c *gin.Context) {

	user := c.Query("user_name")
	counter := c.Query("counter")
	btime := c.Query("time")

	if user == "" || counter == "" {
		JSONR(c, badstatus, errors.New("wrong inputs"))
		return
	}

	ntime, _ := strconv.ParseInt(btime, 10, 64)
	//time = strconv.ParseInt()
	log.Infof("CreateBlockMonitor:%+v", user, counter, btime)
	blockKey := fmt.Sprintf("%s%s_%s", g.BLOCK_MONITOR_KEY_PREFIX, user, counter)
	rc := redi.RedisCluster
	if _, err := rc.Do("SETEX", blockKey, ntime, ntime); err != nil {
		log.Errorf("CreateBlockMonitor_setex_%s_error:%+v", blockKey, err)
		JSONR(c, badstatus, err)
	}
	if _, err := rc.Do("SADD", g.BLOCK_MONITOR_SET, blockKey); err != nil {
		log.Errorf("CreateBlockMonitor_sadd_%s_error:%+v", blockKey, err)
	}
	cron.BlockMonitorCounter.Store(blockKey, string(ntime))
	//发送成功屏蔽的信息
	ts := time.Now().Unix()
	msgBlockSuccess := fmt.Sprintf("[ACK通知]\r\n%s 已ACK报警规则: Counter:%s   %d 分钟\r\n本消息发送时间：%s",
		user,
		counter,
		ntime/60,
		time.Unix(ts, 0).Format("2006-01-02 15:04:05"))
	for _, token := range g.Config().LarkBotTokens {
		realToken := cron.GetTokenFromCache(token)
		if err := cron.LarkBotText(g.Config().Api.IM, msgBlockSuccess, user, realToken); err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}
	JSONR(c, 200, "block success")
}

func DeleteBlockMonitorGet(c *gin.Context) {

	user := c.Query("user_name")
	counter := c.Query("counter")

	if user == "" || counter == "" {
		JSONR(c, badstatus, errors.New("wrong inputs"))
		return
	}

	log.Infof("DeleteBlockMonitor:", user, counter)
	blockKey := fmt.Sprintf("%s%s_%s", g.BLOCK_MONITOR_KEY_PREFIX, user, counter)
	cron.BlockMonitorCounter.Delete(blockKey)
	rc := redi.RedisCluster
	if _, err := rc.Do("DEL", blockKey); err != nil {
		log.Errorf("DeleteBlockMonitor_DEL_%s_error:%+v", blockKey, err)
		JSONR(c, badstatus, err)
	}
	if _, err := rc.Do("SREM", g.BLOCK_MONITOR_SET, blockKey); err != nil {
		log.Errorf("DeleteBlockMonitor_SREM__%s_error:%+v", blockKey, err)
		JSONR(c, badstatus, err)
	}
	//发送成功屏蔽的信息
	ts := time.Now().Unix()
	msgBlockSuccess := fmt.Sprintf("[ACK通知]\r\n%s 已解除屏蔽报警规则: Counter:%s   \r\n本消息发送时间：%s",
		user, counter,
		time.Unix(ts, 0).Format("2006-01-02 15:04:05"))
	for _, token := range g.Config().LarkBotTokens {
		realToken := cron.GetTokenFromCache(token)
		if err := cron.LarkBotText(g.Config().Api.IM, msgBlockSuccess, user, realToken); err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}
	JSONR(c, 200, "delete block success")

}

func CreateBlockMonitor(c *gin.Context) {
	var inputs SingleBlockMonitor
	var err error
	if err = c.Bind(&inputs); err != nil {
		JSONR(c, badstatus, err)
		return
	}
	if inputs.UserName == "" {
		JSONR(c, badstatus, errors.New("wrong inputs"))
		return
	}
	log.Infof("CreateBlockMonitor:%+v", inputs)
	blockKey := fmt.Sprintf("%s%s_%s", g.BLOCK_MONITOR_KEY_PREFIX, inputs.UserName, inputs.Counter)
	rc := redi.RedisCluster
	if _, err := rc.Do("SETEX", blockKey, inputs.Time, inputs.Time); err != nil {
		log.Errorf("CreateBlockMonitor_setex_%s_error:%+v", blockKey, err)
		JSONR(c, badstatus, err)
	}
	if _, err := rc.Do("SADD", g.BLOCK_MONITOR_SET, blockKey); err != nil {
		log.Errorf("CreateBlockMonitor_sadd_%s_error:%+v", blockKey, err)
	}
	cron.BlockMonitorCounter.Store(blockKey, string(inputs.Time))
	//发送成功屏蔽的信息
	//ts := time.Now().Unix()
	//msgBlockSuccess := fmt.Sprintf("[ACK通知]\r\n%s 已ACK报警规则: Counter:%s   %d 分钟\r\n本消息发送时间：%s",
	//	inputs.CustomerParameter.UserName,
	//	inputs.CustomerParameter.Counter,
	//	inputs.CustomerParameter.Time/60,
	//	time.Unix(ts, 0).Format("2006-01-02 15:04:05"))
	//for _, token := range g.Config().LarkBotTokens {
	//
	//	if err = cron.LarkBotText(g.Config().Api.IM, msgBlockSuccess, fmt.Sprintf("%s%s", inputs.CustomerParameter.UserName, g.BYTEMAIL), token, g.JudgeIsLongInt(inputs.CustomerParameter.UserName)); err == nil {
	//		break
	//	}
	//	time.Sleep(time.Second * 1)
	//}
	JSONR(c, 200, "block success")
}

func DeleteBlockMonitor(c *gin.Context) {
	var inputs SingleBlockMonitor
	var err error
	if err = c.Bind(&inputs); err != nil {
		JSONR(c, badstatus, err)
		return
	}
	log.Infof("CreateBlockMonitor:%+v", inputs)
	blockKey := fmt.Sprintf("%s%s_%s", g.BLOCK_MONITOR_KEY_PREFIX, inputs.UserName, inputs.Counter)
	cron.BlockMonitorCounter.Delete(blockKey)
	rc := redi.RedisCluster
	if _, err := rc.Do("DEL", blockKey); err != nil {
		log.Errorf("DeleteBlockMonitor_DEL_%s_error:%+v", blockKey, err)
		JSONR(c, badstatus, err)
	}
	if _, err := rc.Do("SREM", g.BLOCK_MONITOR_SET, blockKey); err != nil {
		log.Errorf("DeleteBlockMonitor_SREM__%s_error:%+v", blockKey, err)
		JSONR(c, badstatus, err)
	}
	////发送成功屏蔽的信息
	//ts := time.Now().Unix()
	//msgBlockSuccess := fmt.Sprintf("[ACK通知]\r\n%s 已解除屏蔽报警规则: Counter:%s   \r\n本消息发送时间：%s",
	//	inputs.UserName,
	//	inputs.Counter,
	//	time.Unix(ts, 0).Format("2006-01-02 15:04:05"))
	//for _, token := range g.Config().LarkBotTokens {
	//	if err = cron.LarkBotText(g.Config().Api.IM, msgBlockSuccess, fmt.Sprintf("%s%s", inputs.CustomerParameter.UserName, g.BYTEMAIL), token, g.JudgeIsLongInt(inputs.CustomerParameter.UserName)); err == nil {
	//		break
	//	}
	//	time.Sleep(time.Second * 1)
	//}
	JSONR(c, 200, "delete block success")

}
