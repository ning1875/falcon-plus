package cron

import (
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/polymetric/g"
	"github.com/open-falcon/falcon-plus/modules/polymetric/model"
	"github.com/open-falcon/falcon-plus/modules/polymetric/redi"
	nlist "github.com/toolkits/container/list"
)

const (
	SEP = PolyStringSep
)

type GroupRes struct {
	Name    string
	Counter string
	Ends    []string
}

func CreatePloyQueueIfNeed(name string) {
	Q := nlist.NewSafeListLimited(MaxQueueSize)
	W := &PolyTickerWorker{}
	W.Name = name
	W.Queue = Q

	PolyWorkerQueueMap.LoadOrStore(name, W)
}

func SyncStrategyToCache() {
	/*
		最终拉到redis中的数据
		162) "ams_tag||system.monitor.falcon.testa||testMetric,falcon_group||system.monitor.falcon.langfang||testMetric"
		163) "n8-037-210||testMetric"
		164) "ams_tag||system.monitor.falcon.testa||testMetric,falcon_group||system.monitor.falcon.langfang||testMetric"
		165) "n19-030-231||testMetric"
		166) "ams_tag||system.monitor.falcon.testa||testMetric,falcon_group||system.monitor.falcon.langfang||testMetric"
		167) "n14-068-145||testMetric"
		168) "ams_tag||system.monitor.falcon.testa||testMetric,falcon_group||system.monitor.falcon.langfang||testMetric"
		169) "n25-090-048||testMetric"
		170) "ams_tag||system.monitor.falcon.testa||testMetric"
		171) "n17-024-025||testMetric"
		172) "ams_tag||system.monitor.falcon.testa||testMetric,falcon_group||system.monitor.falcon.langfang||testMetric"
		173) "n14-068-150||testMetric"
		174) "ams_tag||system.monitor.falcon.testa||testMetric,falcon_group||system.monitor.falcon.langfang||testMetric"
		175) "n19-030-013||testMetric"
		176) "ams_tag||system.monitor.falcon.testa||testMetric,falcon_group||system.monitor.falcon.langfang||testMetric"
		177) "n108-154-241||testMetric"
		178) "ams_tag||system.monitor.falcon.testa||testMetric,falcon_group||system.monitor.falcon.langfang||testMetric"
		179) "n14-068-139||testMetric"
		180) "ams_tag||system.monitor.falcon.testa||testMetric,falcon_group||system.monitor.falcon.langfang||testMetric"
		181) "n17-037-227||testMetric"
		182) "ams_tag||system.monitor.falcon.testa||testMetric,falcon_group||system.monitor.falcon.langfang||testMetric"
		183) "n8-037-215||testMetric"
		184) "ams_tag||system.monitor.falcon.testa||testMetric,falcon_group||system.monitor.falcon.langfang||testMetric"

	*/
	var hsetValue []interface{}
	ThisRedisKey := EndPolyStrategyHashKey + strconv.FormatInt(time.Now().UnixNano(), 10)
	hsetValue = append(hsetValue, ThisRedisKey)

	f := func(k, v interface{}) bool {
		hsetValue = append(hsetValue, k.(string), strings.Join(v.([]string), MultiPolySep))

		return true
	}

	EndPolyMetricMap.Range(f)
	_, err := redi.RedisCluster.Do("HMSET", hsetValue...)
	if err != nil {
		return
	}
	_, err = redi.RedisCluster.Do("EXPIRE", ThisRedisKey, RedisHMapTimeout)
	if err != nil {
		return
	}
	_, err = redi.RedisCluster.Do("SET", EndPolyStrategyHashKey, ThisRedisKey)
	if err != nil {
		return
	}
}

func CommonInitQueue(poly_type string) (map[string][]*model.PolyMetric, int) {
	TagPs, loaded := PolyDbConfigMap.Load(poly_type)
	if loaded == false {
		return nil, 0
	}
	res := TagPs.([]*model.PolyMetric)
	//log.Infof("CommonInitQueue_res:%+v", res)
	resMap := make(map[string][]*model.PolyMetric)
	for _, i := range res {
		queueName := poly_type + SEP + i.Name + SEP + i.Counter
		CreatePloyQueueIfNeed(queueName)
		if tL := resMap[i.Name]; len(tL) >= 1 {
			tL = append(tL, i)
			resMap[i.Name] = tL
		} else {
			var tt []*model.PolyMetric
			tt = append(tt, i)
			resMap[i.Name] = tt
		}
	}
	log.Debugf("CommonInitQueue_resMap, len(res)", resMap, len(res))
	return resMap, len(res)
}

// 更新ams_tag类型的策略
func RenewAmsTagStrategy() {
	res, num := CommonInitQueue(AmsTagPolyType)
	if res == nil {
		return
	}
	gp := &GeneralPoly{}
	gp.Type = AmsTagPolyType
	gp.ActionFunc = gp.AmsTagHttpwork
	gp.ArgMap = res
	gp.Num = num
	MultiRunWorker(gp)
}

func (this *GeneralPoly) AmsTagHttpwork(name string, strategys []*model.PolyMetric) {

	ends := g.GetAmsTagIps(name)
	for _, item := range strategys {
		a := &GroupRes{}
		a.Name = name
		a.Counter = item.Counter
		a.Ends = ends
		this.Result <- a
	}

}
