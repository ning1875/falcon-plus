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

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/db"
	"github.com/toolkits/container/set"
)

type SafeStrategies struct {
	sync.RWMutex
	M map[int]*model.Strategy
}

type SafeUnionStrategies struct {
	sync.RWMutex
	M map[int][]*model.Strategy
}

var (
	Strategies      = &SafeStrategies{M: make(map[int]*model.Strategy)}
	UnionStrategies = &SafeUnionStrategies{M: make(map[int][]*model.Strategy)}
)

func (this *SafeStrategies) GetMap() map[int]*model.Strategy {
	this.RLock()
	defer this.RUnlock()
	return this.M
}

func (this *SafeStrategies) Init(tpls map[int]*model.Template) {
	m, unionM, err := db.QueryStrategies(tpls)
	if err != nil {
		return
	}

	this.Lock()
	defer this.Unlock()
	UnionStrategies.Init(unionM)
	this.M = m
}

func (this *SafeUnionStrategies) Init(unM map[int][]*model.Strategy) {
	this.Lock()
	defer this.Unlock()
	this.M = unM
}

func (this *SafeUnionStrategies) Get() map[int][]*model.Strategy {
	this.RLock()
	defer this.RUnlock()
	return this.M
}

func GetBuiltinMetrics(hostname string) ([]*model.BuiltinMetric, error) {
	ret := []*model.BuiltinMetric{}
	hid, exists := HostMap.GetID(hostname)
	if !exists {
		return ret, nil
	}

	gids, exists := HostGroupsMap.GetGroupIds(hid)
	if !exists {
		return ret, nil
	}

	// 根据gids，获取绑定的所有tids
	tidSet := set.NewIntSet()
	for _, gid := range gids {
		tids, exists := GroupTemplates.GetTemplateIds(gid)
		if !exists {
			continue
		}

		for _, tid := range tids {
			tidSet.Add(tid)
		}
	}

	tidSlice := tidSet.ToSlice()
	if len(tidSlice) == 0 {
		return ret, nil
	}

	// 继续寻找这些tid的ParentId
	allTpls := TemplateCache.GetMap()
	for _, tid := range tidSlice {
		pids := ParentIds(allTpls, tid)
		for _, pid := range pids {
			tidSet.Add(pid)
		}
	}

	// 终于得到了最终的tid列表
	tidSlice = tidSet.ToSlice()

	// 把tid列表用逗号拼接在一起
	count := len(tidSlice)
	tidStrArr := make([]string, count)
	for i := 0; i < count; i++ {
		tidStrArr[i] = strconv.Itoa(tidSlice[i])
	}

	return db.QueryBuiltinMetrics(strings.Join(tidStrArr, ","))
}

//查找父模板id
func ParentIds(allTpls map[int]*model.Template, tid int) (ret []int) {
	depth := 0
	for {
		if tid <= 0 {
			break
		}

		if t, exists := allTpls[tid]; exists { //当前模板存在，则放到返回列表，当前策略设置为父策略
			ret = append(ret, tid)
			tid = t.ParentId
		} else {
			break
		}

		depth++
		if depth == 10 {
			log.Println("[ERROR] template inherit cycle. id:", tid)
			return []int{}
		}
	}

	sz := len(ret)
	if sz <= 1 {
		return
	}
	// 将告警模板id 倒序，保证最下层子节点在最后，最上层父节点在最前
	desc := make([]int, sz)
	for i, item := range ret {
		j := sz - i - 1
		desc[j] = item
	}

	return desc
}
