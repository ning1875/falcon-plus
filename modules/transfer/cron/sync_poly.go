package cron

import (
	"sync"

	"time"

	"github.com/chasex/redis-go-cluster"
	"github.com/open-falcon/falcon-plus/modules/transfer/redi"
)

type SafeEndPolyCache struct {
	Map map[string]string
	sync.RWMutex
}

const EndPolyStrategyHashKey = "end_poly_strategy_hash_key"

var PolyCounterCache = sync.Map{}
var EndPolyCache = &SafeEndPolyCache{}

func RunSync() {
	for {
		EndPolyCache.SyncEndPolyFromRedis()
		//log.Infof("RunSync:%+v", EndPolyCache)
		time.Sleep(60 * time.Second)
	}
}

func (this *SafeEndPolyCache) SyncEndPolyFromRedis() error {
	MapKey, err := redis.String(redi.RedisCluster.Do("GET", EndPolyStrategyHashKey))
	if err != nil {
		return err
	}
	res, err := redis.StringMap(redi.RedisCluster.Do("HGETALL", MapKey))

	if err != nil {
		return nil
	}
	this.Lock()
	defer this.Unlock()
	this.Map = res
	return nil
}

func (this *SafeEndPolyCache) Get(end string) string {
	this.RLock()
	defer this.RUnlock()
	return this.Map[end]
}
