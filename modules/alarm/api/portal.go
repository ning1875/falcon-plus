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

package api

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"strings"

	"strconv"

	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sort"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/patrickmn/go-cache"
	"github.com/toolkits/net/httplib"
)

//TODO:use api/app/model/falcon_portal/action.go
type Action struct {
	Id                 int    `json:"id"`
	Uic                string `json:"uic"`
	LarkGroupId        string `json:"lark_group_id"`
	Url                string `json:"url"`
	Callback           int    `json:"callback"`
	BeforeCallbackSms  int    `json:"before_callback_sms"`
	BeforeCallbackMail int    `json:"before_callback_mail"`
	AfterCallbackSms   int    `json:"after_callback_sms"`
	AfterCallbackMail  int    `json:"after_callback_mail"`
}

type HostGroups struct {
	Name string `json:"grp_name" gorm:"column:grp_name"`
	Id   int64  `json:"grp_id" gorm:"column:grp_id"`
}

type ActionCache struct {
	sync.RWMutex
	M map[int]*Action
}

//var Actions = &ActionCache{M: make(map[int]*Action)}
var (
	ActionsCache     = cache.New(g.CacheTTl, g.CacheTTl)
	EndTplGroupCache = cache.New(g.CacheTTl, g.CacheTTl)
	EndTagsCache     = cache.New(g.CacheTTl, g.CacheTTl)
	OtherCache       = cache.New(g.CacheTTl, g.CacheTTl)
	//ActionsCache = cache.New(g.CacheTTl, g.CacheTTl)

)

//func (this *ActionCache) Get(id int) *Action {
//	this.RLock()
//	defer this.RUnlock()
//	val, exists := this.M[id]
//	if !exists {
//		return nil
//	}
//
//	return val
//}
//
//func (this *ActionCache) Set(id int, action *Action) {
//	this.Lock()
//	defer this.Unlock()
//	this.M[id] = action
//}

func GetAmsTag(endpoint string) string {
	if res, found := EndTagsCache.Get(endpoint); found {
		return res.(string)
	} else {
		res := HttpGetTag(endpoint)
		if res == "" {
			return ""
		}
		ActionsCache.Set(endpoint, res, cache.DefaultExpiration)
		return res
	}
}

func HttpGetTag(endpoint string) string {
	retryLimit := 3
	r_s := 0
	for r_s < retryLimit {
		res := getAmsTag(endpoint)
		if res != "" {
			return res
		}
		time.Sleep(1 * time.Second)
		r_s++
	}
	return ""
}

func getAmsTag(endpoint string) string {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	type Amstag struct {
		Response map[string][]string `json:"response"`
		Ret      int                 `json:"ret"`
	}

	ipaddr, err := net.LookupIP(fmt.Sprintf("%s.xxx.com", endpoint))
	if err != nil || len(ipaddr) < 1 {
		//log.Println(err)
		return ""
	}

	v := url.Values{}
	v.Set("app_id", "135916235682081")
	v.Set("sign", "ff9406de0e8008f3308d4834e06583cb")
	v.Set("ip", ipaddr[0].String())
	sendBody := bytes.NewReader([]byte(v.Encode()))

	send_url := "http://api-ops.xxx.com/ams/host/tag"
	c := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, _ := http.NewRequest("POST", send_url, sendBody)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return ""
	}
	amstag := Amstag{}
	err = json.Unmarshal(body, &amstag)
	if err != nil {
		log.Println(err)
		return ""
	}
	if amstag.Ret != 0 {
		log.Errorf("reqeust to ams tag api return error. %+v", amstag)
		return ""
	}
	sort.Strings(amstag.Response[ipaddr[0].String()])
	return strings.Join(amstag.Response[ipaddr[0].String()], ";")
}

func GetAction(id int) *Action {
	if res, found := ActionsCache.Get(strconv.Itoa(id)); found {
		return res.(*Action)
	} else {
		action := CurlAction(id)
		if action == nil {
			return nil
		}
		ActionsCache.Set(strconv.Itoa(id), action, cache.DefaultExpiration)
		return action
	}

}

func CurlAction(id int) *Action {
	if id <= 0 {
		return nil
	}
	res := HttpGetAction(id)
	return res
}

func GetEndpointTplGroups(tplId int, endpoint string) (grps string) {
	key := fmt.Sprintf("%d_%s", tplId, endpoint)
	if res, found := EndTplGroupCache.Get(key); found {
		return res.(string)
	} else {
		grps = HttpGetEndpointTplGroups(tplId, endpoint)
		if grps == "" {
			return ""
		}
		EndTplGroupCache.Set(key, grps, cache.DefaultExpiration)
		return
	}
}

func HttpGetEndpointTplGroups(tplId int, endpoint string) (grps string) {
	retryLimit := 3
	r_s := 0
	grps = ""
	for r_s < retryLimit {
		uri := fmt.Sprintf("%s/api/v1/tplgrp?tpl_id=%d&endpoint=%s", g.Config().Api.PlusApi, tplId, endpoint)
		req := httplib.Get(uri).SetTimeout(2*time.Second, 30*time.Second)
		token, _ := json.Marshal(map[string]string{
			"name": "falcon-alarm",
			"sig":  g.Config().Api.PlusApiToken,
		})
		req.Header("Apitoken", string(token))
		var res []*HostGroups
		err := req.ToJson(&res)
		if err != nil {
			log.Errorf("curl_tplgrp_round_%v_uri:%s,fail: %v,", r_s, uri, err)
			r_s += 1
			time.Sleep(time.Second)
		} else {
			if len(res) == 1 {
				grps = res[0].Name
				return
			}
			tL := make([]string, len(res))
			for _, v := range res {
				tL = append(tL, v.Name)
			}
			grps = strings.Join(tL, ",")
			return
		}
	}
	log.Errorf("HttpGetEndpointTplGroups_failed_for_tplId:%d,endpoint:%s", tplId, endpoint)
	return
}

func HttpGetAction(id int) *Action {
	retry_limit := 3
	r_s := 0
	for r_s < retry_limit {
		uri := fmt.Sprintf("%s/api/v1/action/%d", g.Config().Api.PlusApi, id)
		req := httplib.Get(uri).SetTimeout(2*time.Second, 30*time.Second)
		token, _ := json.Marshal(map[string]string{
			"name": "falcon-alarm",
			"sig":  g.Config().Api.PlusApiToken,
		})
		req.Header("Apitoken", string(token))
		var act Action
		err := req.ToJson(&act)
		if err != nil {
			log.Errorf("curl_falcon_action_round_%v_uri:%s,fail: %v,", r_s, uri, err)
			r_s += 1
			time.Sleep(time.Second)
		} else {
			return &act
		}
	}
	log.Errorf("HttpGetAction_finally_failed_for_action_id:%s", id)
	return nil
}
