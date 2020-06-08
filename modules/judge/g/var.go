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

package g

import (
	"sync"
	"time"

	"github.com/toolkits/consistent/rings"

	backend "github.com/open-falcon/falcon-plus/common/backend_pool"
	"github.com/open-falcon/falcon-plus/common/model"
	cutils "github.com/open-falcon/falcon-plus/common/utils"
	nset "github.com/toolkits/container/set"
)

type SafeStrategyMap struct {
	sync.RWMutex
	// endpoint:metric => [strategy1, strategy2 ...]
	M map[string][]model.Strategy
}

type SafeUnionStrategies struct {
	sync.RWMutex
	M map[int][]*model.Strategy
}

type SafeExpressionMap struct {
	sync.RWMutex
	// metric:tag1 => [exp1, exp2 ...]
	// metric:tag2 => [exp1, exp2 ...]
	M map[string][]*model.Expression
}

type SafeEventMap struct {
	sync.RWMutex
	M map[string]*model.Event
}

type SafeFilterMap struct {
	sync.RWMutex
	M map[string]string
}

var (
	HbsClient           *SingleConnRpcClient
	JudgeNodeRing       *rings.ConsistentHashNodeRing
	StrategyMap         = &SafeStrategyMap{M: make(map[string][]model.Strategy)}
	UnionStrategyMap    = &SafeUnionStrategies{M: make(map[int][]*model.Strategy)}
	ExpressionMap       = &SafeExpressionMap{M: make(map[string][]*model.Expression)}
	LastEvents          = &SafeEventMap{M: make(map[string]*model.Event)}
	FilterMap           = &SafeFilterMap{M: make(map[string]string)}
	UnionJudgeConnPools *backend.SafeRpcConnPools
	UnionLastEvents     = &SafeEventMap{M: make(map[string]*model.Event)}
)

func InitHbsClient() {
	HbsClient = &SingleConnRpcClient{
		RpcServers: Config().Hbs.Servers,
		Timeout:    time.Duration(Config().Hbs.Timeout) * time.Millisecond,
	}
}

func InitUnionJudgeClientAndRing() {
	JudgeNodeRing = rings.NewConsistentHashNodesRing(Config().UnionJudgeS.Replicas, cutils.KeysOfMap(Config().UnionJudgeS.Cluster))

	graphInstances := nset.NewSafeSet()

	for _, addr := range Config().UnionJudgeS.Cluster {
		graphInstances.Add(addr)
	}

	UnionJudgeConnPools = backend.CreateSafeRpcConnPools(Config().UnionJudgeS.MaxConns, Config().UnionJudgeS.MaxIdle,
		Config().UnionJudgeS.ConnTimeout, Config().UnionJudgeS.CallTimeout, graphInstances.ToSlice())
}

func (this *SafeUnionStrategies) ReInit(m map[int][]*model.Strategy) {
	this.Lock()
	defer this.Unlock()
	this.M = m
}

func (this *SafeUnionStrategies) GetMap() map[int][]*model.Strategy {
	this.RLock()
	defer this.RUnlock()
	return this.M
}

func (this *SafeUnionStrategies) GetById(UnionId int) []*model.Strategy {
	this.RLock()
	defer this.RUnlock()
	return this.M[UnionId]
}

func (this *SafeStrategyMap) ReInit(m map[string][]model.Strategy) {
	this.Lock()
	defer this.Unlock()
	this.M = m
}

func (this *SafeStrategyMap) Get() map[string][]model.Strategy {
	this.RLock()
	defer this.RUnlock()
	return this.M
}

func (this *SafeExpressionMap) ReInit(m map[string][]*model.Expression) {
	this.Lock()
	defer this.Unlock()
	this.M = m
}

func (this *SafeExpressionMap) Get() map[string][]*model.Expression {
	this.RLock()
	defer this.RUnlock()
	return this.M
}

func (this *SafeEventMap) Get(key string) (*model.Event, bool) {
	this.RLock()
	defer this.RUnlock()
	event, exists := this.M[key]
	return event, exists
}

func (this *SafeEventMap) Set(key string, event *model.Event) {
	this.Lock()
	defer this.Unlock()
	this.M[key] = event
}

func (this *SafeFilterMap) ReInit(m map[string]string) {
	this.Lock()
	defer this.Unlock()
	this.M = m
}

func (this *SafeFilterMap) Exists(key string) bool {
	this.RLock()
	defer this.RUnlock()
	if _, ok := this.M[key]; ok {
		return true
	}
	return false
}
