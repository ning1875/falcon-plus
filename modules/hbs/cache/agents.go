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

package cache

// 每个agent心跳上来的时候立马更新一下数据库是没必要的
// 缓存起来，每隔一个小时写一次DB
// 提供http接口查询机器信息，排查重名机器的时候比较有用

import (
	"log"
	"sync"
	"time"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/db"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
)

type SafeAgents struct {
	sync.RWMutex
	M map[string]*model.AgentUpdateInfo
}

var Agents = NewSafeAgents()

var NewAgentUpgradeArgs = &model.AgentUpgradeArgs{WgetUrl: "", Version: ""}

//var UpgradeAgentMap = NewUpgradeAgents()

var NowAgentVersionMap = sync.Map{}

type UpgradeAgent struct {
	Timestamp   int64  `json:"timestamp"`
	LastVersion string `json:"lastversion"`
	ThisVersion string `json:"thisversion"`
}
type UpgradeAgents struct {
	Map sync.Map
}

func NewUpgradeAgents() *UpgradeAgents {
	return &UpgradeAgents{Map: sync.Map{}}
}

func NewSafeAgents() *SafeAgents {
	return &SafeAgents{M: make(map[string]*model.AgentUpdateInfo)}

}

func (this *SafeAgents) Put(req *model.AgentReportRequest) {
	val := &model.AgentUpdateInfo{
		LastUpdate:    time.Now().Unix(),
		ReportRequest: req,
	}

	agentInfo, exists := this.Get(req.Hostname)
	//log.Println("is_exists",exists)
	if !exists {
		//log.Println("cache_miss",req.Hostname,req.IP)
		// 不存在更新db
		go db.UpdateAgentNew(val)
	} else {
		//存在但是信息不同:只打印下信息不更新了
		if agentInfo.ReportRequest.IP != req.IP || agentInfo.ReportRequest.AgentVersion != req.AgentVersion {
			log.Printf("cache_hit_but_updb:%v %v %v %v %v %v %v %v",
				agentInfo.ReportRequest.Hostname,
				req.Hostname,
				agentInfo.ReportRequest.IP,
				req.IP,
				agentInfo.ReportRequest.AgentVersion,
				req.AgentVersion,
				agentInfo.ReportRequest.PluginVersion,
				req.PluginVersion)
		}
	}

	this.Lock()
	this.M[req.Hostname] = val
	defer this.Unlock()
}

func (this *SafeAgents) Get(hostname string) (*model.AgentUpdateInfo, bool) {
	this.RLock()
	defer this.RUnlock()
	val, exists := this.M[hostname]
	return val, exists
}

func (this *SafeAgents) Delete(hostname string) {
	this.Lock()
	defer this.Unlock()
	delete(this.M, hostname)
}

func (this *SafeAgents) Keys() []string {
	this.RLock()
	defer this.RUnlock()
	count := len(this.M)
	keys := make([]string, count)
	i := 0
	for hostname := range this.M {
		keys[i] = hostname
		i++
	}
	return keys
}

func (this *SafeAgents) AgentVersions() map[string]string {
	this.RLock()
	defer this.RUnlock()
	maps := make(map[string]string)
	i := 0
	for hostname := range this.M {
		value, _ := this.Get(hostname)
		maps[hostname] = value.ReportRequest.AgentVersion
		i++
	}
	return maps
}

func (this *SafeAgents) AgentVersionsNew() map[string]string {
	this.RLock()
	defer this.RUnlock()
	maps := make(map[string]string)
	i := 0
	for hostname := range this.M {
		value, _ := this.Get(hostname)
		maps[hostname] = value.ReportRequest.AgentVersion
		i++
	}
	return maps
}

func DeleteStaleAgents() {
	duration := time.Minute * 15 * time.Duration(g.Config().MapCleanInterval)
	for {
		time.Sleep(duration)
		deleteStaleAgents()
	}
}

func deleteStaleAgents() {
	// 一天都没有心跳的Agent，从内存中干掉
	before := time.Now().Unix() - 60*10*g.Config().MapCleanInterval
	keys := Agents.Keys()
	count := len(keys)
	if count == 0 {
		return
	}

	for i := 0; i < count; i++ {
		curr, _ := Agents.Get(keys[i])
		if curr.LastUpdate < before {
			NowAgentVersionMap.Delete(curr.ReportRequest.Hostname)
			Agents.Delete(curr.ReportRequest.Hostname)
		}
	}

}

func (this *UpgradeAgents) UpgradeAgentKeys() (len int, keys []string) {
	f := func(k, v interface{}) bool {
		len++
		keys = append(keys, k.(string))
		return true
	}
	this.Map.Range(f)
	return
}
