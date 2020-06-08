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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/open-falcon/falcon-plus/modules/agent/cron"
	"github.com/open-falcon/falcon-plus/modules/agent/funcs"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"github.com/open-falcon/falcon-plus/modules/agent/http"
)

func main() {

	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	check := flag.Bool("check", false, "check collector")

	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	if *check {
		funcs.CheckCollector()
		os.Exit(0)
	}

	g.ParseConfig(*cfg) //加载cfg文件到config *GlobalConfig

	if g.Config().Debug { //设置日志级别log.SetLevel
		g.InitLog("debug")
	} else {
		g.InitLog("info")
	}

	g.InitRootDir()
	g.InitLocalIp()    //获取本机IP
	g.InitRpcClients() //创建一个HbsClient *SingleConnRpcClient，此时未连接

	//funcs.InitMetricsOfFunction() //初始化metrics 对应的function

	//go cron.SyncDynamicMonitoringConfig() //从Hbs 同步动态配置到dmc

	funcs.BuildMappers() //构造metric 采集函数和采集周期列表
	//go funcs.BuildDynamicMappers() //定时更新动态采集的mapper
	go cron.SyncServiceConfig() //从Hbs 同步service开启proc采集

	// 定期更新本机Cpu和Disk状态，只保留最近两个值
	go cron.InitDataHistory()
	// 开始自升级指令监听
	go g.AgentSelfUpgrade()

	// 上报本机状态
	cron.ReportAgentStatus()
	// 监听是否启动新采集方法
	//go cron.DynamicMapperWatcher()

	// 同步插件
	cron.SyncMinePlugins()
	// 从Hbs 同步监控端口、路径、进程和URL
	cron.SyncBuiltinMetrics()
	// 后门调试agent,允许执行shell指令的ip列表
	cron.SyncTrustableIps()

	// 开始数据次采集
	cron.Collect()

	// 启动dashboard server
	go http.Start()

	select {}

}

/*
cron：间隔执行的代码，即定时任务

funcs：信息采集

g:全局数据结构

http：简单的dashboard的server，获取单机监控指标数据

plugins：插件处理机制

public：静态资源文件
*/
