package redi

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/chasex/redis-go-cluster"
	"github.com/open-falcon/falcon-plus/common/model"
)

const ExpressionsKey = "expressions"

//设置插件map到redis hash，数量级10，直接设置hash
func SetExpressions2Redis(m []*model.Expression) {
	var hsetValue []interface{}
	RedisKey := ExpressionsKey + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, RedisKey)
	for _, exp := range m {
		expByte, err := json.Marshal(exp)
		if err != nil {
			continue
		}
		hsetValue = append(hsetValue, exp.Id, string(expByte))
	}

	_, err := RedisCluster.Do("HMSET", hsetValue...)
	if err != nil {
		log.Println("hash set "+ExpressionsKey+" error:", err)
		return
	}
	_, err = RedisCluster.Do("EXPIRE", RedisKey, TIMEOUT) //设置30分超时
	if err != nil {
		log.Println("set "+ExpressionsKey+" EXPIRE error:", err)
		return
	}
	_, err = RedisCluster.Do("SET", ExpressionsKey, RedisKey)
	if err != nil {
		log.Println("set "+ExpressionsKey+" EXPIRE error:", err)
	}
}

//从redis中获取hash map
func GetExpressionsFromRedis() ([]*model.Expression, error) {
	var m []*model.Expression
	MapKey, err := redis.String(RedisCluster.Do("GET", ExpressionsKey))
	if err != nil {
		return m, err
	}

	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))
	if err != nil {
		return m, nil
	}
	for _, expStr := range res {
		exp := &model.Expression{}
		err = json.Unmarshal([]byte(expStr), exp)
		if err != nil {
			continue
		}
		m = append(m, exp)

	}
	return m, nil
}
