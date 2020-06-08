package proxy

import (
	"encoding/json"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/model"
	AgentG "github.com/open-falcon/falcon-plus/modules/agent/g"
)

var AgentConfigMap = make(map[string]*ApiAgent)
var ApiAddrList []string
var AgentDataMap = &SafeHostMap{M: make(map[string]map[string]int)}

type SafeHostMap struct {
	sync.RWMutex
	M map[string]map[string]int
}

type ApiAgent struct {
	Cfg        ApiConfig
	RpcClients []*AgentG.SingleConnRpcClient
}

type ApiConfig struct {
	HbsRpcAddrs    []string `json:"hbs_rpc_addrs"`
	RpcCallTimeout int      `json:"rpc_call_timeout"`
	ApiAddr        string   `json:"api_addr"`
}

func (this *SafeHostMap) ReInit(m map[string]map[string]int) {
	this.Lock()
	defer this.Unlock()
	this.M = m
}

func (this *SafeHostMap) JudgeRegion(region, hostname string) bool {
	this.RLock()
	defer this.RUnlock()
	if data, ok := this.M[region]; ok {
		if _, okIn := data[hostname]; okIn {
			return true
		}
	}
	return false
}

func InitProxy(apiMap map[string]interface{}) {
	for k, v := range apiMap {
		bytes, _ := json.Marshal(v)
		var p ApiConfig
		_ = json.Unmarshal(bytes, &p)
		var agent ApiAgent
		var rpcS []*AgentG.SingleConnRpcClient
		agent.Cfg = p
		for _, addr := range p.HbsRpcAddrs {
			rpcS = append(rpcS, &AgentG.SingleConnRpcClient{
				RpcServer: addr,
				Timeout:   time.Duration(p.RpcCallTimeout) * time.Millisecond,
			})
		}
		agent.RpcClients = rpcS
		ApiAddrList = append(ApiAddrList, p.ApiAddr)
		AgentConfigMap[k] = &agent
	}
}

func SyncHost() {
	for {
		SyncHostOnce()
		time.Sleep(time.Duration(5) * time.Minute)
	}
}

func SyncHostOnce() {
	m := make(map[string]map[string]int)
	for region, agent := range AgentConfigMap {

		innerMap := make(map[string]int)
		for _, ag := range agent.RpcClients {
			var res model.Agents
			err := ag.Call("Hbs.GetAgents", model.NullRpcRequest{}, &res)
			if err != nil {
				log.Errorf("Hbs.GetAgents_error, server:%+v,error:%+v", ag.RpcServer, err)
				continue
			}
			if len(res.Agent) == 0 {
				continue
			}
			for _, tmp := range res.Agent {
				innerMap[tmp] = 0
			}

		}
		log.Infof("SyncHostOnce_result:reigon:%s num:%d", region, len(innerMap))
		m[region] = innerMap

	}
	AgentDataMap.ReInit(m)
}
