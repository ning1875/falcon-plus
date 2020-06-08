package redi

import (
	"log"
	"time"

	redis "github.com/chasex/redis-go-cluster"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
)

var RedisCluster *redis.Cluster

const SEP = "|||"
const TIMEOUT = 60 * 20

//初始化redis集群
func Init() {
	cluster, err := redis.NewCluster(
		&redis.Options{
			//StartNodes:   []string{"10.14.68.200:7000", "10.14.68.201:7000", "10.14.68.202:7000", "10.14.68.203:7000", "10.14.68.204:7000"},
			StartNodes:   g.Config().RedisClusterNodes,
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
