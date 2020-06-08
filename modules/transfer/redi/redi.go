package redi

import (
	"log"
	"time"

	redisc "github.com/chasex/redis-go-cluster"
	"github.com/open-falcon/falcon-plus/modules/transfer/g"
)

var RedisCluster *redisc.Cluster

const TIMEOUT = 60 * 20

//初始化redis集群
func Init() {
	cluster, err := redisc.NewCluster(
		&redisc.Options{
			StartNodes:   g.Config().Redis.RedisClusterNodes,
			ConnTimeout:  300 * time.Millisecond,
			ReadTimeout:  300 * time.Millisecond,
			WriteTimeout: 300 * time.Millisecond,
			KeepAlive:    16,
			AliveTime:    60 * time.Second,
		})
	if err != nil {
		log.Fatalln("open master redis fail:", err)
	}
	RedisCluster = cluster
}
