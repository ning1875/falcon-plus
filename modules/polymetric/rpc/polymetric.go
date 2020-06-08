package rpc

import (
	"fmt"

	"time"

	log "github.com/Sirupsen/logrus"

	cmodel "github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/polymetric/cron"
)

type Polymetric int

type PolyRequest struct {
	PolyName string      `json:"poly_name"`
	Value    interface{} `json:"value"`
}

type PolymetricResp struct {
	Msg        string
	Total      int
	ErrInvalid int
	Latency    int64
}

func (t *PolymetricResp) String() string {
	s := fmt.Sprintf("TransferResp total=%d, err_invalid=%d, latency=%dms",
		t.Total, t.ErrInvalid, t.Latency)
	if t.Msg != "" {
		s = fmt.Sprintf("%s, msg=%s", s, t.Msg)
	}
	return s
}

func (this *Polymetric) Ping(req cmodel.NullRpcRequest, resp *cmodel.SimpleRpcResponse) error {
	return nil
}

func (t *Polymetric) Update(args []*cmodel.PolyRequest, reply *PolymetricResp) error {
	reply.ErrInvalid = 0
	if len(args) == 0 {
		reply.Msg = "zero item"
		return nil
	}
	for _, item := range args {
		// 非法的聚合规则
		T, loaded := cron.PolyWorkerQueueMap.Load(item.PolyName)
		log.Debugf("item:%+v", item, T, loaded)
		if loaded == false {
			reply.ErrInvalid++
			continue
		}
		//var TickerStep int
		//if item.PolyName == cron.CounterType {
		//	TickerStep = cron.CounterTimeStep
		//} else {
		//	TickerStep = cron.PolyTimeStep
		//}

		worker := T.(*cron.PolyTickerWorker)

		if _, exist := cron.PolyTypeMap.Load(item.PolyName); exist == false {
			cron.PolyTypeMap.Store(item.PolyName, true)
			worker.Started = true
			worker.Ticker = time.NewTicker(time.Duration(cron.PolyTimeStep) * time.Second)
			worker.Start()
		}

		worker.Queue.PushFront(item)
		reply.Total++
	}
	return nil
}
