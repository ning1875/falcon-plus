package redi

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/chasex/redis-go-cluster"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
)

const (
	TemplateKey     = "template" //
	TPLKey          = "tpl"      //
	HostTemplateKey = "host_template"
)

//设置插件map到redis hash，数量级10，直接设置hash
func SetGroupTemplate2Redis(m map[int][]int) {
	var hsetValue []interface{}
	RedisKey := TemplateKey + `_` + g.HostName + `_` + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, RedisKey)
	i := 0
	var err error
	for gid, tids := range m {
		var stringTids []string
		for _, tid := range tids {
			stringTids = append(stringTids, strconv.Itoa(tid))
		}
		hsetValue = append(hsetValue, strconv.Itoa(gid), strings.Join(stringTids, "|||"))
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
				log.Println("hash set "+TemplateKey+" error:", err, " len :", i)
				return
			}
		}
		_, err = RedisCluster.Do("EXPIRE", RedisKey, TIMEOUT) //设置30分超时
		if err != nil {
			log.Println("set "+TemplateKey+" EXPIRE error:", err)
			return
		}
		_, err = RedisCluster.Do("SET", TemplateKey, RedisKey)
		if err != nil {
			log.Println("set "+TemplateKey+" EXPIRE error:", err)
		}
	} else {
		if err != nil {
			log.Println("hash set "+TemplateKey+" error:", err, " len :", len(m))
			return
		}
	}
}

//从redis中获取hash map
func GetGroupTemplateFromRedis() (map[int][]int, error) {
	m := make(map[int][]int)
	MapKey, err := redis.String(RedisCluster.Do("GET", TemplateKey))
	if err != nil {
		return m, err
	}
	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))

	if err != nil {
		return m, nil
	}
	for StringGid, StringTids := range res {
		gid, err := strconv.Atoi(StringGid)
		if err != nil {
			continue
		}
		Tids := strings.Split(StringTids, "|||")
		for _, tid := range Tids {
			IntTid, err := strconv.Atoi(tid)
			if err != nil {
				continue
			}
			m[gid] = append(m[gid], IntTid)
		}
	}
	return m, nil
}

func SetTPL2Redis(m map[int]*model.Template) {
	var hsetValue []interface{}
	RedisKey := TPLKey + `_` + g.HostName + `_` + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, RedisKey)
	i := 0
	var err error
	for id, tpl := range m {
		i++
		tplByte, err := json.Marshal(tpl)
		if err != nil {
			continue
		}
		hsetValue = append(hsetValue, strconv.Itoa(id), string(tplByte))
		if i == 500 {
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
				log.Println("hash set "+TPLKey+" error:", err)
				return
			}
		}
		_, err = RedisCluster.Do("EXPIRE", RedisKey, TIMEOUT) //设置30分超时
		if err != nil {
			log.Println("set "+TPLKey+" EXPIRE error:", err)
			return
		}
		_, err = RedisCluster.Do("SET", TPLKey, RedisKey)
		if err != nil {
			log.Println("set "+TPLKey+" EXPIRE error:", err)
		}
	}
}

//从redis中获取hash map
func GetTPLFromRedis() (map[int]*model.Template, error) {
	m := make(map[int]*model.Template)
	MapKey, err := redis.String(RedisCluster.Do("GET", TPLKey))
	if err != nil {
		return m, err
	}
	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))
	if err != nil {
		return m, nil
	}
	for StrId, StringTPL := range res {
		tpl := &model.Template{}
		err := json.Unmarshal([]byte(StringTPL), tpl)
		if err != nil {
			continue
		}
		id, err := strconv.Atoi(StrId)
		if err != nil {
			continue
		}
		m[id] = tpl
	}
	return m, nil
}

func SetHostTemplate2Redis(m map[int][]int) {
	var hsetValue []interface{}
	RedisKey := HostTemplateKey + `_` + g.HostName + `_` + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, RedisKey)
	i := 0
	var err error
	for tid, hids := range m {
		i++
		var stringhids []string
		for _, hid := range hids {
			stringhids = append(stringhids, strconv.Itoa(hid))
		}
		hsetValue = append(hsetValue, strconv.Itoa(tid), strings.Join(stringhids, SEP))
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
				log.Println("hash set "+HostTemplateKey+" error:", err)
				return
			}
		}
		_, err = RedisCluster.Do("EXPIRE", RedisKey, TIMEOUT) //设置30分超时
		if err != nil {
			log.Println("set "+HostTemplateKey+" EXPIRE error:", err)
			return
		}
		_, err = RedisCluster.Do("SET", HostTemplateKey, RedisKey)
		if err != nil {
			log.Println("set "+HostTemplateKey+" EXPIRE error:", err)
		}
	} else {
		log.Println("hash set "+HostTemplateKey+" error:", err)
	}
}

//从redis中获取hash map
func GetHostTemplateFromRedis() (map[int][]int, error) {
	m := make(map[int][]int)
	MapKey, err := redis.String(RedisCluster.Do("GET", HostTemplateKey))
	if err != nil {
		return m, err
	}
	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))
	if err != nil {
		return m, nil
	}
	for StringTid, StringHid := range res {
		Tid, err := strconv.Atoi(StringTid)
		if err != nil {
			continue
		}
		Hids := strings.Split(StringHid, SEP)
		for _, hid := range Hids {
			IntHid, err := strconv.Atoi(hid)
			if err != nil {
				continue
			}
			m[Tid] = append(m[Tid], IntHid)
		}
	}
	return m, nil
}
