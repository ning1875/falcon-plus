package funcs

import (
	"github.com/open-falcon/falcon-plus/common/model"
)

func Test1() []*model.MetricValue {

	idle1 := GaugeValue("test1.mytest1", 10000)
	idle2 := GaugeValue("test1.mytest2", 10000)
	idle3 := GaugeValue("test1.mytest3", 10000)

	metrics := []*model.MetricValue{idle1, idle2, idle3}

	return metrics

}

func Test2() []*model.MetricValue {

	idle1 := GaugeValue("test2.mytest1", 1000)
	idle2 := GaugeValue("test2.mytest2", 2000)
	idle3 := GaugeValue("test2.mytest3", 3000)

	metrics := []*model.MetricValue{idle1, idle2, idle3}

	return metrics

}
