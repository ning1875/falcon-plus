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

package rpc

import (
	"errors"
	"time"

	"sort"

	"github.com/open-falcon/falcon-plus/common/model"
	cutils "github.com/open-falcon/falcon-plus/common/utils"
	"github.com/open-falcon/falcon-plus/modules/judge/g"
	"github.com/open-falcon/falcon-plus/modules/judge/store"
)

type Judge int

func (this *Judge) Ping(req model.NullRpcRequest, resp *model.SimpleRpcResponse) error {
	return nil
}

// transfer调度了send
func (this *Judge) Send(items []*model.JudgeItem, resp *model.SimpleRpcResponse) error {
	remain := g.Config().Remain //safelinklist的最大长度
	// 把当前时间的计算放在最外层，是为了减少获取时间时的系统调用开销
	now := time.Now().Unix()
	for _, item := range items {
		if item.Timestamp <= 0 {
			g.Logger.Warningf("item.Timestamp_eq_0 item:%+v", item)
			continue
		}
		filterSS := g.CheckNeedJudge(item.Endpoint, item.Metric)
		//g.Logger.Infof("item:%+v,filterSS:%+v", item, filterSS)
		if len(filterSS) <= 0 {
			continue
		}
		//endpoint、metric、tags 计算md5
		pk := item.PrimaryKey()
		store.HistoryBigMap[pk[0:2]].PushFrontAndMaintain(pk, item, remain, now)
	}
	return nil
}

//func (this *Judge) Send(items []*model.JudgeItem, resp *model.SimpleRpcResponse) error {
//	remain := g.Config().Remain //safelinklist的最大长度
//	// 把当前时间的计算放在最外层，是为了减少获取时间时的系统调用开销
//	now := time.Now().Unix()
//	for _, item := range items {
//
//		exists := g.FilterMap.Exists(item.Metric)
//		if !exists {
//			continue
//		}
//		//endpoint、metric、tags 计算md5
//		pk := item.PrimaryKey()
//		store.HistoryBigMap[pk[0:2]].PushFrontAndMaintain(pk, item, remain, now)
//	}
//	return nil
//}

// 接收组合报警中的item
func (this *Judge) UnionJudge(event *model.Event, resp *model.SimpleRpcResponse) error {
	if event.Strategy.UnionStrategyId <= 0 {
		return errors.New("not union strategy")
	}

	thisCounter := cutils.Counter(event.Strategy.Metric, event.Strategy.Tags)
	unionIDSlice := []string{""}
	//thisUnionStrategies := g.UnionStrategyMap.GetById(event.Strategy.UnionStrategyId)
	thisUnionStrategies := g.GetOneUnionStrategy(event.Strategy.UnionStrategyId)

	for _, i := range thisUnionStrategies {
		counter := cutils.Counter(i.Metric, i.Tags)
		unionIDSlice = append(unionIDSlice, counter)

	}
	sort.Strings(unionIDSlice)
	unionID := ""
	for _, i := range unionIDSlice {
		unionID += i
	}
	//g.Logger.Infof("UnionJudge_call_get_res:%+v unionID:s", event, unionID)
	unionJudge, exists := store.UnionJudgeHistoryMap.Get(unionID)

	if !exists {
		// 这个组合报警中的项目第一次过来
		unionJudge = store.NewUnionJudgeEvent()

	}
	unionJudge.SetAndJudge(thisCounter, event)
	store.UnionJudgeHistoryMap.Set(unionID, unionJudge)
	return nil
}
