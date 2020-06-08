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

package rpc

import (
	"fmt"
	"strconv"
	"time"

	"strings"

	log "github.com/Sirupsen/logrus"
	cmodel "github.com/open-falcon/falcon-plus/common/model"
	cutils "github.com/open-falcon/falcon-plus/common/utils"
	"github.com/open-falcon/falcon-plus/modules/transfer/cron"
	"github.com/open-falcon/falcon-plus/modules/transfer/g"
	"github.com/open-falcon/falcon-plus/modules/transfer/proc"
	"github.com/open-falcon/falcon-plus/modules/transfer/sender"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	PolyStringSep = "||"
	MultiPolySep  = "@@"
	GAUGETYPE     = "GAUGE"
	COUNTERTYPE   = "COUNTER"
)

type Transfer int
type TransferResp struct {
	Msg        string
	Total      int
	ErrInvalid int
	Latency    int64
}

func (t *TransferResp) String() string {
	s := fmt.Sprintf("TransferResp total=%d, err_invalid=%d, latency=%dms",
		t.Total, t.ErrInvalid, t.Latency)
	if t.Msg != "" {
		s = fmt.Sprintf("%s, msg=%s", s, t.Msg)
	}
	return s
}

func (this *Transfer) Ping(req cmodel.NullRpcRequest, resp *cmodel.SimpleRpcResponse) error {
	return nil
}

func (t *Transfer) Update(args []*cmodel.MetricValue, reply *cmodel.TransferResponse) error {
	return RecvMetricValues(args, reply, "rpc")
}

// process new metric values
func RecvMetricValues(args []*cmodel.MetricValue, reply *cmodel.TransferResponse, from string) error {
	//log.Printf("[RecvMetricValues called] from %s", from, len(args))
	start := time.Now()
	reply.Invalid = 0
	items := []*cmodel.MetaData{}
	promeItems := []*cmodel.PromeMetaData{}
	polyItems := []*cmodel.PolyRequest{}
	dynamicMetricItems := []*cmodel.MetaData{}
	// 对接promethues

	for _, v := range args {
		if v == nil {
			reply.Invalid += 1
			continue
		}

		// 历史遗留问题.
		// 老版本agent上报的metric=kernel.hostname的数据,其取值为string类型,现在已经不支持了;所以,这里硬编码过滤掉
		if v.Metric == "kernel.hostname" {
			reply.Invalid += 1
			continue
		}

		if v.Metric == "" || v.Endpoint == "" {
			reply.Invalid += 1
			continue
		}

		if v.Type != g.COUNTER && v.Type != g.GAUGE && v.Type != g.DERIVE {
			reply.Invalid += 1
			continue
		}

		if v.Value == "" {
			reply.Invalid += 1
			continue
		}

		if v.Step <= 0 {
			reply.Invalid += 1
			continue
		}

		if len(v.Metric)+len(v.Tags) > 510 {
			reply.Invalid += 1
			continue
		}

		// TODO 呵呵,这里需要再优雅一点
		now := start.Unix()
		if v.Timestamp <= 0 || v.Timestamp > now+3*v.Step || v.Timestamp < now-3*v.Step {
			v.Timestamp = now
		}

		tagsMap := cutils.DictedTagstring(v.Tags)
		fv := &cmodel.MetaData{
			Metric:      v.Metric,
			Endpoint:    v.Endpoint,
			Timestamp:   v.Timestamp,
			Step:        v.Step,
			CounterType: v.Type,
			Tags:        tagsMap, //TODO tags键值对的个数,要做一下限制
		}
		valid := true
		var vv float64
		var err error

		switch cv := v.Value.(type) {
		case string:
			vv, err = strconv.ParseFloat(cv, 64)
			if err != nil {
				valid = false
			}
		case float64:
			vv = cv
		case int64:
			vv = float64(cv)
		default:
			valid = false
		}

		if !valid {
			reply.Invalid += 1
			continue
		}

		fv.Value = vv
		if _, ok := fv.Tags["DynamicMetric"]; ok {
			if fv.Tags["DynamicMetric"] == "true" {
				dynamicMetricItems = append(dynamicMetricItems, fv)
				continue
			}
		}
		counter := cutils.Counter(v.Metric, tagsMap)
		// 聚合策略
		polyStrategy := cron.EndPolyCache.Get(v.Endpoint + PolyStringSep + counter)
		//log.Printf("counter,polyStrategy:%+v,%+v,", counter, polyStrategy)
		polyValue := fv.Value
		if polyStrategy != "" {
			// 检查是否是counter类型,如果是去内存的map中获取上一个点算成gauge
			if fv.CounterType == "COUNTER" {
				endCounterKey := fv.Endpoint + counter
				existRes, loaded := cron.PolyCounterCache.LoadOrStore(endCounterKey, fv)
				if loaded {
					// 说明有了上一次的数据，转换gauge推数据
					lastV := existRes.(*cmodel.MetaData)
					if fv.Timestamp-lastV.Timestamp <= 0 {
						// 做下防护
						continue
					}
					newData := (fv.Value - lastV.Value) / float64(fv.Timestamp-lastV.Timestamp)
					polyValue = newData
				} else {
					// 说明第一次存储 不推数据
					continue
				}
			}

			polys := strings.Split(polyStrategy, MultiPolySep)
			for _, poly := range polys {
				t := &cmodel.PolyRequest{}
				t.PolyName = poly
				t.Value = polyValue
				t.Type = fv.CounterType
				t.EndPoint = fv.Endpoint
				polyItems = append(polyItems, t)
			}
			// 推送至prometheus
			if g.Config().Prome.Enabled {
				tmp := &cmodel.PromeMetaData{}
				tmp.FalconMetaData = fv
				tmp.Polys = polys
				if len(tmp.Polys) >= 1 {
					promeItems = append(promeItems, tmp)
				}

			}

		}
		items = append(items, fv)

	}

	// statistics
	cnt := int64(len(items))
	dynamicMetricCnt := int64(len(dynamicMetricItems))
	proc.RecvCnt.IncrBy(cnt + dynamicMetricCnt)
	if from == "rpc" {
		proc.RpcRecvCnt.IncrBy(cnt + dynamicMetricCnt)
	} else if from == "http" {
		proc.HttpRecvCnt.IncrBy(cnt + dynamicMetricCnt)
	}

	cfg := g.Config()

	if cfg.Graph.Enabled {
		sender.Push2GraphSendQueue(items)
	}

	if cfg.Judge.Enabled {
		sender.Push2JudgeSendQueue(items)
	}

	if cfg.Poly.Enabled {
		sender.Push2PolySendQueue(polyItems)
	}

	if cfg.Tsdb.Enabled {
		sender.Push2TsdbSendQueue(items)
	}
	//新增kafka通道
	if cfg.Kafka.Enabled && g.Config().Kafka.DatabusChannel != "" {
		sender.Push2KafkaSendQueue(items)
	}

	//if cfg.InfluxDB.Enabled && dynamicMetricCnt > 0 {
	//	sender.Push2InfluxDBSendQueue(dynamicMetricItems)
	//}

	if cfg.Prome.Enabled {
		go PromeWork(promeItems)
	}

	reply.Message = "ok"
	reply.Total = len(args)
	reply.Latency = (time.Now().UnixNano() - start.UnixNano()) / 1000000

	return nil
}

func PromeWork(dataS []*cmodel.PromeMetaData) {
	if len(dataS) <= 0 {
		return
	}
	localReg := prometheus.NewRegistry()
	thisEnd := dataS[0].FalconMetaData.Endpoint
	if thisEnd == "" {
		return
	}

	for _, Pdata := range dataS {
		data := Pdata.FalconMetaData

		var promMetric string
		labelMap := make(map[string]string)
		if len(data.Tags) > 0 {
			for key, v := range data.Tags {
				//tagKeys = append(tagKeys, key)
				labelMap[key] = v
			}
			//labels = append(labels, tagKeys...)
		}
		for _, poly := range Pdata.Polys {
			groupName := strings.Split(poly, PolyStringSep)[1]
			newGrpName := strings.Replace(groupName, ".", "_", -1)
			labelMap["falcon_group_"+newGrpName] = newGrpName
		}
		// 把endpoint 作为key 加上

		promMetric = strings.Replace(data.Metric, ".", "_", -1)
		labelMap["endpoint"] = data.Endpoint
		//labelMap["uniq_key_"+strings.Replace(data.Endpoint, "-", "_", -1)+"_"+promMetric] = data.Endpoint
		//labelMap["instance"] = data.Endpoint
		//log.Printf("labelMap:%+v data:%+v", labelMap, data)

		switch data.CounterType {
		case GAUGETYPE:
			metricTmp := prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        promMetric,
				Help:        promMetric + " help info",
				ConstLabels: labelMap,
			})
			//metricTmp := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			//	Name: promMetric,
			//	Help: promMetric + " help info",
			//},
			//	labels,
			//)
			//sender.LocalReg.Register(metricTmp)
			metricTmp.Set(float64(data.Value))
			//localReg.MustRegister(metricTmp)
			err := localReg.Register(metricTmp)
			if err != nil {
				log.Errorf("[Register_GAUGE_error]:%+v", err)
				return
			}

			//metricTmp.With(labelMap).Set(float64(data.Value))

		case COUNTERTYPE:

			//metricTmp := prometheus.NewCounterVec(prometheus.CounterOpts{
			metricTmp := prometheus.NewCounter(prometheus.CounterOpts{
				Name:        promMetric,
				Help:        promMetric + " help info",
				ConstLabels: labelMap,
			},
			)
			metricTmp.Add(float64(data.Value))
			err := localReg.Register(metricTmp)
			if err != nil {
				log.Errorf("[Register_counter_error]:%+v", err)
				return
			}

		}

	}

	//h := md5.New()
	//h.Write([]byte(uniqJobKey))
	//jobName := hex.EncodeToString(h.Sum(nil))[:10]

	promPusher := push.New(g.Config().Prome.Address, "falcon_transfer_"+thisEnd).Gatherer(localReg)

	//pusherError := promPusher.Push()
	pusherError := promPusher.Add()

	//log.Printf("[localReg]:%+v  [PushCron] error:%+v", localReg, pusherError)
	log.Debugf("[localReg]:%+v  [PushCron] error:%+v", localReg, pusherError)

}
