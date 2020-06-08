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

package rpc

import (
	"bytes"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	redis "github.com/chasex/redis-go-cluster"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/common/utils"
	"github.com/open-falcon/falcon-plus/modules/hbs/cache"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
	"github.com/open-falcon/falcon-plus/modules/hbs/redi"
)

func (t *Agent) MinePlugins(args model.AgentHeartbeatRequest, reply *model.AgentPluginsResponse) error {
	if args.Hostname == "" {
		return nil
	}

	//reply.Plugins = cache.GetPlugins(args.Hostname)
	reply.Plugins = cache.JudgeHostPlugin(args.Hostname)
	//log.Printf("host:%v got_plugin:%v", args.Hostname, reply.Plugins)
	reply.Timestamp = time.Now().Unix()

	return nil
}

func (t *Agent) ReportStatus(args *model.AgentReportRequest, reply *model.AgentUpgradeRpcResponse) error {
	if args.Hostname == "" {
		reply.Code = 1
		return nil
	}
	cache.NowAgentVersionMap.Store(args.Hostname, args.AgentVersion)
	go cache.Agents.Put(args)
	reply.Args = nil
	//判断是否需要agent自更新
	//Version为空说明server还没有收到升级命令,或者此次升级已完成,或者升级任务被取消
	if cache.NewAgentUpgradeArgs.Version == "" {
		return nil
	}
	if args.AgentVersion == cache.NewAgentUpgradeArgs.Version {
		log.Printf("Host:%+v upgrade finished,delete from cache", args.Hostname)
		if _, rediError := redi.RedisCluster.Do("SREM", g.REDISAGENTUPGRADESET, args.Hostname); rediError != nil {
			log.Errorf("ReportStatus_DO_REDIS_SREM_FAILED:%+v,end:%s", rediError, args.Hostname)
		}
		return nil
	} else {
		//agentVersions := cache.UpgradeAgent{LastVersion: args.AgentVersion,
		//	ThisVersion: cache.NewAgentUpgradeArgs.Version,
		//	Timestamp:   time.Now().Unix(),
		//}
		//检查并发更新队列就是所有正在更新的agent的keys
		inUpgradeNum := -1
		if args.InUpgrading == true {
			log.Warnf("Host:+%v still in upgrade  do not send args again", args.Hostname)
			return nil
		}
		var rediError error
		if inUpgradeNum, rediError = redis.Int(redi.RedisCluster.Do("SCARD", g.REDISAGENTUPGRADESET)); rediError != nil {
			log.Errorf("ReportStatus_DO_REDIS_SCARD_FAILED:%+v,end:%s", rediError, args.Hostname)
		}
		if inUpgradeNum >= 0 && inUpgradeNum < g.Config().AgentUpdateQueueLength {
			log.Infof("Host:%+v need to  upgrade,send args", args.Hostname)
			_, rediError = redis.Int(redi.RedisCluster.Do("SADD", g.REDISAGENTUPGRADESET, args.Hostname))
			if rediError != nil {
				log.Errorf("ReportStatus_DO_REDIS_SADD_FAILED:%+v,end:%s", rediError, args.Hostname)
			} else {
				redis.Int(redi.RedisCluster.Do("EXPIRE", g.REDISAGENTUPGRADESET, g.Config().RedisAgentUpgradeSetExpire))
				reply.Args = cache.NewAgentUpgradeArgs
				return nil
			}

		}
		//超过了升级的并发
	}

	return nil
}

// 需要checksum一下来减少网络开销？其实白名单通常只会有一个或者没有，无需checksum
func (t *Agent) TrustableIps(args *model.NullRpcRequest, ips *string) error {
	*ips = strings.Join(g.Config().Trustable, ",")
	return nil
}

func (t *Agent) BuiltinMetrics(args *model.AgentHeartbeatRequest, reply *model.BuiltinMetricResponse) error {
	if args.Hostname == "" {
		return nil
	}
	metrics := cache.JudgeHostSpecialMetric(args.Hostname)
	//g.Logger.Infof("metrics:%+v", metrics)
	checksum := ""
	if len(metrics) > 0 {
		checksum = DigestBuiltinMetrics(metrics)
	}

	if args.Checksum == checksum {
		reply.Metrics = []*model.BuiltinMetric{}
	} else {
		reply.Metrics = metrics
	}
	reply.Checksum = checksum
	reply.Timestamp = time.Now().Unix()

	return nil
}

// agent按照server端的配置，按需采集的metric，比如net.port.listen port=22 或者 proc.num name=zabbix_agentd
//func (t *Agent) BuiltinMetrics(args *model.AgentHeartbeatRequest, reply *model.BuiltinMetricResponse) error {
//	if args.Hostname == "" {
//		return nil
//	}
//
//	metrics, err := cache.GetBuiltinMetrics(args.Hostname)
//	//log.Printf("cache.GetBuiltinMetrics for host %s metircs %v err: %v", args.Hostname, metrics, err)
//	if err != nil {
//		return nil
//	}
//
//	checksum := ""
//	if len(metrics) > 0 {
//		checksum = DigestBuiltinMetrics(metrics)
//	}
//
//	if args.Checksum == checksum {
//		reply.Metrics = []*model.BuiltinMetric{}
//	} else {
//		reply.Metrics = metrics
//	}
//	reply.Checksum = checksum
//	reply.Timestamp = time.Now().Unix()
//
//	return nil
//}

func DigestBuiltinMetrics(items []*model.BuiltinMetric) string {
	sort.Sort(model.BuiltinMetricSlice(items))

	var buf bytes.Buffer
	for _, m := range items {
		buf.WriteString(m.String())
	}

	return utils.Md5(buf.String())
}

func (t *Agent) DynamicMonitoring(args *model.AgentReportRequest, reply *model.AgentDynamicMonitoringConfigRpcResponse) error {
	if args.Hostname == "" {
		reply.Code = 1
		return nil
	}

	cfgs := make(map[string]int)
	if _, ok := cache.DynamicConfig.Get()[args.Hostname]; ok {
		for _, cfg := range cache.DynamicConfig.Get()[args.Hostname] {
			cfgs[cfg.Metrics] = cfg.Interval

		}
	}
	reply.Code = 0

	reply.Args = cfgs
	return nil
}

//send services to agent
func (t *Agent) SyncServiceNames(args *model.AgentReportRequest, reply *model.AgentServiceConfigRpcResponse) error {
	if args.Hostname == "" {
		reply.Code = 1
		return nil
	}

	svns, exists := cache.Services.Get()[args.Hostname]
	if !exists || svns == "" {
		reply.Code = 2
		return nil
	}

	reply.Code = 0
	reply.Args = svns
	return nil
}
