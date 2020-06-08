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

package sdk

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"errors"

	cmodel "github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/common/sdk/requests"
	"github.com/open-falcon/falcon-plus/modules/aggregator/g"
	cache "github.com/patrickmn/go-cache"
	"github.com/toolkits/net/httplib"
)

var (
	//一层cache 为了缓解db的压力 group_id :  []endpoint_list
	DbHostGroupCache = cache.New(g.CacheTTl, g.CacheTTl)
)

func HostnamesByID(groupId int64) []string {
	gidStr := strconv.Itoa(int(groupId))
	if res, found := DbHostGroupCache.Get(gidStr); found {
		return res.([]string)
	} else {
		res, err := CurlHostnamesByID(groupId)
		if err == nil {
			DbHostGroupCache.Set(gidStr, res, cache.DefaultExpiration)
		}
		return res
	}

}

func CurlHostnamesByID(groupId int64) (hostnames []string, err error) {
	retryLimit := 3
	rS := 0
	for rS < retryLimit {
		hostnamesTmp, err := DoHttpHostnamesByID(groupId)
		if err != nil {
			rS += 1
			time.Sleep(time.Second)
		} else {
			hostnames = hostnamesTmp
			break
		}
	}
	err = errors.New("CurlHostnamesByID failed")
	return
}

func DoHttpHostnamesByID(groupId int64) ([]string, error) {
	uri := fmt.Sprintf("%s/api/v1/hostgroupsimple/%d", g.Config().Api.PlusApi, groupId)
	req, err := requests.CurlPlus(uri, "GET", "aggregator", g.Config().Api.PlusApiToken,
		map[string]string{}, map[string]string{})

	if err != nil {
		log.Println("[E] get group_id from api  hostgroupsimple err ", groupId)
		return []string{}, err
	}

	req.SetTimeout(time.Duration(g.Config().Api.ConnectTimeout)*time.Millisecond,
		time.Duration(g.Config().Api.RequestTimeout)*time.Millisecond)

	type RESP struct {
		Hosts []string `json:"hosts"`
	}

	resp := &RESP{}
	//var resp map[string][]string
	err = req.ToJson(&resp)
	if err != nil {
		return []string{}, err
	}

	return resp.Hosts, nil
}

func QueryLastPoints(endpoints, counters []string) (resp []*cmodel.GraphLastResp, err error) {
	cfg := g.Config()
	uri := fmt.Sprintf("%s/api/v1/graph/lastpoint", cfg.Api.PlusApi)

	var req *httplib.BeegoHttpRequest
	headers := map[string]string{"Content-type": "application/json"}
	req, err = requests.CurlPlus(uri, "POST", "aggregator", cfg.Api.PlusApiToken,
		headers, map[string]string{})

	if err != nil {

		return
	}

	req.SetTimeout(time.Duration(cfg.Api.ConnectTimeout)*time.Millisecond,
		time.Duration(cfg.Api.RequestTimeout)*time.Millisecond)

	body := []*cmodel.GraphLastParam{}
	for _, e := range endpoints {
		for _, c := range counters {
			body = append(body, &cmodel.GraphLastParam{e, c})
		}
	}

	b, err := json.Marshal(body)
	if err != nil {
		return
	}

	req.Body(b)

	err = req.ToJson(&resp)
	if err != nil {
		return
	}

	return resp, nil
}
