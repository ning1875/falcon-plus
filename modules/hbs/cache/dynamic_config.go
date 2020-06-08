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
	"sync"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/db"
)

type SafeDynamicConfig struct {
	sync.RWMutex
	M map[string][]*model.DynamicConfig
}

var DynamicConfig = &SafeDynamicConfig{M: make(map[string][]*model.DynamicConfig)}

func (this *SafeDynamicConfig) Get() map[string][]*model.DynamicConfig {
	this.RLock()
	defer this.RUnlock()
	return this.M
}

func (this *SafeDynamicConfig) Init() {
	m, err := db.QueryDynamicConfig()
	if err != nil {
		return
	}

	this.Lock()
	defer this.Unlock()
	this.M = m
}
