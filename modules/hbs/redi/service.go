package redi

import (
	"log"
	"strconv"
	"time"

	"github.com/chasex/redis-go-cluster"
)

const HostServiceKey = "host_Service"

func SetHostService2Redis(m map[string]string) {
	var hsetValue []interface{}
	RedisKey := HostServiceKey + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, RedisKey)
	i := 0
	var err error
	for hname, snames := range m {
		i++
		hsetValue = append(hsetValue, hname, snames)
		if i == 1000 {
			_, err := RedisCluster.Do("HMSET", hsetValue...)
			if err != nil {
				break
			}
			hsetValue = hsetValue[0:1]
			i = 0
		}
	}
	if err == nil {
		if i > 0 {
			_, err := RedisCluster.Do("HMSET", hsetValue...)
			if err != nil {
				log.Println("hash set "+HostServiceKey+" error:", err)
				return
			}
		}
		_, err = RedisCluster.Do("EXPIRE", RedisKey, TIMEOUT) //设置30分超时
		if err != nil {
			log.Println("set "+HostServiceKey+" EXPIRE error:", err)
			return
		}
		_, err = RedisCluster.Do("SET", HostServiceKey, RedisKey)
		if err != nil {
			log.Println("set "+HostServiceKey+" EXPIRE error:", err)
		}
	} else {
		log.Println("hash set "+HostServiceKey+" error:", err)
	}
}

//从redis中获取hash map
func GetHostServiceFromRedis() (map[string]string, error) {
	m := make(map[string]string)
	MapKey, err := redis.String(RedisCluster.Do("GET", HostServiceKey))
	if err != nil {
		return m, err
	}
	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))
	if err != nil {
		return m, nil
	}
	for hname, stringsnames := range res {
		if _, ok := m[hname]; ok {
			m[hname] = m[hname] + "," + stringsnames
		} else {
			m[hname] = stringsnames
		}
	}
	return m, nil
}
