// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"fmt"
	"log"
	"time"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
	"github.com/open-falcon/falcon-plus/modules/hbs/redi"
)

func queryHosts() (map[string]int, error) {
	m := make(map[string]int)

	sql := "select id, hostname from host"
	rows, err := DB.Query(sql)
	if err != nil {
		log.Println("ERROR:", err)
		return m, err
	}

	defer rows.Close()
	for rows.Next() {
		var (
			id       int
			hostname string
		)

		err = rows.Scan(&id, &hostname)
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}

		m[hostname] = id
	}

	return m, nil
}
func QueryHosts() (map[string]int, error) {
	log.Println("QueryHosts")
	m := make(map[string]int)
	if redi.GetLock(redi.HostMapLockKey, redi.RedisDisTLockTimeOut) {
		dbm, err := queryHosts()
		if err != nil {
			return m, err
		}
		redi.SetHost2Redis(dbm)
		return dbm, nil
	} else { //没获取到锁 从redis中获取
		begin := time.Now().UnixNano()
		redism, err := redi.GetHostFromRedis()
		log.Println("query redis cost:", (time.Now().UnixNano()-begin)/1000)
		//读取成功返回
		if err == nil {
			return redism, nil
		} else {
			//读取失败，从db中获取
			return queryHosts()
		}
	}
}

func QueryHostMaintain(hostname string) bool {

	rawSql := fmt.Sprintf(`select id, hostname from host where maintain_begin < (select unix_timestamp()) and maintain_end > (select unix_timestamp()) and hostname='%s';`,
		hostname)

	//g.Logger.Infof("rawSql", rawSql)
	rows, _ := DBRO.Query(rawSql)
	defer rows.Close()
	for rows.Next() {
		t := model.Host{}
		err := rows.Scan(&t.Id, &t.Name)
		if err != nil {
			g.Logger.Errorf("WARN:", err)
			continue
		}
		if t.Name == hostname {
			//处于维护中
			return true
		} else {
			return false
		}

	}
	return false
}

func queryMonitoredHosts() (map[int]*model.Host, error) {
	hosts := make(map[int]*model.Host)
	now := time.Now().Unix()
	sql := fmt.Sprintf("select id, hostname from host where maintain_begin > %d or maintain_end < %d", now, now)
	rows, err := DB.Query(sql)
	if err != nil {
		log.Println("ERROR:", err)
		return hosts, err
	}

	defer rows.Close()
	for rows.Next() {
		t := model.Host{}
		err = rows.Scan(&t.Id, &t.Name)
		if err != nil {
			log.Println("WARN:", err)
			continue
		}
		hosts[t.Id] = &t
	}

	return hosts, nil
}

func QueryMonitoredHosts() (map[int]*model.Host, error) {
	log.Println("QueryMonitoredHosts")
	m := make(map[int]*model.Host)
	if redi.GetLock(redi.MonitoredHostsLockKey, redi.RedisDisTLockTimeOut) {
		dbm, err := queryMonitoredHosts()
		if err != nil {
			return m, err
		}
		redi.SetMonitoredHost2Redis(dbm)
		return dbm, nil
	} else { //没获取到锁 从redis中获取
		begin := time.Now().UnixNano()
		redism, err := redi.GetMonitoredHostFromRedis()
		log.Println("query redis cost:", (time.Now().UnixNano()-begin)/1000)
		//读取成功返回
		if err == nil {
			return redism, nil
		} else {
			//读取失败，从db中获取
			return queryMonitoredHosts()
		}
	}
}
