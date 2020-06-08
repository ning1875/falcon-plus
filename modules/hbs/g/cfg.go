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
	"encoding/json"
	"log"
	"sync"
	"time"

	"os"

	"github.com/toolkits/file"
)

type HttpConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type GlobalConfig struct {
	Debug                      bool        `json:"debug"`
	Hosts                      string      `json:"hosts"`
	Database                   string      `json:"database"`
	DatabaseRo                 string      `json:"database_ro"`
	MaxConns                   int         `json:"maxConns"`
	MaxIdle                    int         `json:"maxIdle"`
	Listen                     string      `json:"listen"`
	Trustable                  []string    `json:"trustable"`
	Http                       *HttpConfig `json:"http"`
	MapCleanInterval           int64       `json:"map_clean_interval"`
	AgentUpdateQueueLength     int         `json:"agent_update_queue_length"`
	RedisAgentUpgradeSetExpire int         `json:"redis_agent_upgrade_set_expire"`
	RedisClusterNodes          []string    `json:"redis_cluster_nodes"`
	CacheTtl                   int64       `json:"cache_ttl"`
}

var (
	ConfigFile     string
	config         *GlobalConfig
	configLock     = new(sync.RWMutex)
	HostName, _    = Hostname()
	GlobalCacheTtl time.Duration
)

func Config() *GlobalConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func Hostname() (string, error) {

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("ERROR: os.Hostname() fail", err)
	}
	return hostname, err
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}
	//如果配置文件中没有设置,则给一个默认值
	if c.MapCleanInterval <= 0 {
		c.MapCleanInterval = MapCleanInterval
	}

	if c.CacheTtl <= 0 {
		c.CacheTtl = 60
	}
	GlobalCacheTtl = time.Duration(c.CacheTtl) * time.Second
	//每个hbs并发更新的agent上限
	if c.AgentUpdateQueueLength <= 0 {
		c.AgentUpdateQueueLength = AgentUpdateQueueLength
	}
	//整个升级任务的并发更新set的超时时间
	if c.RedisAgentUpgradeSetExpire <= 0 {
		c.RedisAgentUpgradeSetExpire = REDIS_AGENT_UPGRADE_SET_EXPIRE
	}

	log.Println("MapCleanInterval:AgentUpdateQueueLength", c.MapCleanInterval, c.AgentUpdateQueueLength)
	configLock.Lock()
	defer configLock.Unlock()

	config = &c

	log.Println("read config file:", cfg, "successfully")
}
