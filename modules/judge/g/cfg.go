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

type RpcConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type HbsConfig struct {
	Servers  []string `json:"servers"`
	Timeout  int64    `json:"timeout"`
	Interval int64    `json:"interval"`
}

type UnionJudge struct {
	Enabled     bool              `json:"enabled"`
	Batch       int               `json:"batch"`
	ConnTimeout int               `json:"connTimeout"`
	CallTimeout int               `json:"callTimeout"`
	MaxConns    int               `json:"maxConns"`
	MaxIdle     int               `json:"maxIdle"`
	Replicas    int32             `json:"replicas"`
	Cluster     map[string]string `json:"cluster"`
}

type RedisConfig struct {
	RedisClusterNodes []string `json:"redis_cluster_nodes"`
	Dsn               string   `json:"dsn"`
	MaxIdle           int      `json:"maxIdle"`
	ConnTimeout       int      `json:"connTimeout"`
	ReadTimeout       int      `json:"readTimeout"`
	WriteTimeout      int      `json:"writeTimeout"`
}

type AlarmConfig struct {
	Enabled      bool         `json:"enabled"`
	MinInterval  int64        `json:"minInterval"`
	QueuePattern string       `json:"queuePattern"`
	Redis        *RedisConfig `json:"redis"`
}

type GlobalConfig struct {
	Debug       bool         `json:"debug"`
	LogLevel    string       `json:"log_level"`
	DebugHost   string       `json:"debugHost"`
	Remain      int          `json:"remain"`
	Http        *HttpConfig  `json:"http"`
	Rpc         *RpcConfig   `json:"rpc"`
	Hbs         *HbsConfig   `json:"hbs"`
	UnionJudgeS *UnionJudge  `json:"union_judge_s"`
	Alarm       *AlarmConfig `json:"alarm"`
	// 这个字段代表这个实例是否是组合策略judge
	// 组合策略judge不需要同步单个策略
	IsUnion bool `json:"is_union"`
}

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

	configLock.Lock()
	defer configLock.Unlock()

	config = &c

	log.Println("read config file:", cfg, "successfully")
}
