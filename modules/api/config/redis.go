package config

import (
	"time"

	"github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
)

var RedisConnPool *redis.Pool

func InitRedisConnPool(redisaddr string, idle int) {

	log.Println("redisaddr aaaa", redisaddr, "idle", idle)
	RedisConnPool = &redis.Pool{
		MaxIdle:     idle,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisaddr)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: PingRedis,
	}
}

func PingRedis(c redis.Conn, t time.Time) error {
	_, err := c.Do("ping")
	if err != nil {
		log.Println("[ERROR] ping redis fail", err)
	}
	return err
}

func RedisSet(key, value string, ex int) error {
	rc := RedisConnPool.Get()
	defer rc.Close()
	if ex == 0 {
		_, err := rc.Do("SET", key, value)
		if err != nil {
			log.Error("Set redis", key, "fail:", err, "value:", value)
			return err
		}
	} else {
		_, err := rc.Do("SET", key, value, "EX", ex)
		if err != nil {
			log.Error("Set redis", key, "fail:", err, "value:", value)
			return err
		}
		log.Debug("set_redis_success", key, value, ex)
	}

	return nil
}

func RedisGet(key string) (string, error) {
	rc := RedisConnPool.Get()
	defer rc.Close()
	res, err := redis.String(rc.Do("Get", key))
	if err != nil {
		log.Error("Get redis", key, "fail:", err)
		return "", err
	}
	return res, nil
}
