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
	"time"

	"github.com/open-falcon/falcon-plus/modules/hbs/g"
	"github.com/open-falcon/falcon-plus/modules/hbs/redi"
)

// 一个机器ID对应了多个service ID
func queryHostServices() (map[string]string, error) {
	ret := make(map[string]string)
	rows, err := DB.Query("select svs_name,hostname from grp_svs, host, grp_host where grp_svs.grp_id=grp_host.grp_id and grp_host.host_id=host.id")
	if err != nil {
		log.Println("ERROR:", err)
		return ret, err
	}

	defer rows.Close()

	for rows.Next() {
		var sname, hname string

		err = rows.Scan(&sname, &hname)
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}

		if _, ok := ret[hname]; ok {
			ret[hname] = ret[hname] + "," + sname
		} else {
			ret[hname] = sname
		}
	}
	return ret, nil
}

func QueryServices() (map[string]string, error) {

	m := make(map[string]string)
	if redi.GetLock(redi.HostServiceIdsLockKey, 60) {
		dbm, err := queryHostServices() //map[hname][sname]string
		if err != nil {
			return m, err
		}
		redi.SetHostService2Redis(dbm)
		return dbm, nil
	} else { //没获取到锁 从redis中获取
		now := time.Now()
		redism, err := redi.GetHostServiceFromRedis()
		g.Logger.Infof("Query service redis cost:%+v", time.Since(now))
		//读取成功返回
		if err == nil {
			return redism, nil
		} else {
			//读取失败，从db中获取
			return queryHostServices()
		}
	}
}
