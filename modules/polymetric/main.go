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
	"github.com/open-falcon/falcon-plus/common/sdk/sender"
	"github.com/open-falcon/falcon-plus/modules/polymetric/cron"
	"github.com/open-falcon/falcon-plus/modules/polymetric/g"
	"github.com/open-falcon/falcon-plus/modules/polymetric/kafka"
	"github.com/open-falcon/falcon-plus/modules/polymetric/model"
	"github.com/open-falcon/falcon-plus/modules/polymetric/redi"
	"github.com/open-falcon/falcon-plus/modules/polymetric/rpc"
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
	var Debug bool
	g.InitLog(g.Config().LogLevel)
	if g.Config().LogLevel != "debug" {
		Debug = true
		gin.SetMode(gin.ReleaseMode)
	}

	redi.Init()
	model.InitDatabase()
	//outlier.InitDataDB()
	kafka.InitKafkaConn()
	go cron.RunAllPoly()

	go rpc.StartRpc()

	// start sender
	sender.Debug = Debug
	sender.PostPushUrl = g.Config().Api.PushApi
	sender.StartSender()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		redi.RedisCluster.Close()
		kafka.KafkaAsyncProducer.Close()
		os.Exit(0)
	}()
	select {}
}
