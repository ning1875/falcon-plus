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
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"strconv"

	"github.com/open-falcon/falcon-plus/common/model"

	"github.com/open-falcon/falcon-plus/common/utils"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
	"github.com/open-falcon/falcon-plus/modules/hbs/redi"
	"github.com/toolkits/container/set"
)

// 组合策略的本地map
// key 为组合策略表中的id
// value 为对应的策略列表
//var UnionStrategiesDbMap = make(map[int][]*model.Strategy)

// 获取所有的Strategy列表

func QueryStrategiesForOneEnd(endpoint string) (ss []*model.Strategy) {

	rawSql := fmt.Sprintf(
		"select "+
			"sy.id, "+
			"sy.metric, "+
			"sy.tags, "+
			"sy.func, "+
			"sy.op, "+
			"sy.right_value, "+
			"sy.max_step, "+
			"sy.priority, "+
			"sy.note, "+
			"sy.union_strategy_id, "+
			"tpl.id, "+
			"tpl.tpl_name, "+
			"tpl.parent_id, "+
			"tpl.action_id, "+
			"tpl.create_user "+
			"from strategy sy,tpl ,grp g,grp_tpl gt, grp_host gh,host h where sy.tpl_id=tpl.id  and tpl.id=gt.tpl_id and gt.grp_id=g.id  and gh.grp_id=g.id and gh.host_id=h.id and h.hostname='%s' "+
			"UNION ALL "+
			"select "+
			"sy.id, "+
			"sy.metric, "+
			"sy.tags, "+
			"sy.func, "+
			"sy.op, "+
			"sy.right_value, "+
			"sy.max_step, "+
			"sy.priority, "+
			"sy.note, "+
			"sy.union_strategy_id, "+
			"tpl.id, "+
			"tpl.tpl_name, "+
			"tpl.parent_id, "+
			"tpl.action_id, "+
			"tpl.create_user "+
			"from strategy sy,tpl ,grp g,grp_tpl gt, grp_host gh,host h where sy.tpl_id=tpl.parent_id  and tpl.id=gt.tpl_id and gt.grp_id=g.id  and gh.grp_id=g.id and gh.host_id=h.id and h.hostname='%s' ",
		endpoint,
		endpoint,
	)
	//g.Logger.Infof("rawSql", rawSql)
	rows, err := DBRO.Query(rawSql)
	if err != nil {
		g.Logger.Errorf("ERROR:", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		s := model.Strategy{}
		t := model.Template{}
		var tags string
		err = rows.Scan(&s.Id, &s.Metric, &tags, &s.Func, &s.Operator, &s.RightValue, &s.MaxStep, &s.Priority, &s.Note, &s.UnionStrategyId,
			&t.Id,
			&t.Name,
			&t.ParentId,
			&t.ActionId,
			&t.Creator,
		)
		if err != nil {
			g.Logger.Errorf("ERROR:", err)
			continue
		}

		_, tt := utils.SplitTagsString(tags)
		s.Tags = tt

		s.Tpl = &t
		ss = append(ss, &s)

	}
	return

}

func QueryUnionStrategiesByUnionId(unionId int) (ss []*model.Strategy) {
	rawSql := fmt.Sprintf(" select "+
		"sy.id, "+
		"sy.metric, "+
		"sy.tags, "+
		"sy.func, "+
		"sy.op, "+
		"sy.right_value, "+
		"sy.max_step, "+
		"sy.priority, "+
		"sy.note, "+
		"sy.union_strategy_id, "+
		"tpl.id, "+
		"tpl.tpl_name, "+
		"tpl.parent_id, "+
		"tpl.action_id, "+
		"tpl.create_user "+
		"from strategy sy,tpl where sy.tpl_id=tpl.id and  sy.union_strategy_id=%d",
		unionId)
	g.Logger.Infof("rawSql:%s", rawSql)
	rows, err := DBRO.Query(rawSql)
	if err != nil {
		g.Logger.Errorf("ERROR:", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		s := model.Strategy{}
		t := model.Template{}
		var tags string
		err = rows.Scan(&s.Id, &s.Metric, &tags, &s.Func, &s.Operator, &s.RightValue, &s.MaxStep, &s.Priority, &s.Note, &s.UnionStrategyId,
			&t.Id,
			&t.Name,
			&t.ParentId,
			&t.ActionId,
			&t.Creator,
		)
		if err != nil {
			g.Logger.Errorf("ERROR:", err)
			continue
		}

		_, tt := utils.SplitTagsString(tags)
		s.Tags = tt

		s.Tpl = &t
		ss = append(ss, &s)

	}
	return
}

func queryStrategies(tpls map[int]*model.Template) (map[int]*model.Strategy, map[int][]*model.Strategy, error) {
	ret := make(map[int]*model.Strategy)
	chainMap := make(map[int][]*model.Strategy)
	if tpls == nil || len(tpls) == 0 {
		return ret, chainMap, fmt.Errorf("illegal argument")
	}

	now := time.Now().Format("15:04")
	sql := fmt.Sprintf(
		"select %s from strategy as s where (s.run_begin='' and s.run_end='') "+
			"or (s.run_begin <= '%s' and s.run_end > '%s')"+
			"or (s.run_begin > s.run_end and !(s.run_begin > '%s' and s.run_end < '%s'))",
		"s.id, s.metric, s.tags, s.func, s.op, s.right_value, s.max_step, s.priority, s.note, s.tpl_id,s.union_strategy_id",
		now,
		now,
		now,
		now,
	)

	rows, err := DB.Query(sql)
	if err != nil {
		log.Println("ERROR:", err)
		return ret, chainMap, err
	}

	defer rows.Close()
	for rows.Next() {
		s := model.Strategy{}
		var tags string
		var tid int
		err = rows.Scan(&s.Id, &s.Metric, &tags, &s.Func, &s.Operator, &s.RightValue, &s.MaxStep, &s.Priority, &s.Note, &tid, &s.UnionStrategyId)
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}

		tt := make(map[string]string)

		if tags != "" {
			arr := strings.Split(tags, ",")
			for _, tag := range arr {
				kv := strings.SplitN(tag, "=", 2)
				if len(kv) != 2 {
					continue
				}
				tt[kv[0]] = kv[1]
			}
		}

		s.Tags = tt
		s.Tpl = tpls[tid]
		if s.Tpl == nil {
			//log.Printf("WARN: tpl is nil. strategy id=%d, tpl id=%d", s.Id, tid)
			// 如果Strategy没有对应的Tpl，那就没有action，就没法报警，无需往后传递了
			continue
		}

		ret[s.Id] = &s
	}
	// 查询组合策略这张表
	// 根据union_id 反查所有的 strategy
	//  chainMap组合id --> s_chain
	//chainMap := make(map[int][]*model.Strategy)
	type unionStrategy struct {
		Id          int
		StrategyIds string
	}
	unionSql := "select id,strategy_ids from union_strategy"

	rowsU, err := DB.Query(unionSql)
	defer rowsU.Close()

	for rowsU.Next() {
		us := unionStrategy{}
		err = rowsU.Scan(&us.Id, &us.StrategyIds)
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}

		var sChain []*model.Strategy

		for _, oneId := range strings.Split(us.StrategyIds, ",") {

			oneIdN, _ := strconv.ParseInt(oneId, 10, 64)

			if thisS := ret[int(oneIdN)]; thisS != nil {
				sChain = append(sChain, thisS)
			}

		}
		if len(sChain) > 1 {
			chainMap[us.Id] = sChain
		}

	}
	//if len(chainMap) >= 1 {
	//	UnionStrategiesDbMap = chainMap
	//}
	return ret, chainMap, nil
}

func RemoveIndex(s []int, index int) []int {

	s = append(s[:index], s[index+1:]...)
	return s
}

func SliceDiffMap(all []int) map[int][]int {

	//fmt.Println(all) //[0 1 2 3 4 5 6 7 8 9]
	m := make(map[int][]int)
	leng := len(all)
	for i, v := range all {
		tmp := make([]int, leng)
		copy(tmp, all)
		n := RemoveIndex(tmp, i)
		m[v] = n
	}
	return m
}

func QueryStrategies(tpls map[int]*model.Template) (map[int]*model.Strategy, map[int][]*model.Strategy, error) {
	m := make(map[int]*model.Strategy)
	unionM := make(map[int][]*model.Strategy)
	if redi.GetLock(redi.StrategiesLockKey, redi.RedisDisTLockTimeOut) {
		dbm, uniondbM, err := queryStrategies(tpls)
		log.Debugf("[QueryStrategies]dbm, uniondbM, err :%+v,%+v,", uniondbM, err)
		if err != nil {
			return m, unionM, err
		}
		redi.SetStrategies2Redis(dbm)
		redi.SetUnionStrategies2Redis(uniondbM)
		return dbm, uniondbM, nil
	} else { //没获取到锁 从redis中获取
		redism, err := redi.GetStrategiesFromRedis()
		redisUnM, errU := redi.GetUnionStrategiesFromRedis()
		log.Debugf("[QueryStrategies]抢锁失败,union_map:%+v", redisUnM)
		//读取成功返回
		if err == nil && errU == nil {
			return redism, redisUnM, nil
		} else {
			//读取失败，从db中获取
			return queryStrategies(tpls)
		}
	}
}

func QueryHostSpecialMetric(hostname string) (sps []*model.BuiltinMetric) {
	rawSql := fmt.Sprintf(""+
		" select "+
		" sy.metric,"+
		" sy.tags "+
		" from strategy sy,tpl ,grp g,grp_tpl gt, grp_host gh,host h where sy.tpl_id=tpl.id "+
		" and tpl.id=gt.tpl_id and gt.grp_id=g.id  and gh.grp_id=g.id and gh.host_id=h.id "+
		" and sy.metric in ('net.port.listen', 'proc.num', 'du.bs', 'url.check.health') "+
		" and h.hostname='%s' "+
		" UNION ALL "+
		" select "+
		" sy.metric,"+
		" sy.tags "+
		" from strategy sy,tpl ,grp g,grp_tpl gt, grp_host gh,host h where sy.tpl_id=tpl.parent_id "+
		" and tpl.id=gt.tpl_id and gt.grp_id=g.id  and gh.grp_id=g.id and gh.host_id=h.id "+
		" and sy.metric in ('net.port.listen', 'proc.num', 'du.bs', 'url.check.health') "+
		" and h.hostname='%s' ",
		hostname,
		hostname,
	)

	//g.Logger.Infof("rawSql:%s", rawSql)
	rows, err := DBRO.Query(rawSql)
	if err != nil {
		g.Logger.Errorf("ERROR:", err)
		return
	}
	metricTagsSet := set.NewStringSet()

	defer rows.Close()
	for rows.Next() {
		builtinMetric := model.BuiltinMetric{}
		err = rows.Scan(&builtinMetric.Metric, &builtinMetric.Tags)
		if err != nil {
			g.Logger.Warningf("WARN:", err)
			continue
		}

		k := fmt.Sprintf("%s%s", builtinMetric.Metric, builtinMetric.Tags)
		if metricTagsSet.Exists(k) {
			continue
		}

		sps = append(sps, &builtinMetric)
		metricTagsSet.Add(k)
	}

	return

}

func QueryBuiltinMetrics(tids string) ([]*model.BuiltinMetric, error) {
	sql := fmt.Sprintf(
		"select metric, tags from strategy where tpl_id in (%s) and metric in ('net.port.listen', 'proc.num', 'du.bs', 'url.check.health')",
		tids,
	)
	ret := []*model.BuiltinMetric{}

	rows, err := DB.Query(sql)
	if err != nil {
		log.Println("ERROR:", err)
		return ret, err
	}

	metricTagsSet := set.NewStringSet()

	defer rows.Close()
	for rows.Next() {
		builtinMetric := model.BuiltinMetric{}
		err = rows.Scan(&builtinMetric.Metric, &builtinMetric.Tags)
		if err != nil {
			log.Println("WARN:", err)
			continue
		}

		k := fmt.Sprintf("%s%s", builtinMetric.Metric, builtinMetric.Tags)
		if metricTagsSet.Exists(k) {
			continue
		}

		ret = append(ret, &builtinMetric)
		metricTagsSet.Add(k)
	}

	return ret, nil
}
