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

	"github.com/open-falcon/falcon-plus/modules/judge/cron"
	"github.com/open-falcon/falcon-plus/modules/judge/g"
	"github.com/open-falcon/falcon-plus/modules/judge/http"
	"github.com/open-falcon/falcon-plus/modules/judge/rpc"
	"github.com/open-falcon/falcon-plus/modules/judge/store"
)

/*
	从hbs同步策略；						cron.SyncStrategies
	接收来自transfer的数据，并进行报警判断，	store/receiver.go/send
	产生的报警事件存入redis，等alarm读取；
	并提供http接口，提供api获取内存中的数据；	http/info.go
*/
func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	g.ParseConfig(*cfg)

	if g.Config().IsUnion {
		g.Logger.Infof("[main]-role:union")
	} else {
		g.Logger.Infof("[main]-role:normal")
	}

	// g.InitRedisConnPool()
	// judge生成的报警事件写入redis，alarm读取并报警
	g.InitRedisCluster()
	// 用于主动从hbs接收策略
	g.InitHbsClient()
	// 用于做组合告警的rpc call
	g.InitUnionJudgeClientAndRing()

	//16*16个key:list链，hold住历史数据，类似all(#3)要hold住3个点以上
	store.InitHistoryBigMap()

	// 提供api接口获取judge内存信息->info.go
	go http.Start()
	//接收来自transfer的数据
	go rpc.Start()

	// 通过HbsClient同步策略、表达式、过滤器（所有要judge的metrics）
	go cron.SyncStrategies()
	// 清理一周前的历史数据
	go cron.CleanStale()

	select {}
}
