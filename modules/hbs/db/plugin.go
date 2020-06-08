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
	"crypto/md5"
	"fmt"
	"log"

	"github.com/open-falcon/falcon-plus/modules/hbs/g"
	"github.com/open-falcon/falcon-plus/modules/hbs/redi"
)

func QueryPluginsForEnd(hostname string) (ps []string) {
	rawSql := fmt.Sprintf("select pd.dir from plugin_dir pd,grp_host gh,host h  ,grp g where gh.grp_id=g.id and gh.host_id=h.id and pd.grp_id=gh.grp_id "+
		"and h.hostname='%s'", hostname)
	//g.Logger.Infof("rawSql:%s", rawSql)
	rows, err := DBRO.Query(rawSql)
	if err != nil {
		g.Logger.Errorf("ERROR:", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		var p string
		err = rows.Scan(&p)
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}
		ps = append(ps, p)
	}
	return
}

func queryPlugins() (map[int][]string, error) {
	m := make(map[int][]string)

	sql := "select grp_id, dir from plugin_dir"
	rows, err := DB.Query(sql)
	if err != nil {
		log.Println("ERROR:", err)
		return m, err
	}

	defer rows.Close()
	for rows.Next() {
		var (
			id  int
			dir string
		)

		err = rows.Scan(&id, &dir)
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}

		// 1个group对应多个plugin
		if _, exists := m[id]; exists {
			m[id] = append(m[id], dir)
		} else {
			m[id] = []string{dir}
		}
	}

	return m, nil
}

func getmd5(data []byte) string {
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)

}
func QueryPlugins() (map[int][]string, error) {
	m := make(map[int][]string)
	//获取锁成功，则同步db中数据，并刷新到缓存
	if redi.GetLock(redi.GroupPluginsLockKey, redi.RedisDisTLockTimeOut) {
		dbm, err := queryPlugins()
		if err != nil {
			return m, err
		}
		redi.SetPlugins2Redis(dbm)
		return dbm, nil
	} else { //没获取到锁 从redis中获取
		redism, err := redi.GetPluginsFromRedis()
		//读取成功返回
		if err == nil {
			return redism, nil
		} else {
			//读取失败，从db中获取
			return queryPlugins()
		}
	}
}
