package cron

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/orm"
	cmodel "github.com/open-falcon/falcon-plus/common/model"
	nlist "github.com/toolkits/container/list"

	"golang.org/x/sync/semaphore"

	"github.com/open-falcon/falcon-plus/common/rediscluster"
	"github.com/open-falcon/falcon-plus/common/sdk/sender"
	"github.com/open-falcon/falcon-plus/modules/polymetric/g"
	"github.com/open-falcon/falcon-plus/modules/polymetric/kafka"
	"github.com/open-falcon/falcon-plus/modules/polymetric/model"
	"github.com/open-falcon/falcon-plus/modules/polymetric/redi"
)

const (
	PolyTimeStep           = 30
	CounterTimeStep        = 30
	CounterType            = "COUNTER"
	GAUGEType              = "GAUGE"
	PolyStringSep          = "||"
	MultiPolySep           = "@@"
	RedisLockTimeout       = 55
	MaxQueueSize           = 400000 // 最多的为所有机器的点 不会超过40w
	RedisHMapTimeout       = 60 * 60
	RunPolyInterval        = 60
	PloyMetricStrategyLock = "ploy_metric_strategy_lock"
	EndPolyStrategyHashKey = "end_poly_strategy_hash_key"
	FalconGroupPolyType    = "falcon_group"
	AmsTagPolyType         = "ams_tag"
)

var (
	EndPolyMetricMap            = sync.Map{}
	EndPolyAmsTagMetricMap      = sync.Map{}
	EndPolyFalconGroupMetricMap = sync.Map{}
	PolyWorkerQueueMap          = sync.Map{}
	PolyTypeMap                 = sync.Map{}
	PolyDbConfigMap             = sync.Map{}
	PolyHistoryDataMap          = sync.Map{}
	GroupPolyMethod             = []string{"sum", "avg", "max", "min", "tp50", "tp90", "tp99"}
)

type PolyTickerWorker struct {
	Queue    *nlist.SafeListLimited
	Name     string
	Ticker   *time.Ticker
	Quit     chan struct{}
	Started  bool
	Interval int
}

type GeneralPoly struct {
	Result     chan interface{}
	ArgMap     map[string][]*model.PolyMetric
	MaxWorker  int64
	ActionFunc func(name string, strategys []*model.PolyMetric)
	Type       string
	Num        int
	//ResMap     sync.Map
}

type SingleEnd struct {
	Endpoint string  `json:"endpoint"`
	Value    float64 `json:"value"`
}

func (this *PolyTickerWorker) Start() {
	go func() {
		for {
			select {
			case <-this.Ticker.C:
				this.WorkerRun()
			case <-this.Quit:
				if g.Config().LogLevel == "debug" {
					log.Println("[I] drop worker")
				}
				this.Ticker.Stop()
				return
			}
		}
	}()
}

func (this *PolyTickerWorker) WorkerRun() {
	switch strings.Split(this.Name, PolyStringSep)[0] {
	case AmsTagPolyType:
		GeneralPolyMethods(this.Name, this.Queue)
	case FalconGroupPolyType:
		GeneralPolyMethods(this.Name, this.Queue)
	default:
		// 错误的聚合类型
		return
	}
}

func MultiRunWorker(gp *GeneralPoly) {

	if gp.Num <= 0 {
		return
	}
	thisPolyType := gp.Type
	if gp.Num > 10 {
		gp.MaxWorker = 10
	} else {
		tmp := int64(gp.Num) - 1
		if tmp <= 0 {
			gp.MaxWorker = 1
		} else {
			gp.MaxWorker = tmp
		}

	}
	gp.NewWorkPool()
	var resMap sync.Map
	for i := range gp.Result {

		//log.Infof("gp.Result:%+v", i)
		gpRes := i.(*GroupRes)
		if len(gpRes.Ends) == 0 && len(gp.Result) == 0 {
			switch thisPolyType {
			case AmsTagPolyType:
				EndPolyAmsTagMetricMap = sync.Map{}
			case FalconGroupPolyType:
				EndPolyFalconGroupMetricMap = sync.Map{}
			}
			break
		} else if len(gpRes.Ends) == 0 {
			continue
		}

		for _, end := range gpRes.Ends {
			endKey := end + SEP + gpRes.Counter
			polyKey := thisPolyType + SEP + gpRes.Name + SEP + gpRes.Counter
			tmp := []string{polyKey}
			res, loaded := resMap.LoadOrStore(endKey, tmp)
			if loaded {
				old := res.([]string)
				//old := cast.ToStringSlice(res)
				tm := make(map[string]string)
				for _, item := range old {
					tm[item] = item
				}
				if tm[polyKey] == "" {
					old = append(old, tmp...)
					resMap.Store(endKey, old)
				}

			}
		}

		if len(gp.Result) == 0 {
			break
		}

	}

	//f := func(k, v interface{}) bool {
	//	key := k.(string)
	//	va := v.([]string)
	//	log.Infof("wwwwwwwwwwww-k,v", key, va)
	//	return true
	//}
	//resMap.Range(f)
	//log.Printf("thisPolyType_resMap_thisPolyType:%+v", thisPolyType, resMap)
	switch thisPolyType {
	case AmsTagPolyType:
		EndPolyAmsTagMetricMap = resMap
	case FalconGroupPolyType:
		EndPolyFalconGroupMetricMap = resMap
	}

}

func (this *GeneralPoly) NewWorkPool() {
	ctx := context.TODO()
	MaxNum := this.Num
	result := make(chan interface{}, MaxNum)
	this.Result = result
	sem := semaphore.NewWeighted(int64(this.MaxWorker))
	for name, args := range this.ArgMap {
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Errorf("Failed to acquire semaphore: %v", err)
			break
		}
		go func(name string, strategys []*model.PolyMetric) {
			defer sem.Release(1)
			this.ActionFunc(name, strategys)
		}(name, args)
	}

	if err := sem.Acquire(ctx, int64(this.MaxWorker)); err != nil {
		log.Errorf("Failed to acquire semaphore: %v", err)
	}

}

func InitGroupStrategy() {
	//locked := rediscluster.GetDisLock(redi.RedisCluster, PloyMetricStrategyLock, RedisLockTimeout)
	locked := rediscluster.GetDisLock(redi.RedisCluster, PloyMetricStrategyLock, RedisLockTimeout)
	if locked {
		/*抢到锁后更新所有策略到redis map中*/
		//查询所有类型
		log.Debugf("get_redis_lock_success_start_work")
		SyncGroupStrategyFromDb()
		// type: Ams_tag
		RenewAmsTagStrategy()

		// type: falcon_group
		RenewFalconGroupStrategy()
		// 聚合所有map 的ends结果
		AggAllEndMap()
		// 跟新到redis 缓存中供 transfer查询
		SyncStrategyToCache()
		log.Debugf("get_redis_lock_success_end_work")
	} else {
		log.Debugf("get_redis_lock_failed_start_work")
		SyncGroupStrategyFromDb()
		CommonInitQueue(AmsTagPolyType)
		CommonInitQueue(FalconGroupPolyType)

	}
	log.Infof("")

}

func AggAllEndMap() {
	var tmpMap sync.Map

	f := func(k, v interface{}) bool {
		key := k.(string)
		va := v.([]string)
		//log.Infof("AggAllEndMap-k,v", key, va)
		var target []string

		if res, loaded := tmpMap.LoadOrStore(key, va); loaded == true {
			target = res.([]string)
			target = append(target, va...)
			tmpMap.Store(key, target)
		}
		return true
	}

	EndPolyAmsTagMetricMap.Range(f)
	EndPolyFalconGroupMetricMap.Range(f)
	EndPolyMetricMap = tmpMap

}

func SyncGroupStrategyFromDb() {
	// TODO mock的数据，上线后要删掉||策略的来源
	//var Mockres []*model.PolyMetric
	//a := &model.PolyMetric{Name: "system.monitor.falcon.testa", Counter: "testMetric", PolyType: "ams_tag"}
	////b := &model.PolyMetric{Name: "inf.abase", Counter: "testMetric", PolyType: "ams_tag"}
	//c := &model.PolyMetric{Name: "system.grafana", Counter: "testMetric", PolyType: "falcon_group"}
	//d := &model.PolyMetric{Name: "system.monitor.falcon.langfang", Counter: "testMetric", PolyType: "falcon_group"}
	////Mockres = append(Mockres, a, b, c, d)
	//Mockres = append(Mockres, a, c, d)
	db := orm.NewOrm()
	var Qres []*model.PolyMetric
	_, error := db.QueryTable("poly_metric").All(&Qres)
	if error != nil {
		log.Errorf("GetGroupStrategyFromDb_error:%+v", error)
	}

	var tmpMap sync.Map

	//Qres = Mockres
	for _, item := range Qres {
		//log.Infof("SyncGroupStrategyFromDb_item:%+v", item)
		tt := []*model.PolyMetric{item}
		if pList, loaded := tmpMap.LoadOrStore(item.PolyType, tt); loaded {
			tmp := pList.([]*model.PolyMetric)
			tmp = append(tmp, item)
			tmpMap.Store(item.PolyType, tmp)
		}
	}
	PolyDbConfigMap = tmpMap
	log.Debugf("SyncGroupStrategyFromDb_res_len:%d:%+v", len(Qres), PolyDbConfigMap)
}

func GeneralPolyMethods(Name string, Q *nlist.SafeListLimited) {
	Len := Q.Len()

	item := Q.PopBackBy(Len)
	count := len(item)
	if count == 0 {
		return
	}
	log.Infof("[GeneralPolyMethods]RunGroupPoly_called:Name:%s,len:%d", Name, count)
	var dataList []float64
	var numpList []SingleEnd
	var sum, avg, max, min, tp50, tp90, tp99 float64
	counterType := GAUGEType
	// 为了给出max、min等极值对应的endpoint
	singMap := make(map[float64]string)

	for _, i := range item {
		iF := i.(*cmodel.PolyRequest)
		//if counterType == "" {
		//	counterType = iF.Type
		//}

		va := iF.Value.(float64)
		endP := iF.EndPoint
		t := SingleEnd{
			Endpoint: endP,
			Value:    va,
		}
		numpList = append(numpList, t)
		if singMap[va] == "" {
			singMap[va] = endP
		}
		sum += va
		dataList = append(dataList, va)
	}
	realCount := len(dataList)
	if realCount == 0 {
		return
	}
	var pushSetp int64
	pushSetp = PolyTimeStep

	if realCount == 1 {
		sum = dataList[0]
		avg = dataList[0]
		max = dataList[0]
		min = dataList[0]
		tp50 = dataList[0]
		tp90 = dataList[0]
		tp99 = dataList[0]
	} else {
		sort.Float64s(dataList)

		max = dataList[realCount-1]
		min = dataList[0]
		avg = sum / float64(realCount)
		tp50 = dataList[int(float64(realCount)*0.5)]
		tp90 = dataList[int(float64(realCount)*0.95)]
		tp99 = dataList[int(float64(realCount)*0.99)]

	}
	// 本地map 做循环技术用
	localDataMap := make(map[string]float64)
	promeDataMap := make(map[string]float64)
	localDataMap["sum"] = sum
	localDataMap["avg"] = avg
	localDataMap["max"] = max
	localDataMap["min"] = min
	localDataMap["tp50"] = tp50
	localDataMap["tp90"] = tp90
	localDataMap["tp99"] = tp99

	names := strings.Split(Name, SEP)

	polyType := names[0]
	polyName := names[1]
	metric := names[2]
	endp := polyType + "_poly_" + polyName
	log.Infof("poly_res:endp sum, avg, max, min, tp50, tp90, tp99", endp, sum, avg, max, min, tp50, tp90, tp99)
	endNew := strings.Replace(endp, ".", "_", -1)

	tagPre := "method="
	//log.Infof("sum,avg,max,min,tp50,", sum, avg, max, min)
	sender.Push(endNew, metric, tagPre+"sum", sum, counterType, int64(pushSetp))
	sender.Push(endNew, metric, tagPre+"avg", avg, counterType, int64(pushSetp))
	sender.Push(endNew, metric, tagPre+"max", max, counterType, int64(pushSetp))
	sender.Push(endNew, metric, tagPre+"min", min, counterType, int64(pushSetp))
	sender.Push(endNew, metric, tagPre+"tp50", tp50, counterType, int64(pushSetp))
	sender.Push(endNew, metric, tagPre+"tp90", tp90, counterType, int64(pushSetp))
	sender.Push(endNew, metric, tagPre+"tp99", tp99, counterType, int64(pushSetp))

	/*
		根据内存中的值计算 rate 和delta
	*/
	for k, v := range localDataMap {
		promeDataMap[k] = v
		rate := 0.0
		delta := 0.0
		uniqueResultKey := endNew + metric + tagPre + k

		if lastPoint, loaded := PolyHistoryDataMap.Load(uniqueResultKey); loaded {
			log.Debugf("[localDataMap_lastPoint] key,this_value,last_value,%+v,%+v,%+v", k, v, lastPoint)
			lastP := lastPoint.(float64)
			delta = v - lastP
			if lastP == 0.0 {
				rate = 0.0
			} else {
				//rate = delta / lastP * 100.0
				rate = delta / lastP
			}

		}
		// 本次计算完毕，更新cache中的值
		PolyHistoryDataMap.Store(uniqueResultKey, v)
		log.Debugf("[localDataMap] key,this_value,rate delta ,%+v,%+v,%+v,%+v", k, v, rate, delta)
		sender.Push(endNew, metric+"_rate", tagPre+k, rate, counterType, int64(pushSetp))
		sender.Push(endNew, metric+"_delta", tagPre+k, delta, counterType, int64(pushSetp))
		promeDataMap[k+"_rate"] = rate
		promeDataMap[k+"_delta"] = delta

	}
	// push to prome
	if g.Config().Prome.Enabled {
		PushToProme(metric, polyName, promeDataMap)
	}
	// push到kafka
	if kafka.KafkaAsyncProducer != nil {
		maxEnd := singMap[max]
		minEnd := singMap[min]
		tp50End := singMap[tp50]
		tp90End := singMap[tp90]
		tp99End := singMap[tp99]
		AsyncPushKafka(polyType, polyName, maxEnd, metric, "max", max)
		AsyncPushKafka(polyType, polyName, minEnd, metric, "min", min)
		AsyncPushKafka(polyType, polyName, tp50End, metric, "tp50", tp50)
		AsyncPushKafka(polyType, polyName, tp90End, metric, "tp90", tp90)
		AsyncPushKafka(polyType, polyName, tp99End, metric, "tp99", tp99)
	}
	RpcCallNumpApi(metric, polyName, numpList)
	////outlier check
	//outlierStr := outlierCheck(dataList, singMap)
	//outPoint := outlier.GrpOutlier{
	//	GrpName:   polyName,
	//	PolyType:  polyType,
	//	Counter:   metric,
	//	Timestamp: time.Now().Unix(),
	//	Value:     outlierStr,
	//}
	//saveOutlier2DB(&outPoint)
}

func AsyncPushKafka(polyType, polyName, endpoint, counter, method string, value float64) {
	if kafka.KafkaAsyncProducer == nil {
		return
	}
	if endpoint == "" {
		log.Warnf("polyType, polyName, endpoint, counter, method string, value_empty_endpoint||", polyType, polyName, endpoint, counter, method, value)
		return
	}
	data := make(map[string]interface{})
	data["poly_type"] = polyType
	data["poly_name"] = polyName
	data["endpoint"] = endpoint
	data["counter"] = counter
	data["method"] = method
	data["timestamp"] = time.Now().Unix()
	data["value"] = value
	byteData, _ := json.Marshal(data)
	msg := &sarama.ProducerMessage{
		Topic: g.Config().Kafka.TopicName,
		//Key:   sarama.StringEncoder("go_test"),
	}

	msg.Value = sarama.ByteEncoder(byteData)
	log.Debugf("[AsyncPushKafka]input [%s]\n", msg.Value)

	// send to chain
	kafka.KafkaAsyncProducer.Input() <- msg
	// check send res
	select {
	case msg := <-kafka.KafkaAsyncProducer.Successes():
		log.Debugf("Success push to kafka:%+v,msg:%+v", data, msg)
	case fail := <-kafka.KafkaAsyncProducer.Errors():
		log.Errorf("err: %s\n", fail.Err.Error())

	}
}
