package cache

import (
	"strconv"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/db"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
	cache "github.com/patrickmn/go-cache"
)

var (
	HostMaintainCache      = cache.New(g.GlobalCacheTtl, g.GlobalCacheTtl)
	HostSSCache            = cache.New(g.GlobalCacheTtl, g.GlobalCacheTtl)
	UnionStrategiesCache   = cache.New(g.GlobalCacheTtl, g.GlobalCacheTtl)
	HostSpecialMetricCache = cache.New(g.GlobalCacheTtl, g.GlobalCacheTtl)
	HostPluginCache        = cache.New(g.GlobalCacheTtl, g.GlobalCacheTtl)
)

func JudgeHostMaintain(hostname string) bool {

	if res, found := HostMaintainCache.Get(hostname); found {
		return res.(bool)
	} else {
		res := db.QueryHostMaintain(hostname)

		HostMaintainCache.Set(hostname, res, g.GlobalCacheTtl)
		return res
	}

}

func JudgeHostStrategies(hostname string) []*model.Strategy {
	if res, found := HostSSCache.Get(hostname); found {

		return res.([]*model.Strategy)
	} else {
		res := db.QueryStrategiesForOneEnd(hostname)

		HostSSCache.Set(hostname, res, g.GlobalCacheTtl)
		return res
	}
}

func JudgeUnionStrategies(unionId int) []*model.Strategy {
	if res, found := UnionStrategiesCache.Get(strconv.Itoa(unionId)); found {

		return res.([]*model.Strategy)
	} else {
		res := db.QueryUnionStrategiesByUnionId(unionId)

		UnionStrategiesCache.Set(strconv.Itoa(unionId), res, g.GlobalCacheTtl)
		return res
	}
}

func JudgeHostSpecialMetric(hostname string) []*model.BuiltinMetric {
	if res, found := HostSpecialMetricCache.Get(hostname); found {

		return res.([]*model.BuiltinMetric)
	} else {
		res := db.QueryHostSpecialMetric(hostname)

		HostSpecialMetricCache.Set(hostname, res, g.GlobalCacheTtl)
		return res
	}
}

func JudgeHostPlugin(hostname string) []string {
	if res, found := HostPluginCache.Get(hostname); found {

		return res.([]string)
	} else {
		res := db.QueryPluginsForEnd(hostname)

		HostPluginCache.Set(hostname, res, g.GlobalCacheTtl)
		return res
	}
}
