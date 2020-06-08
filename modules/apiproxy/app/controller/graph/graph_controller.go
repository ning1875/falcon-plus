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

package graph

import (
	"encoding/json"
	"time"

	"bytes"
	"io/ioutil"
	"net/http"

	"fmt"

	"net/url"

	"strconv"

	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	cmodel "github.com/open-falcon/falcon-plus/common/model"
	graphApi "github.com/open-falcon/falcon-plus/modules/api/app/controller/graph"
	h "github.com/open-falcon/falcon-plus/modules/apiproxy/app/helper"
	"github.com/open-falcon/falcon-plus/modules/apiproxy/proxy"
	tcache "github.com/toolkits/cache/localcache/timedcache"
)

var (
	localStepCache = tcache.New(600*time.Second, 60*time.Second)
)

type EndPoint struct {
	Id       uint   `json:"id"`
	Endpoint string `json:"endpoint"`
}

type EndpointCounterQuery struct {
	Eid         string `json:"eid"`
	MetricQuery string `json:"metricQuery"`
	Limit       int    `json:"limit"`
	Page        int    `json:"page"`
}

type EndpointObj struct {
	Id       int    `json:"id"`
	Endpoint string `json:"endpoint"`
	Ts       int    `json:"ts"`
}

func CreateApiMap(hostnames []string) (apiMap map[string][]string) {
	apiMap = make(map[string][]string)
	for _, h := range hostnames {
		var found bool
		for region, v := range proxy.AgentConfigMap {
			if ok := proxy.AgentDataMap.JudgeRegion(region, h); ok {
				found = true
				apiAddr := v.Cfg.ApiAddr
				if agentList, get := apiMap[apiAddr]; get {
					agentList = append(agentList, h)
					apiMap[apiAddr] = agentList
				} else {
					tmpList := []string{h}
					apiMap[apiAddr] = tmpList
				}
				break
			}
		}
		//在所有的map中都没找到,需要把这个host添加到所有api队列中
		if found == false {
			for _, apiAddr := range proxy.ApiAddrList {
				if agentList, got := apiMap[apiAddr]; got {
					agentList = append(agentList, h)
					apiMap[apiAddr] = agentList
				} else {
					tmpList := []string{h}
					apiMap[apiAddr] = tmpList
				}
			}
		}
	}
	return
}

func QueryHistoryProxy(c *gin.Context) {

	var inputs graphApi.APIQueryGraphDrawData
	var err error
	if err = c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	//{langf:[1,2,3],huailai:[4,5,6]}
	//AgentConfigMap{langfang:http:lanfangapi,huai}

	respData := []*cmodel.GraphQueryResponseProxy{}
	apiMap := CreateApiMap(inputs.HostNames)
	//log.Infof("QueryHistoryProxy_APiMap", apiMap)
	dataMap := make(map[string]*cmodel.GraphQueryResponseProxy)
	for apiAddr, newHosts := range apiMap {
		inputs.HostNames = newHosts
		newAddr := fmt.Sprintf("%s%s", apiAddr, c.Request.URL.Path)
		if resp := BuildNewHttpPostReq(newAddr, inputs, c.Request.Header); resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Errorf("QueryHistory_http_error:url:%+v,header:%+v,param:%+v,error:%+v", newAddr, c.Request.Header, inputs, resp.Status)
				continue
			}
			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err.Error())
			}
			var resV []*cmodel.GraphQueryResponseProxy
			err = json.Unmarshal(respBytes, &resV)
			if err != nil {
				log.Error(err)
				continue
			}
			if len(resV) == 0 {
				continue
			}
			for _, d := range resV {
				if len(d.Values) > 0 {
					dataMap[fmt.Sprintf("%s_%s", d.Endpoint, d.Counter)] = d

				}
			}
		}
	}
	for _, v := range dataMap {
		respData = append(respData, v)
	}

	h.JSONR(c, respData)
}

func QueryLastPointProxy(c *gin.Context) {
	var inputs []cmodel.GraphLastParam
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	respData := []*cmodel.GraphLastResp{}

	apiMap := make(map[string][]cmodel.GraphLastParam)

	for _, param := range inputs {
		var found bool
		for region, v := range proxy.AgentConfigMap {
			if ok := proxy.AgentDataMap.JudgeRegion(region, param.Endpoint); ok {
				found = true
				apiAddr := v.Cfg.ApiAddr
				if agentList, get := apiMap[apiAddr]; get {
					agentList = append(agentList, param)
					apiMap[apiAddr] = agentList
				} else {
					tmpList := []cmodel.GraphLastParam{param}
					apiMap[apiAddr] = tmpList
				}
				break
			}
		}
		//在所有的map中都没找到,需要把这个host添加到所有api队列中
		if found == false {
			for _, apiAddr := range proxy.ApiAddrList {
				if agentList, got := apiMap[apiAddr]; got {
					agentList = append(agentList, param)
					apiMap[apiAddr] = agentList
				} else {
					tmpList := []cmodel.GraphLastParam{param}
					apiMap[apiAddr] = tmpList
				}
			}
		}

	}
	//log.Infof("QueryLastProxy_APiMap", apiMap)
	dataMap := make(map[string]*cmodel.GraphLastResp)
	for apiAddr, param := range apiMap {
		newAddr := fmt.Sprintf("%s%s", apiAddr, c.Request.URL.Path)
		resp := BuildNewHttpPostReq(newAddr, param, c.Request.Header)
		if resp == nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Errorf("QueryLastPoint_http_error:url:%+v,header:%+v,param:%+v,error:%+v", newAddr, c.Request.Header, param, resp.Status)
			//h.JSONR(c, resp.Status)
			continue
		}
		respBytes, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			log.Error(err.Error())
		}
		var resV []*cmodel.GraphLastResp
		err = json.Unmarshal(respBytes, &resV)
		if err != nil {
			log.Error(err)
			continue
		}
		if len(resV) == 0 {
			continue
		}

		for _, d := range resV {
			//log.Infof("resV:%+v,dataMap:%+v", d.Value, dataMap, apiAddr, param)
			// ts=0 代表没取到数据
			if d.Value.Timestamp == 0 {
				continue
			}
			//respData = append(respData, d)
			//h.JSONR(c, respData)
			//return
			dataMap[fmt.Sprintf("%s_%s", d.Endpoint, d.Counter)] = d
		}

	}
	for _, v := range dataMap {
		respData = append(respData, v)
	}
	h.JSONR(c, respData)
}

func QueryGraphInfoProxy(c *gin.Context) {
	var inputs cmodel.GraphInfoParam
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	hostnames := []string{inputs.Endpoint}
	apiMap := CreateApiMap(hostnames)
	//log.Infof("QueryGraphInfoProxy_APiMap", apiMap)
	var resV *cmodel.GraphFullyInfo
	for apiAddr, newHosts := range apiMap {
		inputs.Endpoint = newHosts[0]
		newAddr := fmt.Sprintf("%s%s", apiAddr, c.Request.URL.Path)
		if resp := BuildNewHttpPostReq(newAddr, inputs, c.Request.Header); resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Errorf("QueryGraphInfoProxy_http_error:url:%+v,header:%+v,param:%+v,error:%+v", newAddr, c.Request.Header, inputs, resp.Status)
				//h.JSONR(c, resp.Status)
				continue
			}
			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			err = json.Unmarshal(respBytes, &resV)
			if err != nil {
				log.Error(err)
				continue
			}
			h.JSONR(c, resV)
			return
		}
	}

}

func GrafanaProxy(c *gin.Context) {
	inputs := graphApi.APIGrafanaRenderInput{}
	//set default step is 60
	inputs.Step = 60
	inputs.ConsolFun = "AVERAGE"
	if err := c.Bind(&inputs); err != nil {
		log.Errorf("GrafanaRender", c.Request, err.Error())
		h.JSONR(c, badstatus, err.Error())
		return
	}
	respList := []*cmodel.GraphQueryResponseProxy{}
	apiMap := make(map[string][]string)
	for _, target := range inputs.Target {
		// target: n8-037-204,n8-037-205#cpu#idle
		hostnames, _, counterTmp := graphApi.CutEndpointCounterHelp(target)
		for _, h := range hostnames {
			originTarget := fmt.Sprintf("%s#%s", h, counterTmp)
			var found bool
			for region, v := range proxy.AgentConfigMap {
				if ok := proxy.AgentDataMap.JudgeRegion(region, h); ok {
					found = true
					apiAddr := v.Cfg.ApiAddr
					if targetList, get := apiMap[apiAddr]; get {

						targetList = append(targetList, originTarget)
						apiMap[apiAddr] = targetList
					} else {
						targetList := []string{originTarget}
						apiMap[apiAddr] = targetList
					}
					break
				}
			}
			//在所有的map中都没找到,需要把这个host添加到所有api队列中
			if found == false {
				for _, apiAddr := range proxy.ApiAddrList {
					if targetList, got := apiMap[apiAddr]; got {
						targetList = append(targetList, originTarget)
						apiMap[apiAddr] = targetList
					} else {
						targetList := []string{originTarget}
						apiMap[apiAddr] = targetList
					}
				}
			}
		}
	}
	//log.Infof("GrafanaProxy_APiMap", apiMap)

	dataMap := make(map[string]*cmodel.GraphQueryResponseProxy)
	for apiAddr, targets := range apiMap {
		inputs.Target = targets
		newAddr := fmt.Sprintf("%s%s", apiAddr, c.Request.URL.Path)
		//resp := BuildNewHttpPostReq(newAddr, inputs, c.Request.Header)
		data := url.Values{
			"target":        inputs.Target,
			"from":          {strconv.FormatInt(inputs.From, 10)},
			"until":         {strconv.FormatInt(inputs.Until, 10)},
			"format":        {inputs.Format},
			"maxDataPoints": {strconv.FormatInt(inputs.MaxDataPoints, 10)},
			"step":          {strconv.Itoa(inputs.Step)},
			"consolFun":     {inputs.ConsolFun}}
		//log.Infof("GrafanaProxy_data", data)
		resp, err := http.PostForm(newAddr, data)
		if err != nil {
			log.Errorf("GrafanaProxy_http_error:url:%+v,header:%+v,param:%+v,error:%+v", newAddr, c.Request.Header, inputs, resp.Status)
			log.Errorf("GrafanaProxy_http_error:resp.Body:%+v", resp.Body)
			continue
		}
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
		}
		var resV []*cmodel.GraphQueryResponseProxy
		err = json.Unmarshal(respBytes, &resV)
		if err != nil {
			log.Error(err)
			continue
		}
		if len(resV) == 0 {
			continue
		}
		// 使用endpoint + counter 做去重
		for _, d := range resV {
			dataMap[fmt.Sprintf("%s_%s", d.Endpoint, d.Counter)] = d
		}
	}
	for _, d := range dataMap {
		respList = append(respList, d)
	}
	c.JSON(200, respList)
	return
}

func EndpointObjProxy(c *gin.Context) {
	inputs := graphApi.APIEndpointObjGetInputs{
		Deadline: 0,
	}
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	if len(inputs.Endpoints) == 0 {
		h.JSONR(c, http.StatusBadRequest, "endpoints missing")
		return
	}
	apiMap := CreateApiMap(inputs.Endpoints)
	dataMap := make(map[string]EndpointObj)
	for apiAddr, newHosts := range apiMap {
		newAddr := fmt.Sprintf("%s%s", apiAddr, c.Request.URL.Path)
		req, err := http.NewRequest("GET", newAddr, nil)
		if err != nil {
			log.Errorf("GrafanaGet_http_error_1:url:%s,%+v", newAddr, err)
			continue
		}
		req.Header = c.Request.Header
		q := req.URL.Query()
		for _, e := range newHosts {
			q.Add("endpoints", e)
		}
		q.Add("deadline", strconv.FormatInt(inputs.Deadline, 10))
		req.URL.RawQuery = q.Encode()
		var resp *http.Response
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Errorf("GrafanaGet_http_error_2:url:%s,err:%+v", newAddr, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Errorf("GrafanaGet_http_error_3:url:%s,status:%+v", newAddr, resp.Status)
			continue
		}
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		defer resp.Body.Close()
		var output []EndpointObj
		err = json.Unmarshal(respBytes, &output)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		for _, e := range output {
			dataMap[e.Endpoint] = e
		}

	}

	endpoints := []map[string]interface{}{}
	for _, r := range dataMap {
		endpoints = append(endpoints, map[string]interface{}{"id": r.Id, "endpoint": r.Endpoint, "ts": r.Ts})
	}

	h.JSONR(c, endpoints)
}

func EndpointProxy(c *gin.Context) {
	inputs := graphApi.APIEndpointRegexpQueryInputs{
		//set default is 500
		Limit: 500,
		Page:  1,
	}
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	if inputs.Q == "" && inputs.Label == "" {
		h.JSONR(c, http.StatusBadRequest, "q and labels are all missing")
		return
	}

	qs := []string{}
	if inputs.Q != "" {
		qs = strings.Split(inputs.Q, " ")
	}

	apiMap := CreateApiMap(qs)
	endpoints := []map[string]interface{}{}
	endMap := make(map[string]uint)
	for apiAddr, newHosts := range apiMap {
		inputs.Q = strings.Join(newHosts, " ")
		newAddr := fmt.Sprintf("%s%s", apiAddr, c.Request.URL.Path)
		req, err := http.NewRequest("GET", newAddr, nil)
		if err != nil {
			log.Errorf("EndpointProxy_http_error_1:url:%s,%+v", apiAddr, err)
			continue
		}
		req.Header = c.Request.Header
		q := req.URL.Query()
		q.Add("q", inputs.Q)
		q.Add("tags", inputs.Label)
		q.Add("limit", strconv.Itoa(inputs.Limit))
		q.Add("page", strconv.Itoa(inputs.Page))
		req.URL.RawQuery = q.Encode()
		var resp *http.Response
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Errorf("EndpointProxy_http_error_2:url:%s,err:%+v", apiAddr, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Errorf("EndpointProxy_http_error_3:url:%s,status:%+v", apiAddr, resp.Status)
			continue
		}

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
		}
		defer resp.Body.Close()

		var endpoint []*EndPoint
		err = json.Unmarshal(respBytes, &endpoint)
		if err != nil {
			log.Error(err)
			continue
		}

		for _, e := range endpoint {
			endMap[e.Endpoint] = e.Id
		}
	}
	for e, id := range endMap {
		endpoints = append(endpoints, map[string]interface{}{"id": id, "endpoint": e})
	}
	h.JSONR(c, endpoints)

}

type EndCounter struct {
	EndpointID int    `json:"endpoint_id"`
	Counter    string `json:"counter"`
	Step       int    `json:"step"`
	Type       string `json:"type"`
}

func EndpointCounterProxy(c *gin.Context) {
	eid := c.DefaultQuery("eid", "")
	metricQuery := c.DefaultQuery("metricQuery", ".+")
	limitTmp := c.DefaultQuery("limit", "500")
	limit, err := strconv.Atoi(limitTmp)
	if err != nil {
		h.JSONR(c, http.StatusBadRequest, err)
		return
	}
	pageTmp := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageTmp)

	if err != nil {
		h.JSONR(c, http.StatusBadRequest, err)
		return
	}
	countersResp := []interface{}{}

	dataMap := make(map[string]*EndCounter)
	for _, apiAddr := range proxy.ApiAddrList {
		newAddr := fmt.Sprintf("%s%s", apiAddr, c.Request.URL.Path)
		req, err := http.NewRequest("GET", newAddr, nil)
		if err != nil {
			log.Errorf("EndpointProxy_http_error_1:url:%s,%+v", apiAddr, err)
			continue
		}
		req.Header = c.Request.Header
		q := req.URL.Query()
		q.Add("eid", eid)
		q.Add("metricQuery", metricQuery)
		q.Add("limit", strconv.Itoa(limit))
		q.Add("page", strconv.Itoa(page))
		req.URL.RawQuery = q.Encode()
		var resp *http.Response
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Errorf("EndpointProxy_http_error_2:url:%s,err:%+v", apiAddr, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Errorf("EndpointProxy_http_error_3:url:%s,status:%+v", apiAddr, resp.Status)
			continue
		}

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
		}
		defer resp.Body.Close()
		var counters []*EndCounter
		err = json.Unmarshal(respBytes, &counters)
		if err != nil {
			log.Error(err)
		}
		if len(counters) == 0 {
			continue
		}
		for _, c := range counters {
			dataMap[fmt.Sprintf("%s_%s", c.EndpointID, c.Counter)] = c
		}

	}
	for _, d := range dataMap {
		countersResp = append(countersResp, map[string]interface{}{
			"endpoint_id": d.EndpointID,
			"counter":     d.Counter,
			"step":        d.Step,
			"type":        d.Type,
		})
	}
	h.JSONR(c, countersResp)

}

func GrafanaGet(uri string, header http.Header, input graphApi.APIGrafanaMainQueryInputs) []graphApi.APIGrafanaMainQueryOutputs {
	var outputs []graphApi.APIGrafanaMainQueryOutputs
	for _, apiAddr := range proxy.ApiAddrList {
		newAddr := fmt.Sprintf("%s%s", apiAddr, uri)
		req, err := http.NewRequest("GET", newAddr, nil)
		if err != nil {
			log.Errorf("GrafanaGet_http_error_1:url:%s,%+v", newAddr, err)
			continue
		}
		req.Header = header
		q := req.URL.Query()
		q.Add("query", input.Query)
		q.Add("limit", strconv.Itoa(input.Limit))
		req.URL.RawQuery = q.Encode()
		var resp *http.Response
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Errorf("GrafanaGet_http_error_2:url:%s,err:%+v", newAddr, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Errorf("GrafanaGet_http_error_3:url:%s,status:%+v", newAddr, resp.Status)
			continue
		}

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
		}
		defer resp.Body.Close()
		var output []graphApi.APIGrafanaMainQueryOutputs
		err = json.Unmarshal(respBytes, &output)
		if err != nil {
			log.Error(err)
			continue
		}
		if len(output) == 0 {
			continue
		}
		outputs = append(outputs, output...)
	}

	return outputs
}

func GrafanaMainQueryProxy(c *gin.Context) {
	inputs := graphApi.APIGrafanaMainQueryInputs{}
	inputs.Limit = 1000
	inputs.Query = "!N!"
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err.Error())
		return
	}
	//log.Debugf("got query string: %s", inputs.Query)
	output := GrafanaGet(c.Request.URL.Path, c.Request.Header, inputs)
	c.JSON(200, output)
	return
}

func BuildNewHttpPostReq(url string, data interface{}, header http.Header) *http.Response {
	bytesData, err := json.Marshal(data)
	if err != nil {
		log.Errorf("url:%+v_BuildNewHttpPostReq__json.Marshal_data:%+v,error:%+v", url, data, err)
		return nil
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	// http短连接
	//request.Close = true
	if err != nil {
		//log.Errorf("BuildNewHttpPostReq_ERROR[1]_[url:%+v]__[data:%+v]__[error:%+v]", url, data, err)
		log.Errorf("BuildNewHttpPostReq_ERROR[1]_[url:%+v]__[data]__[error:%+v]", url, err)
		return nil
	}
	request.Header = header
	client := http.Client{}
	client.Timeout = time.Duration(time.Second * 5)
	resp, err := client.Do(request)
	if err != nil {
		//log.Errorf("BuildNewHttpPostReq_ERROR[2]_[url:%+v]__[data:%+v]__[error:%+v]", url, data, err)
		log.Errorf("BuildNewHttpPostReq_ERROR[2]_[url:%+v]__[data]__[error:%+v]", url, err)
		return nil
	}
	return resp

}
