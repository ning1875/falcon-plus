package cron

import (
	"sync"

	"time"

	log "github.com/Sirupsen/logrus"
	redisc "github.com/chasex/redis-go-cluster"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/alarm/redi"
)

var BlockMonitorCounter sync.Map

func RefreshBlockMonitor() {
	for {
		go GetBlockMonitors()
		time.Sleep(1 * time.Minute)
		//time.Sleep(5 * time.Second)
	}
}

func GetBlockMonitors() {
	rc := redi.RedisCluster
	reply, err := redisc.Strings(rc.Do("SMEMBERS", g.BLOCK_MONITOR_SET))
	if err != nil {
		log.Errorf("GetBlockMonitors_SMEMBERS_%s_error:%+v", g.BLOCK_MONITOR_SET, err)
		return
	}

	if len(reply) == 0 {
		return
	}
	for _, item := range reply {
		log.Infof("item:%s", item)
		if replyIner, errIner := redisc.String(rc.Do("GET", item)); err != nil {
			log.Errorf("GetBlockMonitors_GET_%s_error:%+v", item, errIner)
			continue
		} else {
			//说明key已经过期-->报警屏蔽已经失效
			if replyIner == "" {
				BlockMonitorCounter.Delete(item)
				if _, errIner := rc.Do("SREM", g.BLOCK_MONITOR_SET, item); errIner != nil {
					log.Errorf("GetBlockMonitors_SREM_%s_error:%+v", item, errIner)
				}
			} else {
				BlockMonitorCounter.Store(item, replyIner)
			}
		}

	}
}
