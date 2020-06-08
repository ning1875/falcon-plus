package g

import (
	"log"
	"time"

	redisc "github.com/chasex/redis-go-cluster"
)

var RedisCluster *redisc.Cluster

const SEP = "|||"
const TIMEOUT = 60 * 20

//初始化redis集群
func InitRedisCluster() {
	cluster, err := redisc.NewCluster(
		&redisc.Options{
			//StartNodes:   []string{"10.14.68.200:7000", "10.14.68.201:7000", "10.14.68.202:7000", "10.14.68.203:7000", "10.14.68.204:7000"},
			StartNodes:   Config().Alarm.Redis.RedisClusterNodes,
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
