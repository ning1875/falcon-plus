package redi

import (
	"log"
	"strconv"
	"strings"
	"time"

	redis "github.com/chasex/redis-go-cluster"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
)

const GroupKeyPerfix = "group" //

//设置插件map到redis hash，数量级10，直接设置hash
func SetHostGroup2Redis(m map[int][]int) {
	var hsetValue []interface{}
	RedisKey := GroupKeyPerfix + `_` + g.HostName + `_` + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, RedisKey)
	i := 0
	var err error
	for gid, hids := range m {
		i++
		var stringHids []string
		for _, hid := range hids {
			stringHids = append(stringHids, strconv.Itoa(hid))
		}
		hsetValue = append(hsetValue, strconv.Itoa(gid), strings.Join(stringHids, SEP))
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
				log.Println("hash set "+GroupKeyPerfix+" error:", err)
				return
			}
		}
		_, err = RedisCluster.Do("EXPIRE", RedisKey, TIMEOUT) //设置30分超时
		if err != nil {
			log.Println("set "+GroupKeyPerfix+" EXPIRE error:", err)
			return
		}
		_, err = RedisCluster.Do("SET", GroupKeyPerfix, RedisKey)
		if err != nil {
			log.Println("set "+GroupKeyPerfix+" EXPIRE error:", err)
		}
	} else {
		log.Println("hash set "+GroupKeyPerfix+" error:", err)
	}
}

//从redis中获取hash map
func GetHostGroupFromRedis() (map[int][]int, error) {
	m := make(map[int][]int)
	MapKey, err := redis.String(RedisCluster.Do("GET", GroupKeyPerfix))
	if err != nil {
		return m, err
	}

	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))
	if err != nil {
		return m, nil
	}
	for StringGid, StringHids := range res {
		gid, err := strconv.Atoi(StringGid)
		if err != nil {
			continue
		}
		Hids := strings.Split(StringHids, SEP)
		for _, hid := range Hids {
			IntHid, err := strconv.Atoi(hid)
			if err != nil {
				continue
			}
			m[gid] = append(m[gid], IntHid)
		}
	}
	return m, nil
}
