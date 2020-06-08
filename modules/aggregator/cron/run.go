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

package cron

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/open-falcon/falcon-plus/common/sdk/sender"
	"github.com/open-falcon/falcon-plus/modules/aggregator/g"
	"github.com/open-falcon/falcon-plus/modules/aggregator/sdk"
)

func WorkerRun(item *g.Cluster) {
	debug := g.Config().Debug
	/*
		Numerator代表分子    例如 $(cpu.user)+$(cpu.system) 代表求cpu.user和cpu.system的和
		Denominator代表分母  例如 $# 代表所有机器
	*/
	//cleanParam去除\r等字符
	numeratorStr := cleanParam(item.Numerator)
	denominatorStr := cleanParam(item.Denominator)
	//判断分子分母是否合法
	if !expressionValid(numeratorStr) || !expressionValid(denominatorStr) {
		log.Warnf("[W] invalid numerator or denominator", item)
		return
	}
	//判断分子分母是否需要计算
	needComputeNumerator := needCompute(numeratorStr)
	needComputeDenominator := needCompute(denominatorStr)
	//如果分子分母都不需要计算就不需要用到聚合器了
	if !needComputeNumerator && !needComputeDenominator {
		log.Warnf("[W] no need compute", item)
		return
	}
	//比如分子是这样的: "($(cpu.busy)+$(cpu.idle)-$(cpu.nice))>80"
	//那么parse的返回值为 [cpu.busy cpu.idle cpu.nice] [+ -] >80
	numeratorOperands, numeratorOperators, numeratorComputeMode := parse(numeratorStr, needComputeNumerator)
	denominatorOperands, denominatorOperators, denominatorComputeMode := parse(denominatorStr, needComputeDenominator)

	if !operatorsValid(numeratorOperators) || !operatorsValid(denominatorOperators) {
		log.Warnf("[W] operators invalid", item)
		return
	}
	/*add retry for gethostname bygid
	这里源码是动过sdk根据group_id查找组里面机器列表
	这里我进行了两点优化:
	1.sdk调用时没有加重试,http失败导致这次没有get到机器所以这个点就不算了导致断点
	2.原来的接口在机器量超过1k时就效率就会很慢 2w+机器需要8s,看了代码是用orm进行了多次查询而且附带了很多别的信息
	这里我只需要group_id对应endpoint_list所以我写了一个新的接口用一条raw_sql进行查询
	测试2w+的机器0.2s就能返回
	*/

	hostnames := sdk.HostnamesByID(item.GroupId)
	//没有机器当然不用算了
	if len(hostnames) == 0 {
		log.Warnf("[E] get 0 record hostname item:", item)
		return
	}

	now := time.Now().Unix()

	/*这里是调用graph/lastpoint这个api 查询最近一个点的数据
	1.机器是上面查到的主机列表
	2.counter这里做了合并 把所有要查的metirc都放在一个请求里面查询了
	3.查询的时候在api那边做了for循环 逐个item查询 估计这里也会拖慢速度
	4.查完之后计算下值推到发送队列
	*/
	valueMap, err := queryCounterLast(numeratorOperands, denominatorOperands, hostnames, now-int64(item.Step*2), now)
	if err != nil {
		log.Errorf("[E] get queryCounterLast", err, item)
		return
	}

	var numerator, denominator float64
	var validCount int

	for _, hostname := range hostnames {
		var numeratorVal, denominatorVal float64
		var err error

		if needComputeNumerator {
			numeratorVal, err = compute(numeratorOperands, numeratorOperators, numeratorComputeMode, hostname, valueMap)

			if debug && err != nil {
				log.Printf("[W] [hostname:%s] [numerator:%s] id:%d, err:%v", hostname, item.Numerator, item.Id, err)
			} else if debug {
				log.Printf("[D] [hostname:%s] [numerator:%s] id:%d, value:%0.4f", hostname, item.Numerator, item.Id, numeratorVal)
			}

			if err != nil {
				continue
			}
		}

		if needComputeDenominator {
			denominatorVal, err = compute(denominatorOperands, denominatorOperators, denominatorComputeMode, hostname, valueMap)

			if debug && err != nil {
				log.Warnf("[W] [hostname:%s] [denominator:%s] id:%d, err:%v", hostname, item.Denominator, item.Id, err)
			} else if debug {
				log.Debugf("[D] [hostname:%s] [denominator:%s] id:%d, value:%0.4f", hostname, item.Denominator, item.Id, denominatorVal)
			}

			if err != nil {
				continue
			}
		}

		if debug {
			log.Debugf("[D] hostname:%s  numerator:%0.4f  denominator:%0.4f  per:%0.4f\n", hostname, numeratorVal, denominatorVal, numeratorVal/denominatorVal)
		}
		numerator += numeratorVal
		denominator += denominatorVal
		validCount += 1
	}

	if !needComputeNumerator {
		if numeratorStr == "$#" {
			numerator = float64(validCount)
		} else {
			numerator, err = strconv.ParseFloat(numeratorStr, 64)
			if err != nil {
				log.Errorf("[E] strconv.ParseFloat(%s) fail %v, id:%d", numeratorStr, err, item.Id)
				return
			}
		}
	}

	if !needComputeDenominator {
		if denominatorStr == "$#" {
			denominator = float64(validCount)
		} else {
			denominator, err = strconv.ParseFloat(denominatorStr, 64)
			if err != nil {
				log.Errorf("[E] strconv.ParseFloat(%s) fail %v, id:%d", denominatorStr, err, item.Id)
				return
			}
		}
	}

	if denominator == 0 {
		//log.Println("[W] denominator == 0, id:", item.Id)
		return
	}

	if validCount == 0 {
		//log.Println("[W] validCount == 0, id:", item.Id)
		return
	}

	if debug {
		log.Debugf("[D] hostname:all  numerator:%0.4f  denominator:%0.4f  per:%0.4f\n", numerator, denominator, numerator/denominator)
	}
	sender.Push(item.Endpoint, item.Metric, item.Tags, numerator/denominator, item.DsType, int64(item.Step))
}

func parse(expression string, needCompute bool) (operands []string, operators []string, computeMode string) {
	if !needCompute {
		return
	}

	// e.g. $(cpu.busy)
	// e.g. $(cpu.busy)+$(cpu.idle)-$(cpu.nice)
	// e.g. $(cpu.busy)>=80
	// e.g. ($(cpu.busy)+$(cpu.idle)-$(cpu.nice))>80

	splitCounter, _ := regexp.Compile(`[\$\(\)]+`)
	items := splitCounter.Split(expression, -1)

	count := len(items)
	for i, val := range items[1 : count-1] {
		if i%2 == 0 {
			operands = append(operands, val)
		} else {
			operators = append(operators, val)
		}
	}
	computeMode = items[count-1]

	return
}

func cleanParam(val string) string {
	val = strings.TrimSpace(val)
	val = strings.Replace(val, " ", "", -1)
	val = strings.Replace(val, "\r", "", -1)
	val = strings.Replace(val, "\n", "", -1)
	val = strings.Replace(val, "\t", "", -1)
	return val
}

// $#
// 200
// $(cpu.busy) + $(cpu.idle)
func needCompute(val string) bool {
	return strings.Contains(val, "$(")
}

func expressionValid(val string) bool {
	// use chinese character?

	if strings.Contains(val, "（") || strings.Contains(val, "）") {
		return false
	}

	if val == "$#" {
		return true
	}

	// e.g. $(cpu.busy)
	// e.g. $(cpu.busy)+$(cpu.idle)-$(cpu.nice)
	matchMode0 := `^(\$\([^\(\)]+\)[+-])*\$\([^\(\)]+\)$`
	if ok, err := regexp.MatchString(matchMode0, val); err == nil && ok {
		return true
	}

	// e.g. $(cpu.busy)>=80
	matchMode1 := `^\$\([^\(\)]+\)(>|=|<|>=|<=)\d+(\.\d+)?$`
	if ok, err := regexp.MatchString(matchMode1, val); err == nil && ok {
		return true
	}

	// e.g. ($(cpu.busy)+$(cpu.idle)-$(cpu.nice))>80
	matchMode2 := `^\((\$\([^\(\)]+\)[+-])*\$\([^\(\)]+\)\)(>|=|<|>=|<=)\d+(\.\d+)?$`
	if ok, err := regexp.MatchString(matchMode2, val); err == nil && ok {
		return true
	}

	// e.g. 纯数字
	matchMode3 := `^\d+$`
	if ok, err := regexp.MatchString(matchMode3, val); err == nil && ok {
		return true
	}

	return false
}

func operatorsValid(ops []string) bool {
	count := len(ops)
	for i := 0; i < count; i++ {
		if ops[i] != "+" && ops[i] != "-" {
			return false
		}
	}
	return true
}
