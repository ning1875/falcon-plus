package cron

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	cutils "github.com/open-falcon/falcon-plus/common/utils"
	"github.com/open-falcon/falcon-plus/modules/polymetric/g"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

func PushToProme(metric, polyName string, dataMap map[string]float64) {
	if g.Config().Prome.Enabled == false {
		log.Infof("[PushToProme][disabled]")
		return
	}
	/*
		给prometheus 准备
	*/
	newM, tagM := cutils.ConunterToMetricAndTags(metric)
	labelMap := make(map[string]string)
	if len(tagM) > 0 {
		for key, v := range tagM {
			labelMap[key] = v
		}
	}
	labelMap["group_name"] = polyName
	localReg := prometheus.NewRegistry()
	promPusher := push.New(g.Config().Prome.Address, "falcon_poly_"+polyName)
	for k, v := range dataMap {
		// poly_res_cpu_busy_max{group_name="system.monitor"}
		// poly_res_cpu_busy_avg
		promMetric := "poly_res_" + strings.Replace(newM, ".", "_", -1) + "_" + k

		metricTmp := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        promMetric,
			Help:        promMetric + " help info",
			ConstLabels: labelMap,
		})
		err := localReg.Register(metricTmp)
		if err != nil {
			log.Errorf("[Register_counter_error]:%+v", err)
			return
		}
		metricTmp.Set(float64(v))

	}
	//log.Printf("PushToProme:%+v", localReg)
	pusherError := promPusher.Gatherer(localReg).Add()
	log.Debugf("[PushCron] error:%+v", pusherError)
}
