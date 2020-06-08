// liangyuntao
// 获取动态监控项

package db

import (
	"log"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/redi"
)

func queryDynamicConfig() (map[string][]*model.DynamicConfig, error) {
	dynamicConfig := make(map[string][]*model.DynamicConfig)
	sql := "select id, endpoint,metrics,tag,intervals from dynamic_config where status = 1"
	rows, err := DB.Query(sql)
	if err != nil {
		log.Println("ERROR:", err)
		return dynamicConfig, err
	}

	defer rows.Close()
	for rows.Next() {
		t := model.DynamicConfig{}
		err = rows.Scan(&t.Id, &t.EndPoint, &t.Metrics, &t.Tag, &t.Interval)
		if err != nil {
			log.Println("WARN:", err)
			continue
		}
		dynamicConfig[t.EndPoint] = append(dynamicConfig[t.EndPoint], &t)
	}

	return dynamicConfig, nil
}

func QueryDynamicConfig() (map[string][]*model.DynamicConfig, error) {
	m := make(map[string][]*model.DynamicConfig)
	if redi.GetLock(redi.DynamicConfigLockKey, redi.RedisDisTLockTimeOut) {
		dbm, err := queryDynamicConfig()
		if err != nil {
			return m, err
		}
		redi.SetDynamicConfig2Redis(dbm)
		return dbm, nil
	} else { //没获取到锁 从redis中获取
		redism, err := redi.GetDynamicConfigRedis()
		//读取成功返回
		if err == nil {
			return redism, nil
		} else {
			//读取失败，从db中获取
			return queryDynamicConfig()
		}
	}
}
