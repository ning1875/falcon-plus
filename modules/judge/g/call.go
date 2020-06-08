package g

import (
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
	cache"github.com/patrickmn/go-cache"
	"time"
)



var (
	GlobalCacheTtl = 60*2*time.Second
	HostSSCache            = cache.New(GlobalCacheTtl, GlobalCacheTtl)
)

func JudgeHostStrategies(hostname string) []*model.Strategy {
	if res, found := HostSSCache.Get(hostname); found {

		return res.([]*model.Strategy)
	} else {
		res := GetOneStrategy(hostname)

		HostSSCache.Set(hostname, res, GlobalCacheTtl)
		return res
	}
}


func CheckNeedJudge(hostname, metric string) (filterSS []*model.Strategy) {


	ss := JudgeHostStrategies(hostname)
	if len(ss) == 0 {
		return
	}
	for _, i := range ss {
		//g.Logger.Infof("metric_a:%s,metric_b:%s", i.Metric, metric)
		if i.Metric == metric {
			filterSS = append(filterSS, i)
		}
	}
	return

}

func GetOneStrategy(hostname string) []*model.Strategy {
	var res model.OneStrategiesResponse
	err := HbsClient.Call("Hbs.GetOneEndStrategies", model.SingleEndRpcRequest{Endpoint: hostname}, &res)
	if err != nil {
		g.Logger.Errorf(" Hbs.GetOneEndStrategies:", err)
		return nil
	}
	return res.Strategies
}

func GetOneUnionStrategy(id int) []*model.Strategy {
	g.Logger.Infof("GetOneUnionStrategy_id:%d", id)
	var res model.OneStrategiesResponse
	err := HbsClient.Call("Hbs.GetOneUnionStrategies", model.SingleIdRequest{Id: id}, &res)
	if err != nil {
		g.Logger.Errorf(" Hbs.GetOneUnionStrategies:", err)
		return nil
	}
	return res.Strategies
}
