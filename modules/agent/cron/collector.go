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

package cron

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/agent/funcs"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
)

func InitDataHistory() {
	for {
		funcs.UpdateCpuStat()
		funcs.UpdateDiskStats()
		funcs.UpdateProcCpuStat(Services)
		time.Sleep(g.COLLECT_INTERVAL)
	}
}

//Colleet：配置信息读取，读取Mapper中的FuncsAndInterval，根据func调用采集函数，采集所有信息（并非先过滤采集项），从所有采集到的数据中过滤ignore的项，并上报到transfer。
func Collect() {
	// 配置信息判断
	if !g.Config().Transfer.Enabled {
		return
	}

	if len(g.Config().Transfer.Addrs) == 0 {
		return
	}
	// 读取mapper中的FuncsAndInterval集,并通过不同的goroutine运行
	for _, v := range funcs.Mappers {
		go collect(int64(v.Interval), v.Fs)
		//go collectService(int64(v.Interval), v.Fs, v.Service)   起一个新进程采集proc信息，参数 service:[]pids
	}

}

//接收变更信号，停止老的采集方法，启动新的采集方法
func DynamicMapperWatcher() {
	ctx, cancel := context.WithCancel(context.Background())
	for {
		log.Infoln("funcs.GetDynamicMappers() :", funcs.GetDynamicMappers())
		for _, v := range funcs.GetDynamicMappers() {
			go collectDynamic(ctx, int64(v.Interval), v.Fs)
		}
		<-funcs.DynamicMapperConfChangeSign
		cancel()
		ctx, cancel = context.WithCancel(context.Background())
	}

}

//接收变更信号，查看是否需要启动且启动哪个service proc采集
func CreateProcMapperWatcher(services map[string][]string) {
	log.Infoln("新增proc 采集 services: ", services)
	Rootctx := context.Background()
	for name, pids := range services { //多个service并行

		go func(n string, p []string) {
			//每一个service 开启一个ctx看管，如果当前service消失，则停止
			ctx, cancel := context.WithCancel(Rootctx)
			log.Infoln("Proc collect start for service:pid", n, p)
			for _, v := range funcs.ProcMappers {
				//v.fs= []fn(string, []int)
				go collectProcDynamic(ctx, int64(v.Interval), n, p, v.Fs)
			}
			//如果service不存在了，则停止采集
			<-g.ServiceChangeSigns[n]
			cancel()
		}(name, pids)
	}
}

func collectProcDynamic(ctx context.Context, sec int64, name string, pids []string, fns []func(string, []string) []*model.MetricValue) {
	log.Infoln("CollectProcDynamic，时间间隔：", sec, name)
	ExecProcDynamicFunc(sec, name, pids, fns)
	// 启动断续器,间隔执行
	t := time.NewTicker(time.Second * time.Duration(sec))
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Infoln("收到proc context信号，退出proc collect goroutine", name)
			return
		case <-t.C:
			ExecProcDynamicFunc(sec, name, pids, fns)
		}
	}

}
func ExecProcDynamicFunc(sec int64, name string, pids []string, fns []func(string, []string) []*model.MetricValue) {
	hostname, err := g.Hostname()
	if err != nil {
		return
	}

	mvs := []*model.MetricValue{}
	ignoreMetrics := g.Config().IgnoreMetrics
	// 读取采集的metric名单
	// 从funcs的list中取出每个采集函数
	for _, fn := range fns {
		items := fn(name, pids)
		if items == nil {
			continue
		}

		if len(items) == 0 {
			continue
		}
		for _, mv := range items {
			if b, ok := ignoreMetrics[mv.Metric]; ok && b {
				continue
			} else {
				mvs = append(mvs, mv)
			}
		}
	}

	// 获取上报时间
	now := time.Now().Unix()
	// 设置上报采集项的间隔、agent主机、上报时间
	for j := 0; j < len(mvs); j++ {
		mvs[j].Step = sec
		mvs[j].Endpoint = hostname
		mvs[j].Timestamp = now
	}
	// 调用transfer发送采集数据
	log.Debugln("Proc采集结果:", name, pids, mvs)
	g.SendToTransfer(mvs)
}

func collectDynamic(ctx context.Context, sec int64, fn func() []*model.MetricValue) {
	log.Infoln("启动collectDynamic，时间间隔：", sec)
	t := time.NewTicker(time.Second * time.Duration(sec))
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Infoln("收到context信号，退出采集goroutine")
			return
		case <-t.C:
			ExecDynamicFunc(sec, fn)
		}
	}

}
func ExecDynamicFunc(sec int64, fn func() []*model.MetricValue) {
	hostname, err := g.Hostname()
	if err != nil {
		return
	}

	mvs := []*model.MetricValue{}

	items := fn()
	if items == nil {
		return
	}

	if len(items) == 0 {
		return
	}
	//// 读取采集数据,根据忽略的metric忽略部分采集数据
	//for _, mv := range items {
	//	mvs = append(mvs, mv)
	//}
	legalMetrics := funcs.GetDynamicMetrics()

	// 读取采集数据,根据采集metric忽略部分采集数据
	for _, mv := range items {
		if _, ok := legalMetrics[mv.Metric]; ok {
			mvs = append(mvs, mv)
		}

	}

	now := time.Now().Unix()
	// 设置上报采集项的间隔、agent主机、上报时间
	for j := 0; j < len(mvs); j++ {
		mvs[j].Step = sec
		mvs[j].Endpoint = hostname
		mvs[j].Timestamp = now
		if len(mvs[j].Tags) <= 0 {
			mvs[j].Tags = "DynamicMetric=true"
		} else {
			mvs[j].Tags += ",DynamicMetric=true"
		}
	}
	// 调用transfer发送采集数据
	g.SendToTransfer(mvs)
}

// 间隔采集信息
func collect(sec int64, fns []func() []*model.MetricValue) {
	// 启动断续器,间隔执行
	t := time.NewTicker(time.Second * time.Duration(sec))
	defer t.Stop()
	for {
		<-t.C

		hostname, err := g.Hostname()
		if err != nil {
			continue
		}

		mvs := []*model.MetricValue{}
		ignoreMetrics := g.Config().IgnoreMetrics
		// 读取采集的metric名单
		// 从funcs的list中取出每个采集函数
		for _, fn := range fns {
			items := fn()
			if items == nil {
				continue
			}

			if len(items) == 0 {
				continue
			}
			for _, mv := range items {
				if b, ok := ignoreMetrics[mv.Metric]; ok && b {
					continue
				} else {
					mvs = append(mvs, mv)
				}
			}
		}

		// 获取上报时间
		now := time.Now().Unix()
		// 设置上报采集项的间隔、agent主机、上报时间
		for j := 0; j < len(mvs); j++ {
			mvs[j].Step = sec
			mvs[j].Endpoint = hostname
			mvs[j].Timestamp = now
		}
		// 调用transfer发送采集数据
		g.SendToTransfer(mvs)

	}
}
