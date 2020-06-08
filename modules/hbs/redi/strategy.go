package redi

import (
	"encoding/json"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/chasex/redis-go-cluster"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
)

const StrategiesKey = "strategies" //
const UnionStrategiesKey = "union_strategies"

//设置插件map到redis hash，数量级10，直接设置hash
func SetStrategies2Redis(m map[int]*model.Strategy) {
	var hsetValue []interface{}
	RedisKey := StrategiesKey + `_` + g.HostName + `_` + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, RedisKey)
	i := 0
	var err error
	for id, strategy := range m {
		i++
		strategyByte, err := json.Marshal(strategy)
		if err != nil {
			continue
		}
		hsetValue = append(hsetValue, id, string(strategyByte))
		if i >= 500 {
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
				log.Println("hash set "+StrategiesKey+" error:", err)
				return
			}
		}
		_, err = RedisCluster.Do("EXPIRE", RedisKey, TIMEOUT) //设置30分超时
		if err != nil {
			log.Println("set "+StrategiesKey+" EXPIRE error:", err)
			return
		}
		_, err = RedisCluster.Do("SET", StrategiesKey, RedisKey)
		if err != nil {
			log.Println("set "+StrategiesKey+" EXPIRE error:", err)
		}
	} else {
		log.Println("hash set "+StrategiesKey+" error:", err)
	}
}

//从redis中获取hash map
func GetStrategiesFromRedis() (map[int]*model.Strategy, error) {
	m := make(map[int]*model.Strategy)
	MapKey, err := redis.String(RedisCluster.Do("GET", StrategiesKey))
	if err != nil {
		return m, err
	}

	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))
	if err != nil {
		return m, nil
	}
	for idStr, strategyStr := range res {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		strategy := &model.Strategy{}
		err = json.Unmarshal([]byte(strategyStr), strategy)
		if err != nil {
			continue
		}
		m[id] = strategy
	}
	return m, nil
}

func SetUnionStrategies2Redis(m map[int][]*model.Strategy) {
	var hsetValue []interface{}
	redisKey := UnionStrategiesKey + `_` + g.HostName + `_` + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, redisKey)
	i := 0
	var err error
	for id, strategy := range m {
		i++
		strategyByte, err := json.Marshal(strategy)
		if err != nil {
			continue
		}
		hsetValue = append(hsetValue, id, string(strategyByte))
		if i >= 500 {
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
				log.Errorf("hash set "+UnionStrategiesKey+" error:", err)
				return
			}
		}
		_, err = RedisCluster.Do("EXPIRE", redisKey, TIMEOUT) //设置30分超时
		if err != nil {
			log.Errorf("set "+UnionStrategiesKey+" EXPIRE error:", err)
			return
		}
		_, err = RedisCluster.Do("SET", UnionStrategiesKey, redisKey)
		if err != nil {
			log.Errorf("set "+UnionStrategiesKey+" EXPIRE error:", err)
		}
	} else {
		log.Errorf("hash set "+UnionStrategiesKey+" error:", err)
	}
}

func GetUnionStrategiesFromRedis() (map[int][]*model.Strategy, error) {
	m := make(map[int][]*model.Strategy)
	MapKey, err := redis.String(RedisCluster.Do("GET", UnionStrategiesKey))
	log.Debugf("GetUnionStrategiesFromRedis_%s", MapKey)
	if err != nil {
		return m, err
	}

	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))
	if err != nil {
		return m, nil
	}
	for idStr, strategyStr := range res {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		obj := []json.RawMessage{}
		err = json.Unmarshal([]byte(strategyStr), &obj)
		if err != nil {
			continue
		}

		var sts []*model.Strategy
		for _, o := range obj {
			var st *model.Strategy
			if err = json.Unmarshal(o, &st); err == nil {
				sts = append(sts, st)
			}
		}
		m[id] = sts
	}
	return m, nil
}
