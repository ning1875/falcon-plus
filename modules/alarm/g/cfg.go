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

	"github.com/toolkits/file"
)

type HttpConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type RedisConfig struct {
	Addr              string   `json:"addr"`
	RedisClusterNodes []string `json:"redis_cluster_nodes"`
	MaxIdle           int      `json:"maxIdle"`
	HighQueues        []string `json:"highQueues"`
	LowQueues         []string `json:"lowQueues"`
	UserIMQueue       string   `json:"userIMQueue"`
	UserSmsQueue      string   `json:"userSmsQueue"`
	UserMailQueue     string   `json:"userMailQueue"`
}

type ApiConfig struct {
	Sms                      string `json:"sms"`
	Mail                     string `json:"mail"`
	Phone                    string `json:"phone"`
	Dashboard                string `json:"dashboard"`
	PlusApi                  string `json:"plus_api"`
	MainApi                  string `json:"main_api"`
	PlusApiToken             string `json:"plus_api_token"`
	IM                       string `json:"im"`
	LarkTenantAccessTokenUrl string `json:"lark_tenant_access_token_url"`
}

type FalconPortalConfig struct {
	Addr string `json:"addr"`
	Idle int    `json:"idle"`
	Max  int    `json:"max"`
}

type WorkerConfig struct {
	IM    int `json:"im"`
	Sms   int `json:"sms"`
	Mail  int `json:"mail"`
	Phone int `json:"phone"`
}

type HousekeeperConfig struct {
	EventRetentionDays int `json:"event_retention_days"`
	EventDeleteBatch   int `json:"event_delete_batch"`
}

type GlobalConfig struct {
	LogLevel     string              `json:"log_level"`
	FalconPortal *FalconPortalConfig `json:"falcon_portal"`
	Http         *HttpConfig         `json:"http"`
	Redis        *RedisConfig        `json:"redis"`
	Api          *ApiConfig          `json:"api"`
	Worker       *WorkerConfig       `json:"worker"`
	Housekeeper  *HousekeeperConfig  `json:"Housekeeper"`
	//LarkConfig      *LarkConfig         `json:"lark_config"`
	LarkBotTokens   []string `json:"lark_bot_tokens"`
	AlarmRetryTimes int      `json:"alarm_retry_times"`
	AlarmApi        string   `json:"alarm_api"`
}

//type LarkConfig struct {
//}

var (
	ConfigFile string
	config     *GlobalConfig
	configLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
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
	if c.AlarmRetryTimes <= 0 {
		c.AlarmRetryTimes = 2
	}
	if c.Api.MainApi == "" {
		c.Api.MainApi = "http://falcon-api.d.xxx.com:8080"
	}
	configLock.Lock()
	defer configLock.Unlock()
	config = &c
	log.Println("read config file:", cfg, "successfully")
}
