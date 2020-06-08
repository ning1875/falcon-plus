package rediscluster

import (
	"strconv"
	"time"

	redisc "github.com/chasex/redis-go-cluster"
)

func GetDisLock(rc *redisc.Cluster, lockkey string, timeout int64) bool {
	expireAt := strconv.FormatInt(time.Now().Unix()+timeout, 10)

	setNxres, err := redisc.Int64(rc.Do("SETNX", lockkey, expireAt))
	if err != nil {
		return false
	}
	//设置成功获取锁
	if setNxres == int64(1) {
		// 设置锁过期时间，防止死锁
		rc.Do("EXPIRE", lockkey, timeout)
		return true
	}

	//抢锁失败，判断过期时间
	if setNxres == int64(0) {
		getRes, err := redisc.Int64(rc.Do("GET", lockkey))
		if err != nil {
			return false
		}
		// 设置过期时间可能失败，由时间值判断是否超时
		if time.Now().Unix()-getRes > timeout {
			rc.Do("EXPIRE", lockkey, timeout)
			return true
		}
	}
	return false
}
