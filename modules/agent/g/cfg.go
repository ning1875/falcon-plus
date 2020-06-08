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
	"fmt"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"

	"strings"

	"github.com/toolkits/file"
)

type PluginConfig struct {
	Enabled bool   `json:"enabled"`
	Dir     string `json:"dir"`
	Git     string `json:"git"`
	LogDir  string `json:"logs"`
}

type HeartbeatConfig struct {
	Enabled  bool   `json:"enabled"`
	Addr     string `json:"addr"`
	Interval int    `json:"interval"`
	Timeout  int    `json:"timeout"`
}

type TransferConfig struct {
	Enabled  bool     `json:"enabled"`
	Addrs    []string `json:"addrs"`
	Interval int      `json:"interval"`
	Timeout  int      `json:"timeout"`
}

type HttpConfig struct {
	Enabled  bool   `json:"enabled"`
	Listen   string `json:"listen"`
	Backdoor bool   `json:"backdoor"`
}

type CollectorConfig struct {
	IfacePrefix []string `json:"ifacePrefix"`
	MountPoint  []string `json:"mountPoint"`
}

type MetricLevelIntervalConfig struct {
	Level1 int `json:"level_1"`
	Level2 int `json:"level_2"`
	Level3 int `json:"level_3"`
	Level4 int `json:"level_4"`
}

type GlobalConfig struct {
	Debug                     bool                       `json:"debug"`
	Hostname                  string                     `json:"hostname"`
	IP                        string                     `json:"ip"`
	Plugin                    *PluginConfig              `json:"plugin"`
	Heartbeat                 *HeartbeatConfig           `json:"heartbeat"`
	Transfer                  *TransferConfig            `json:"transfer"`
	Http                      *HttpConfig                `json:"http"`
	Collector                 *CollectorConfig           `json:"collector"`
	DefaultTags               map[string]string          `json:"default_tags"`
	IgnoreMetrics             map[string]bool            `json:"ignore"`
	MetricLevelIntervalConfig *MetricLevelIntervalConfig `json:"metric_level_interval"`
	AppBaseDir                string                     `json:"app_base_dir"`
	SelfUpgrade               bool                       `json:"self_upgrade"`
	CpuPerCoreCollect         bool                       `json:"cpu_per_core_collect"`
}

type Service struct {
	Name string
	Pids []int
}

var (
	ConfigFile         string
	config             *GlobalConfig
	lock               = new(sync.RWMutex)
	dmc                = make(map[string]int)
	dmcLock            = new(sync.RWMutex)
	ServiceChangeSigns = make(map[string]chan int)
)

func GetDynamicMonitoringConfig() map[string]int {
	dmcLock.RLock()
	defer dmcLock.RUnlock()
	return dmc
}

func SetDynamicMonitoringConfig(cfg map[string]int) {
	dmcLock.Lock()
	defer dmcLock.Unlock()
	dmc = cfg
}

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

func validateHostName(name string) bool {
	if !strings.HasPrefix(name, "n") {
		fmt.Println(1, name)
		return false
	}

	if len(strings.Split(name, "-")) != 3 {
		fmt.Println(2, name)
		return false
	}
	return true
}

func convertIpToHostName(ip string) string {
	ips := strings.Split(ip, ".")
	ends := "n"
	for index, i := range ips {
		switch index {
		case 0:
			continue
		case 1:
			ends += fmt.Sprintf("%s-", i)
		case 2:
			ends += fmt.Sprintf("%03s-", i)
		case 3:
			ends += fmt.Sprintf("%03s", i)
		}

	}
	return ends

}

func Hostname() (string, error) {
	hostname := Config().Hostname
	if hostname != "" {
		return hostname, nil
	}

	if os.Getenv("FALCON_ENDPOINT") != "" {
		hostname = os.Getenv("FALCON_ENDPOINT")
		return hostname, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("ERROR: os.Hostname() fail", err)
	}
	// 判断hostname是否标准 n3-021-225
	// 不是的话通过ip转换
	if !validateHostName(hostname) {
		hostname = convertIpToHostName(LocalIp)
		if !validateHostName(hostname) {
			log.Errorf("ERROR: convertIpToHostName() fail", err)
		}
	}

	return hostname, err
}

func IP() string {
	ip := Config().IP
	if ip != "" {
		// use ip in configuration
		return ip
	}

	if len(LocalIp) > 0 {
		ip = LocalIp
	}

	return ip
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
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
	//设置各个级别的采集level
	if c.MetricLevelIntervalConfig == nil {
		tmp := &MetricLevelIntervalConfig{}
		tmp.Level1 = 10
		tmp.Level2 = 30
		tmp.Level3 = 60
		tmp.Level4 = 120
		c.MetricLevelIntervalConfig = tmp
	}
	// falcon-agent 家目录
	if c.AppBaseDir == "" {
		c.AppBaseDir = "/opt/open-falcon/agent"
	}
	// 配置自更新开关
	if c.SelfUpgrade == false {
		c.SelfUpgrade = true
	}
	lock.Lock()
	defer lock.Unlock()

	config = &c

	log.Println("read config file:", cfg, "successfully")
}
