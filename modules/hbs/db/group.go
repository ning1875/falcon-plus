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
	"log"

	"github.com/open-falcon/falcon-plus/modules/hbs/redi"
)

func queryHostGroups() (map[int][]int, error) {
	m := make(map[int][]int)

	sql := "select grp_id, host_id from grp_host"
	rows, err := DB.Query(sql)
	if err != nil {
		log.Println("ERROR:", err)
		return m, err
	}

	defer rows.Close()
	for rows.Next() {
		var gid, hid int
		err = rows.Scan(&gid, &hid)
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}

		if _, exists := m[hid]; exists {
			m[hid] = append(m[hid], gid)
		} else {
			m[hid] = []int{gid}
		}
	}

	return m, nil
}

func QueryHostGroups() (map[int][]int, error) {
	log.Println("QueryHostGroups")
	m := make(map[int][]int)
	if redi.GetLock(redi.HostGroupsMapLockKey, redi.RedisDisTLockTimeOut) {
		dbm, err := queryHostGroups()
		if err != nil {
			return m, err
		}
		redi.SetHostGroup2Redis(dbm)
		return dbm, nil
	} else { //没获取到锁 从redis中获取
		redism, err := redi.GetHostGroupFromRedis()
		//读取成功返回
		if err == nil {
			return redism, nil
		} else {
			//读取失败，从db中获取
			return queryHostGroups()
		}
	}
}
