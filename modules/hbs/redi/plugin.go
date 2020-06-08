package redi

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/chasex/redis-go-cluster"
)

const PluginKey = "plugin" //

//设置插件map到redis hash，数量级10，直接设置hash
func SetPlugins2Redis(m map[int][]string) {
	var hsetValue []interface{}
	RedisKey := PluginKey + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, RedisKey)
	for grpid, dir := range m {
		hsetValue = append(hsetValue, strconv.Itoa(grpid), strings.Join(dir, SEP))
	}

	_, err := RedisCluster.Do("HMSET", hsetValue...)
	if err != nil {
		log.Println("hash set "+PluginKey+" error:", err)
		return
	}
	_, err = RedisCluster.Do("EXPIRE", RedisKey, TIMEOUT) //设置30分超时
	if err != nil {
		log.Println("set "+PluginKey+" EXPIRE error:", err)
		return
	}
	_, err = RedisCluster.Do("SET", PluginKey, RedisKey)
	if err != nil {
		log.Println("set "+PluginKey+" EXPIRE error:", err)
	}

}

//从redis中获取hash map
func GetPluginsFromRedis() (map[int][]string, error) {
	m := make(map[int][]string)
	MapKey, err := redis.String(RedisCluster.Do("GET", PluginKey))
	if err != nil {
		return m, err
	}
	res, err := redis.StringMap(RedisCluster.Do("HGETALL", MapKey))

	if err != nil {
		return m, nil
	}
	for grpid, dir := range res {
		gid, err := strconv.Atoi(grpid)
		if err != nil {
			continue
		}
		m[gid] = strings.Split(dir, SEP)
	}
	return m, nil
}
