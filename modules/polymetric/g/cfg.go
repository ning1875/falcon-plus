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
	RedisClusterNodes []string `json:"redis_cluster_nodes"`
}

type ApiConfig struct {
	Group   string `json:"group"`
	PushApi string `json:"push_api"`
}

type KafkaConfig struct {
	KafkaNodes []string `json:"kafka_nodes"`
	TopicName  string   `json:"topic_name"`
}

type FalconPortalConfig struct {
	Addr string `json:"addr"`
	Idle int    `json:"idle"`
	Max  int    `json:"max"`
}

type FalconOutlierConfig struct {
	Addr string `json:"addr"`
	Idle int    `json:"idle"`
	Max  int    `json:"max"`
}

type GlobalConfig struct {
	LogLevel      string               `json:"log_level"`
	GaussianNum   float64              `json:"gaussian_num"`
	Rpc           *RpcConfig           `json:"rpc"`
	FalconPortal  *FalconPortalConfig  `json:"falcon_portal"`
	FalconOutlier *FalconOutlierConfig `json:"falcon_outlier"`
	Http          *HttpConfig          `json:"http"`
	Redis         *RedisConfig         `json:"redis"`
	Api           *ApiConfig           `json:"api"`
	Kafka         *KafkaConfig         `json:"kafka"`
	NumpRpc       NumpConfig           `json:"nump_rpc"`
	Prome         *PromeConfig         `json:"prome"`
}

type NumpConfig struct {
	Url     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type PromeConfig struct {
	Enabled      bool   `json:"enabled"`
	Address      string `json:"address"`
	PushInterval int    `json:"push_interval"`
}

type RpcConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
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

	if c.GaussianNum == 0 {
		//gaussian num 默认值1.5
		c.GaussianNum = 1.5
	}

	configLock.Lock()
	defer configLock.Unlock()
	config = &c
	log.Println("read config file:", cfg, "successfully")
}
