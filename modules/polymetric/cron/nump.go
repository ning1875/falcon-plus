package cron

import (
	"time"

	"bytes"
	ojson "encoding/json"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/rpc/json"
	"github.com/open-falcon/falcon-plus/modules/polymetric/g"
)

type NumpReq struct {
	Metric    string             `json:"metric"`
	PolyName  string             `json:"poly_name"`
	TimeStamp int64              `json:"time_stamp"`
	ValueDict map[float64]string `json:"value_dict"`
}

type NumpReqNew struct {
	Metric    string      `json:"metric"`
	PolyName  string      `json:"poly_name"`
	TimeStamp int64       `json:"time_stamp"`
	ValueList []SingleEnd `json:"value_list"`
}

func RpcCallNumpApi(metric, polyName string, endList []SingleEnd) {

	if g.Config().NumpRpc.Enabled == false {
		log.Infof("[RpcCallNumpApi]disabled ")
		return
	}

	Nr := NumpReqNew{
		Metric:    metric,
		PolyName:  polyName,
		TimeStamp: time.Now().Unix(),
		ValueList: endList,
	}

	url := g.Config().NumpRpc.Url
	message, err := json.EncodeClientRequest("gen_cluster_normal_dis", Nr)

	if err != nil {
		log.Errorf("json.EncodeClientRequest error:%s", err)
		return

	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(message))
	if err != nil {
		return
	}
	if resp.Body == nil {
		return
	}
	defer resp.Body.Close()
	if err != nil {
		log.Errorf("http.Post error:%s", err)
		return

	}

	type callRes struct {
		Id      int    `json:"id"`
		Result  string `json:"result"`
		Jsonrpc string `json:"jsonrpc"`
	}
	CallRes := &callRes{}
	//err = json.DecodeClientResponse(resp.Body, CallRes)
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("http.Post error:%s", err)
		return

	}

	err = ojson.Unmarshal(respBytes, CallRes)
	if err != nil {
		log.Errorf("ojson.Unmarshal error:%s", err)

	}

}
