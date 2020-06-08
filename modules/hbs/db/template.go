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

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/redi"
)

func queryGroupTemplates() (map[int][]int, error) {
	m := make(map[int][]int)

	sql := "select grp_id, tpl_id from grp_tpl"
	rows, err := DB.Query(sql)
	if err != nil {
		log.Println("ERROR:", err)
		return m, err
	}

	defer rows.Close()
	for rows.Next() {
		var gid, tid int
		err = rows.Scan(&gid, &tid)
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}

		if _, exists := m[gid]; exists {
			m[gid] = append(m[gid], tid)
		} else {
			m[gid] = []int{tid}
		}
	}

	return m, nil
}
func QueryGroupTemplates() (map[int][]int, error) {
	m := make(map[int][]int)
	if redi.GetLock(redi.GroupTemplatesLockKey, redi.RedisDisTLockTimeOut) {
		dbm, err := queryGroupTemplates()
		if err != nil {
			return m, err
		}
		redi.SetGroupTemplate2Redis(dbm)
		return dbm, nil
	} else { //没获取到锁 从redis中获取
		redism, err := redi.GetGroupTemplateFromRedis()
		//读取成功返回
		if err == nil {
			return redism, nil
		} else {
			//读取失败，从db中获取
			return queryGroupTemplates()
		}
	}
}

// 获取所有的策略模板列表
func queryTemplates() (map[int]*model.Template, error) {

	templates := make(map[int]*model.Template)

	sql := "select id, tpl_name, parent_id, action_id, create_user from tpl"
	rows, err := DB.Query(sql)
	if err != nil {
		log.Println("ERROR:", err)
		return templates, err
	}

	defer rows.Close()
	for rows.Next() {
		t := model.Template{}
		err = rows.Scan(&t.Id, &t.Name, &t.ParentId, &t.ActionId, &t.Creator)
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}
		templates[t.Id] = &t
	}

	return templates, nil
}
func QueryTemplates() (map[int]*model.Template, error) {
	m := make(map[int]*model.Template)
	if redi.GetLock(redi.TemplateCacheLockKey, redi.RedisDisTLockTimeOut) {
		dbm, err := queryTemplates()
		if err != nil {
			return m, err
		}
		redi.SetTPL2Redis(dbm)
		return dbm, nil
	} else { //没获取到锁 从redis中获取
		redism, err := redi.GetTPLFromRedis()
		//读取成功返回
		if err == nil {
			return redism, nil
		} else {
			//读取失败，从db中获取
			return queryTemplates()
		}
	}
}

// 一个机器ID对应了多个模板ID
func queryHostTemplateIds() (map[int][]int, error) {
	ret := make(map[int][]int)
	rows, err := DB.Query("select a.tpl_id, b.host_id from grp_tpl as a inner join grp_host as b on a.grp_id=b.grp_id")
	if err != nil {
		log.Println("ERROR:", err)
		return ret, err
	}

	defer rows.Close()

	for rows.Next() {

		var tid, hid int

		err = rows.Scan(&tid, &hid)
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}

		if _, ok := ret[hid]; ok {
			ret[hid] = append(ret[hid], tid)
		} else {
			ret[hid] = []int{tid}
		}
	}
	return ret, nil
}

func QueryHostTemplateIds() (map[int][]int, error) {
	log.Println("QueryHostTemplateIds")
	m := make(map[int][]int)
	if redi.GetLock(redi.HostTemplateIdsLockKey, redi.RedisDisTLockTimeOut) {
		dbm, err := queryHostTemplateIds()
		if err != nil {
			return m, err
		}
		redi.SetHostTemplate2Redis(dbm)
		return dbm, nil
	} else { //没获取到锁 从redis中获取
		begin := time.Now().UnixNano()
		redism, err := redi.GetHostTemplateFromRedis()
		log.Println("query redis cost:", (time.Now().UnixNano()-begin)/1000)
		//读取成功返回
		if err == nil {
			return redism, nil
		} else {
			//读取失败，从db中获取
			return queryHostTemplateIds()
		}
	}
}
