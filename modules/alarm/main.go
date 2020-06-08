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
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/open-falcon/falcon-plus/modules/alarm/cron"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/alarm/http"
	"github.com/open-falcon/falcon-plus/modules/alarm/model"
	"github.com/open-falcon/falcon-plus/modules/alarm/redi"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	help := flag.Bool("h", false, "help")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	g.ParseConfig(*cfg)

	g.InitLog(g.Config().LogLevel)
	if g.Config().LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	//g.InitRedisConnPool()
	redi.Init()
	model.InitDatabase()
	cron.InitSenderWorker()

	// 同步redis中的报警屏蔽项目到alarm map中
	go cron.RefreshBlockMonitor()
	go http.Start()
	go cron.ReadHighEvent()
	go cron.ReadLowEvent()
	// 合并低优先级的短信
	go cron.CombineSms()
	// 合并低优先级的邮件
	go cron.CombineMail()
	// 合并低优先级的lark
	go cron.CombineIM()
	// 从 /im pop出消息 发送并将失败的push 到 /failed/im
	go cron.ConsumeIM()
	// 从 /failed/im pop出消息 发送  /finallyfailed/im
	go cron.ConsumeFailedIM()
	go cron.ConsumeSms()
	go cron.ConsumeMail()
	go cron.ConsumePhone()
	go cron.CleanExpiredEvent()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println()
		redi.RedisCluster.Close()
		//g.RedisConnPool.Close()
		os.Exit(0)
	}()

	select {}
}
