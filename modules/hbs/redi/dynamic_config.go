package redi

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/chasex/redis-go-cluster"
	"github.com/open-falcon/falcon-plus/common/model"
)

const DynamicConfigKey = "dynamic_config" //

//设置插件map到redis hash，数量级10，直接设置hash
func SetDynamicConfig2Redis(m map[string][]*model.DynamicConfig) {
	var hsetValue []interface{}
	RedisKey := DynamicConfigKey + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, RedisKey)
	for endpoint, cfgs := range m {
		var cfgArr []string
		for _, cfg := range cfgs {
			cfgByte, err := json.Marshal(cfg)

			if err != nil {
				continue
			}
			cfgArr = append(cfgArr, string(cfgByte))
		}
		hsetValue = append(hsetValue, endpoint, strings.Join(cfgArr, SEP))
	}

	_, err := RedisCluster.Do("HMSET", hsetValue...)
	if err != nil {
		log.Println("hash set "+DynamicConfigKey+" error:", err)
		return
	}
	_, err = RedisCluster.Do("EXPIRE", RedisKey, TIMEOUT) //设置30分超时
	if err != nil {
		log.Println("set "+DynamicConfigKey+" EXPIRE error:", err)
		return
	}
	_, err = RedisCluster.Do("SET", DynamicConfigKey, RedisKey)
	if err != nil {
		log.Println("set "+DynamicConfigKey+" EXPIRE error:", err)
	}
}

//从redis中获取hash map
func GetDynamicConfigRedis() (map[string][]*model.DynamicConfig, error) {
	m := make(map[string][]*model.DynamicConfig)
	MapKey, err := redis.String(RedisCluster.Do("GET", DynamicConfigKey))
	if err != nil {
		return m, err
	}

	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))
	if err != nil {
		return m, err
	}
	for endpoint, cfgsStr := range res {
		cfgsArr := strings.Split(cfgsStr, SEP)
		for _, cfgStr := range cfgsArr {
			cfg := &model.DynamicConfig{}
			err := json.Unmarshal([]byte(cfgStr), cfg)
			if err != nil {
				continue
			}
			m[endpoint] = append(m[endpoint], cfg)
		}
	}

	return m, nil
}
