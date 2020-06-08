package redi

import (
	"log"
	"strconv"
	"time"

	"github.com/chasex/redis-go-cluster"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
)

const (
	HostKey          = "host"           //
	MonitoredHostKey = "monitored_host" //
)

//设置插件map到redis hash，数量级10，直接设置hash
func SetHost2Redis(m map[string]int) {
	log.Println("SetHost2Redis")
	var hsetValue []interface{}
	RedisKey := HostKey + `_` + g.HostName + `_` + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, RedisKey)
	i := 0
	var err error
	for host, id := range m {
		i++
		hsetValue = append(hsetValue, host, strconv.Itoa(id))
		if i == 500 {
			_, err = RedisCluster.Do("HMSET", hsetValue...)
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
				log.Println("hash set "+HostKey+" error:", err)
				return
			}
		}
		_, err = RedisCluster.Do("EXPIRE", RedisKey, TIMEOUT) //设置30分超时
		if err != nil {
			log.Println("set "+HostKey+" EXPIRE error:", err)
			return
		}
		_, err = RedisCluster.Do("SET", HostKey, RedisKey)
		if err != nil {
			log.Println("set "+HostKey+"  error:", err)
		}
	} else {
		log.Println("hash set "+HostKey+" error:", err)
	}
}

//从redis中获取hash map
func GetHostFromRedis() (map[string]int, error) {
	m := make(map[string]int)
	MapKey, err := redis.String(RedisCluster.Do("GET", HostKey))
	if err != nil {
		return m, err
	}
	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))
	if err != nil {
		return m, nil
	}
	for host, StringId := range res {
		id, err := strconv.Atoi(StringId)
		if err != nil {
			continue
		}
		m[host] = id
	}
	return m, nil
}

//设置插件map到redis hash，数量级10，直接设置hash
func SetMonitoredHost2Redis(m map[int]*model.Host) {
	var hsetValue []interface{}
	RedisKey := MonitoredHostKey + `_` + g.HostName + `_` + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, RedisKey)
	i := 0
	var err error
	for id, host := range m {
		i++
		//存入数据更新时间
		hsetValue = append(hsetValue, id, host.Name)
		if i == 500 {
			_, err = RedisCluster.Do("HMSET", hsetValue...)
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
				log.Println("hash set "+MonitoredHostKey+" error:", err)
				return
			}
		}
		_, err = RedisCluster.Do("EXPIRE", RedisKey, TIMEOUT) //设置30分超时
		if err != nil {
			log.Println("set "+MonitoredHostKey+" EXPIRE error:", err)
			return
		}
		_, err = RedisCluster.Do("SET", MonitoredHostKey, RedisKey)
		if err != nil {
			log.Println("set "+MonitoredHostKey+" EXPIRE error:", err)
		}
	} else {
		log.Println("hash set "+MonitoredHostKey+" error:", err, " len ", i)
	}
}

//从redis中获取hash map
func GetMonitoredHostFromRedis() (map[int]*model.Host, error) {
	m := make(map[int]*model.Host)
	MapKey, err := redis.String(RedisCluster.Do("GET", MonitoredHostKey))
	if err != nil {
		return m, err
	}
	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))
	if err != nil {
		return m, nil
	}
	for idStr, Name := range res {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		m[id] = &model.Host{
			Id:   id,
			Name: Name,
		}

	}
	return m, nil
}
