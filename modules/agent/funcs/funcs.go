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

package funcs

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
)

//FuncsAndInterval 拆分不同的采集函数集，方便通过不同goroutine运行
type FuncsAndInterval struct {
	Fs       []func() []*model.MetricValue
	Interval int
}

type ProcFuncsAndInterval struct {
	Fs       []func(string, []string) []*model.MetricValue
	Interval int
}

type DynamicFuncsAndInterval struct {
	Fs       func() []*model.MetricValue
	Interval int
}

var (
	Mappers                     []FuncsAndInterval
	ProcMappers                 []ProcFuncsAndInterval
	DynamicMappersLock          = new(sync.RWMutex)
	DynamicMappers              []DynamicFuncsAndInterval
	DynamicMappersJson          = ""
	DynamicMapperConfChangeSign = make(chan int)
	MetricsOfFunctionName       = make(map[string]string)
	ProcMetricsOfFunctionName   = make(map[string]string)
	FunctionNameOfFunction      = make(map[string]func() []*model.MetricValue)
	ProcFunctionNameOfFunction  = make(map[string]func(string, []string) []*model.MetricValue)
	DynamicMetrics              = []string{}
	DynamicMetricsLock          = new(sync.RWMutex)
)

func SetDynamicMappers(mapper []DynamicFuncsAndInterval) {
	DynamicMappersLock.Lock()
	defer DynamicMappersLock.Unlock()
	log.Println("设置动态配置mappers:", mapper)
	DynamicMappers = mapper
}

func GetDynamicMappers() []DynamicFuncsAndInterval {
	DynamicMappersLock.RLock()
	defer DynamicMappersLock.RUnlock()
	return DynamicMappers
}

func SetDynamicMetrics(metrics []string) {
	DynamicMetricsLock.Lock()
	defer DynamicMetricsLock.Unlock()
	DynamicMetrics = metrics
}

func GetDynamicMetrics() map[string]int {
	DynamicMetricsLock.RLock()
	defer DynamicMetricsLock.RUnlock()
	metrics := make(map[string]int)
	for _, v := range DynamicMetrics {
		metrics[v] = 1
	}
	return metrics
}

// 根据调用指令类型和是否容易被挂起而分类(通过不同的goroutine去执行,避免相互之间的影响)
func BuildMappers() {
	Mappers = []FuncsAndInterval{
		{
			//level 1 10s
			Fs: []func() []*model.MetricValue{
				CpuMetrics,
			},
			Interval: g.Config().MetricLevelIntervalConfig.Level1,
		},
		{
			//level 2 30s
			Fs: []func() []*model.MetricValue{
				UdpMetrics,
				DiskIOMetrics,
				IOStatsMetrics,
				SocketStatSummaryMetrics,
				NetstatMetrics,
				KernelMetrics,
				LoadAvgMetrics,
				MemMetrics,
				ProcMetrics,
				PortMetrics,
				NetMetrics,
			},
			Interval: g.Config().MetricLevelIntervalConfig.Level2,
		},
		{
			//level 3 60s
			Fs: []func() []*model.MetricValue{
				DeviceMetrics,
				DuMetrics,
				UrlMetrics,
				NtpMetrics,
				SysUptimeMetric,
			},
			Interval: g.Config().MetricLevelIntervalConfig.Level3,
		},
		{
			//level 4 120s
			Fs: []func() []*model.MetricValue{
				AgentMetrics,
				GpuMetrics,
			},
			Interval: g.Config().MetricLevelIntervalConfig.Level4,
		},
	}

	ProcMappers = []ProcFuncsAndInterval{
		{
			//level 2 30s
			Fs: []func(string, []string) []*model.MetricValue{
				ProcCpuMetrics,
				ProcMemMetrics,
				//ProcNetMetrics,
				//ProcSsMetrics,
			},
			Interval: g.Config().MetricLevelIntervalConfig.Level2,
		},
	}
}

//定时更新动态采集的mapper
func BuildDynamicMappers() {
	//初始化动态配置间隔时间的采集方法
	t := time.NewTicker(time.Second * time.Duration(3))
	defer t.Stop()
	for {
		<-t.C
		tempDynamicMappersJson, err := json.Marshal(g.GetDynamicMonitoringConfig())
		if err != nil {
			continue
		}
		if string(tempDynamicMappersJson) == DynamicMappersJson { //配置没有变更
			continue
		}
		log.Println("动态采集方法配置变更为：", string(tempDynamicMappersJson))
		DynamicMappersJson = string(tempDynamicMappersJson)

		tempDynamicMappers := []DynamicFuncsAndInterval{} //临时存放动态采集指标
		tempFunctions := make(map[string]int)             //临时存放metric指标
		tempMetrics := []string{}
		for metric, interval := range g.GetDynamicMonitoringConfig() {
			log.Println("metric:", metric, " interval:", interval)
			if interval < 0 {
				continue
			}
			tempMetrics = append(tempMetrics, metric)

			if functionName, okm := MetricsOfFunctionName[metric]; okm { //如果metric存在，寻找functionname，添加到待执行队列
				//已存在则看interval是否最小
				log.Println("functionName:", functionName)
				if _, ok := tempFunctions[functionName]; ok {
					if tempFunctions[functionName] > interval {
						tempFunctions[functionName] = interval
					}
				} else { //不存在则添加到待执行队列
					tempFunctions[functionName] = interval
				}

			}
		}

		if len(tempFunctions) <= 0 {
			log.Println("动态采集函数为空，返回！")
			continue
		}

		// 根据functionname 寻找对应的采集函数，添加到临时队列
		for functionName, interval := range tempFunctions {
			if fs, okf := FunctionNameOfFunction[functionName]; okf {
				item := DynamicFuncsAndInterval{}
				item.Fs = fs
				item.Interval = interval
				tempDynamicMappers = append(tempDynamicMappers, item)
			} else {
				continue
			}
		}

		if len(tempDynamicMappers) <= 0 {
			log.Println("动态采集指标为空，返回！")
			continue
		}
		log.Println("变更动态采集指标为:", tempDynamicMappers)
		log.Println("变更动态采集function为:", tempFunctions)
		log.Println("变更动态采集metrics为:", tempMetrics)
		//如果动态配置变更，则更新对应的mapper
		SetDynamicMappers(tempDynamicMappers)
		//如果动态配置变更，则更新采集的metrics
		SetDynamicMetrics(tempMetrics)
		//通知老的goroutine退出，启动新的采集线程
		DynamicMapperConfChangeSign <- 1

		log.Println("DynamicMappers:", GetDynamicMappers())
	}
}

func InitMetricsOfFunction() {
	log.Println("初始化metrics 对应的function")
	//初始化functionName 所对应的function
	FunctionNameOfFunction["CpuMetrics"] = CpuMetrics
	FunctionNameOfFunction["MemMetrics"] = MemMetrics
	FunctionNameOfFunction["NetMetrics"] = NetMetrics
	FunctionNameOfFunction["LoadAvgMetrics"] = LoadAvgMetrics
	FunctionNameOfFunction["DiskIOMetrics"] = DiskIOMetrics
	FunctionNameOfFunction["DuMetrics"] = DuMetrics
	FunctionNameOfFunction["Test1"] = Test1
	FunctionNameOfFunction["Test2"] = Test2
	//进程指标
	ProcFunctionNameOfFunction["CpuMetrics"] = ProcCpuMetrics
	ProcFunctionNameOfFunction["MemMetrics"] = ProcMemMetrics
	ProcFunctionNameOfFunction["IoMetrics"] = ProcIoMetrics

	//TEST指标
	MetricsOfFunctionName["test1.mytest1"] = "Test1"
	MetricsOfFunctionName["test1.mytest2"] = "Test1"
	MetricsOfFunctionName["test1.mytest3"] = "Test1"
	MetricsOfFunctionName["test2.mytest1"] = "Test2"
	MetricsOfFunctionName["test2.mytest2"] = "Test2"
	MetricsOfFunctionName["test2.mytest3"] = "Test2"
	//cpu指标
	MetricsOfFunctionName["cpu.idle"] = "CpuMetrics"
	MetricsOfFunctionName["cpu.busy"] = "CpuMetrics"
	MetricsOfFunctionName["cpu.user"] = "CpuMetrics"
	MetricsOfFunctionName["cpu.nice"] = "CpuMetrics"
	MetricsOfFunctionName["cpu.system"] = "CpuMetrics"
	MetricsOfFunctionName["cpu.iowait"] = "CpuMetrics"
	MetricsOfFunctionName["cpu.irq"] = "CpuMetrics"
	MetricsOfFunctionName["cpu.softirq"] = "CpuMetrics"
	MetricsOfFunctionName["cpu.steal"] = "CpuMetrics"
	MetricsOfFunctionName["cpu.guest"] = "CpuMetrics"
	MetricsOfFunctionName["cpu.switches"] = "CpuMetrics"
	//内存指标
	MetricsOfFunctionName["mem.memtotal"] = "MemMetrics"
	MetricsOfFunctionName["mem.memused"] = "MemMetrics"
	MetricsOfFunctionName["mem.memfree"] = "MemMetrics"
	MetricsOfFunctionName["mem.cached"] = "MemMetrics"
	MetricsOfFunctionName["mem.buffers"] = "MemMetrics"
	MetricsOfFunctionName["mem.swaptotal"] = "MemMetrics"
	MetricsOfFunctionName["mem.swapused"] = "MemMetrics"
	MetricsOfFunctionName["mem.swapfree"] = "MemMetrics"
	MetricsOfFunctionName["mem.memfree.percent"] = "MemMetrics"
	MetricsOfFunctionName["mem.memused.percent"] = "MemMetrics"
	MetricsOfFunctionName["mem.swapfree.percent"] = "MemMetrics"
	MetricsOfFunctionName["mem.swapused.percent"] = "MemMetrics"
	MetricsOfFunctionName["mem.shmem"] = "MemMetrics"
	MetricsOfFunctionName["mem.memavailable"] = "MemMetrics"
	//网卡指标
	MetricsOfFunctionName["net.if.in.bytes"] = "NetMetrics"
	MetricsOfFunctionName["net.if.in.packets"] = "NetMetrics"
	MetricsOfFunctionName["net.if.in.errors"] = "NetMetrics"
	MetricsOfFunctionName["net.if.in.dropped"] = "NetMetrics"
	MetricsOfFunctionName["net.if.in.fifo.errs"] = "NetMetrics"
	MetricsOfFunctionName["net.if.in.frame.errs"] = "NetMetrics"
	MetricsOfFunctionName["net.if.in.compressed"] = "NetMetrics"
	MetricsOfFunctionName["net.if.in.multicast"] = "NetMetrics"
	MetricsOfFunctionName["net.if.out.bytes"] = "NetMetrics"
	MetricsOfFunctionName["net.if.out.packets"] = "NetMetrics"
	MetricsOfFunctionName["net.if.out.errors"] = "NetMetrics"
	MetricsOfFunctionName["net.if.out.dropped"] = "NetMetrics"
	MetricsOfFunctionName["net.if.out.fifo.errs"] = "NetMetrics"
	MetricsOfFunctionName["net.if.out.collisions"] = "NetMetrics"
	MetricsOfFunctionName["net.if.out.carrier.errs"] = "NetMetrics"
	MetricsOfFunctionName["net.if.out.compressed"] = "NetMetrics"
	MetricsOfFunctionName["net.if.total.bytes"] = "NetMetrics"
	MetricsOfFunctionName["net.if.total.packets"] = "NetMetrics"
	MetricsOfFunctionName["net.if.total.errors"] = "NetMetrics"
	MetricsOfFunctionName["net.if.total.dropped"] = "NetMetrics"
	MetricsOfFunctionName["net.if.speed.bits"] = "NetMetrics"
	MetricsOfFunctionName["net.if.in.percent"] = "NetMetrics"
	MetricsOfFunctionName["net.if.out.percent"] = "NetMetrics"
	//负载指标
	MetricsOfFunctionName["load.1min"] = "LoadAvgMetrics"
	MetricsOfFunctionName["load.5min"] = "LoadAvgMetrics"
	MetricsOfFunctionName["load.15min"] = "LoadAvgMetrics"
	//磁盘指标
	MetricsOfFunctionName["disk.io.read_requests"] = "DiskIOMetrics"
	MetricsOfFunctionName["disk.io.read_merged"] = "DiskIOMetrics"
	MetricsOfFunctionName["disk.io.read_sectors"] = "DiskIOMetrics"
	MetricsOfFunctionName["disk.io.msec_read"] = "DiskIOMetrics"
	MetricsOfFunctionName["disk.io.write_requests"] = "DiskIOMetrics"
	MetricsOfFunctionName["disk.io.write_merged"] = "DiskIOMetrics"
	MetricsOfFunctionName["disk.io.write_sectors"] = "DiskIOMetrics"
	MetricsOfFunctionName["disk.io.msec_write"] = "DiskIOMetrics"
	MetricsOfFunctionName["disk.io.ios_in_progress"] = "DiskIOMetrics"
	MetricsOfFunctionName["disk.io.msec_total"] = "DiskIOMetrics"
	MetricsOfFunctionName["disk.io.msec_weighted_total"] = "DiskIOMetrics"
	//挂载点指标
	MetricsOfFunctionName["df.bytes.total"] = "DuMetrics"
	MetricsOfFunctionName["df.bytes.used"] = "DuMetrics"
	MetricsOfFunctionName["df.bytes.free"] = "DuMetrics"
	MetricsOfFunctionName["df.bytes.used.percent"] = "DuMetrics"
	MetricsOfFunctionName["df.bytes.free.percent"] = "DuMetrics"
	MetricsOfFunctionName["df.inodes.total"] = "DuMetrics"
	MetricsOfFunctionName["df.inodes.used"] = "DuMetrics"
	MetricsOfFunctionName["df.inodes.free"] = "DuMetrics"
	MetricsOfFunctionName["df.inodes.used.percent"] = "DuMetrics"
	MetricsOfFunctionName["df.inodes.free.percent"] = "DuMetrics"
	MetricsOfFunctionName["df.statistics.total"] = "DuMetrics"
	MetricsOfFunctionName["df.statistics.used"] = "DuMetrics"
	MetricsOfFunctionName["df.statistics.used.percent"] = "DuMetrics"
}
