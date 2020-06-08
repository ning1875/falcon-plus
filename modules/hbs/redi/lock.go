package redi

import (
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/chasex/redis-go-cluster"
)

const LockKey = "lock"

const (
	GroupPluginsLockKey    = "GroupPluginsLock"
	GroupTemplatesLockKey  = "GroupTemplatesLock"
	HostGroupsMapLockKey   = "HostGroupsMapLock"
	HostMapLockKey         = "HostMapLock"
	TemplateCacheLockKey   = "TemplateCacheLock"
	StrategiesLockKey      = "StrategiesLock"
	HostTemplateIdsLockKey = "HostTemplateIdsLock"
	ExpressionCacheLockKey = "ExpressionCacheLock"
	MonitoredHostsLockKey  = "MonitoredHostsLock"
	DynamicConfigLockKey   = "DynamicConfigLock"
	HostServiceIdsLockKey  = "HostServiceIdsLockKey"

	RedisDisTLockTimeOut = 55

	GroupPluginsWLockKey    = "GroupPluginsWLock"
	GroupTemplatesWLockKey  = "GroupTemplatesWLock"
	HostGroupsMapWLockKey   = "HostGroupsMapWLock"
	HostMapWLockKey         = "HostMapWLock"
	TemplateCacheWLockKey   = "TemplateCacheWLock"
	StrategiesWLockKey      = "StrategiesWLock"
	HostTemplateIdsWLockKey = "HostTemplateIdsWLock"
	ExpressionCacheWLockKey = "ExpressionCacheWLock"
	MonitoredHostsWLockKey  = "MonitoredHostsWLock"
	DynamicConfigWLockKey   = "DynamicConfigWLock"
	HostServiceIdsWLockKey  = "HostServiceIdsLockKey"
)

//获取redis分布式锁
func GetLock(LockKey string, Timeout int64) bool {
	expireAt := strconv.FormatInt(time.Now().Unix()+Timeout, 10) //设置锁的值为过期时间
	setRes, err := redis.Int64(RedisCluster.Do("SETNX", LockKey, expireAt))
	if err != nil {
		return false
	}
	//设置成功则获取锁
	if setRes == int64(1) {
		//锁过期时间为60s，防止死锁
		RedisCluster.Do("EXPIRE", LockKey, Timeout)
		return true
	}
	//没有设置成功，则判断值的过期时间
	if setRes == int64(0) {
		getRes, err := redis.Int64(RedisCluster.Do("GET", LockKey))
		if err != nil {
			return false
		}
		//设置过期时间可能失败，由值再次判断是否锁过期
		if time.Now().Unix()-getRes > Timeout {
			RedisCluster.Do("EXPIRE", LockKey, Timeout)
			return true
		}
	}
	return false
}

//解锁
func WUnlock(LockKey string) {
	RedisCluster.Do("DEL", LockKey)
}

//加写锁
func WLock(LockKey string, Timeout int64) error {
	for {
		expireAt := strconv.FormatInt(time.Now().Unix()+Timeout, 10)
		res, err := redis.Int(RedisCluster.Do("SETNX", LockKey, expireAt))
		if err != nil {
			return errors.New("redis error" + err.Error())
		}
		if res == 1 {
			RedisCluster.Do("EXPIRE", LockKey, Timeout)
			return nil
		} else {
			//log.Println("wait lock:", LockKey)
			time.Sleep(time.Millisecond * 30)
			//getRes, err := redis.Int64(RedisCluster.Do("GET", LockKey))
			//if err != nil {
			//	time.Sleep(time.Millisecond * 10)
			//}
			////锁超时，重新设置，视为获取到锁
			//if time.Now().Unix()-getRes > Timeout {
			//	RedisCluster.Do("EXPIRE", LockKey, Timeout)
			//	return nil
			//}else{
			//	time.Sleep(time.Millisecond * 10)
			//}
		}
	}
}
