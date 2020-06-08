package store

import (
	"sync"
	"time"

	"strconv"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/judge/g"
)

var UnionJudgeHistoryMap = SafeUnionJudgeHistoryMap{M: make(map[string]*UnionJudge)}

type SafeUnionJudgeHistoryMap struct {
	sync.RWMutex
	M map[string]*UnionJudge
}

func (this *SafeUnionJudgeHistoryMap) Exists(key string) bool {
	this.RLock()
	defer this.RUnlock()
	if _, ok := this.M[key]; ok {
		return true
	}
	return false
}

func (this *SafeUnionJudgeHistoryMap) Get(key string) (*UnionJudge, bool) {
	this.RLock()
	defer this.RUnlock()
	val, ok := this.M[key]
	return val, ok
}

func (this *SafeUnionJudgeHistoryMap) Set(key string, val *UnionJudge) {
	this.Lock()
	defer this.Unlock()
	this.M[key] = val
}

type UnionJudge struct {
	sync.RWMutex
	M map[string]*model.Event
}

func NewUnionJudgeEvent() *UnionJudge {
	return &UnionJudge{M: make(map[string]*model.Event)}
}

func (this *UnionJudge) Get(key string) (*model.Event, bool) {
	this.RLock()
	defer this.RUnlock()
	val, ok := this.M[key]
	return val, ok
}

func (this *UnionJudge) SetAndJudge(key string, event *model.Event) {
	this.Lock()
	defer this.Unlock()

	//sChain := g.UnionStrategyMap.GetById(val.Strategy.UnionStrategyId)
	unionId := event.Strategy.UnionStrategyId
	sChain := g.GetOneUnionStrategy(unionId)
	sNums := len(sChain)

	this.M[key] = event
	//g.Logger.Infof("config_len:%d ,now_len:%d sChain:%+v, map%+v", len(sChain), len(this.M), sChain, this.M)
	if sNums != len(this.M) {
		//	数据还不全，
		//g.Logger.Warningf("Union_judge_data_not_complete event:%+v", event)
		return
	}

	//now := time.Now().Unix()
	var eChain []*model.Event
	okNum := 0
	for keyIn, event := range this.M {
		// 根据event_time 和 step 的对比查看是否过期
		if event.Status == "OK" {
			okNum++
			//isStrategied = false
			//break
		}

		if key != keyIn {
			eChain = append(eChain, event)
		}
	}

	// 全部恢复或者全部触发
	event.EventChain = eChain
	now := time.Now().Unix()
	if okNum == len(sChain) {
		g.Logger.Infof("组合告警全部恢复:key:%s 组合id:%d sChain:%+v", key, unionId, sChain)
		//sendEventToRedis(event)
		g.UnionLastEvents.Set(strconv.Itoa(unionId), event)
		sendEvent(event)
		return
	}
	if okNum == 0 {

		lastEvent, exists := g.UnionLastEvents.Get(strconv.Itoa(unionId))

		if exists == false {
			g.Logger.Infof("组合告警第1次触发:key:%s 组合id:%d sChain:%+v", key, unionId, sChain)
			g.UnionLastEvents.Set(strconv.Itoa(unionId), event)
			sendEvent(event)

			return
		}
		//g.Logger.Infof("last_status:%s this_status:%s ", lastEvent.Status, event.Status)

		if exists == true && lastEvent.Status == "OK" {
			// 组合告警又再次触发了
			g.Logger.Infof("组合告警第1次触发:key:%s 组合id:%d sChain:%+v", key, unionId, sChain)
			event.CurrentStep = 1
			g.UnionLastEvents.Set(strconv.Itoa(unionId), event)
			sendEvent(event)

			return

		}
		if lastEvent.CurrentStep >= lastEvent.MaxStep() {
			// 报警次数已经足够多，到达了最多报警次数了，不再报警
			return
		}

		if now-lastEvent.EventTime < g.Config().Alarm.MinInterval {
			// 报警不能太频繁，两次报警之间至少要间隔MinInterval秒，否则就不能报警
			return
		}
		event.CurrentStep = lastEvent.CurrentStep + 1
		g.UnionLastEvents.Set(strconv.Itoa(unionId), event)
		g.Logger.Infof("组合告警第[%d]次触发:key:%s 组合id:%d sChain:%+v", event.CurrentStep, key, unionId, sChain)
		sendEvent(event)
		return
	}

}
