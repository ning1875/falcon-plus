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

package http

import (
	"fmt"
	"net/http"

	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/hbs/cache"
)

func configProcRoutes() {
	http.HandleFunc("/expressions", func(w http.ResponseWriter, r *http.Request) {
		RenderDataJson(w, cache.ExpressionCache.Get())
	})

	http.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		RenderDataJson(w, cache.Agents.Keys())
	})

	http.HandleFunc("/agentversions", func(w http.ResponseWriter, r *http.Request) {
		//RenderDataJson(w, cache.Agents.AgentVersions())
		//RenderDataJson(w, cache.Agents.M)
		tmpMap := make(map[string]string)
		cache.NowAgentVersionMap.Range(func(k, v interface{}) bool {
			tmpMap[k.(string)] = v.(string)
			return true
		})
		RenderDataJson(w, tmpMap)
	})

	http.HandleFunc("/hosts", func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]*model.Host, len(cache.MonitoredHosts.Get()))
		for k, v := range cache.MonitoredHosts.Get() {
			data[fmt.Sprint(k)] = v
		}
		RenderDataJson(w, data)
	})

	http.HandleFunc("/strategies", func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]*model.Strategy, len(cache.Strategies.GetMap()))
		for k, v := range cache.Strategies.GetMap() {
			data[fmt.Sprint(k)] = v
		}
		RenderDataJson(w, data)
	})

	http.HandleFunc("/templates", func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]*model.Template, len(cache.TemplateCache.GetMap()))
		for k, v := range cache.TemplateCache.GetMap() {
			data[fmt.Sprint(k)] = v
		}
		RenderDataJson(w, data)
	})

	http.HandleFunc("/plugins/", func(w http.ResponseWriter, r *http.Request) {
		hostname := r.URL.Path[len("/plugins/"):]
		RenderDataJson(w, cache.GetPlugins(hostname))
	})
	//发布升级任务
	http.HandleFunc("/agent/upgrade", func(w http.ResponseWriter, r *http.Request) {
		//取消升级
		if r.Method == "DELETE" {
			var args model.AgentUpgradeArgs
			cache.NewAgentUpgradeArgs = &args
			RenderDataJson(w, "取消升级成功")
			return
		}

		if r.Method != "POST" {
			http.Error(w, "发布升级任务只允许POST操作", http.StatusMethodNotAllowed)
			return
		}
		if strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
			var args model.AgentUpgradeArgs
			body, _ := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			//bodystr := string(body)
			if err := json.Unmarshal(body, &args); err != nil {
				http.Error(w, "解析出错", http.StatusBadRequest)
				return
			}
			if args.WgetUrl == "" || args.Version == "" {
				http.Error(w, "参数错误", http.StatusBadRequest)
				return
			}
			switch args.Type {
			case 0:
				if args.BinFileMd5 == "" {
					http.Error(w, "只升级bin:缺失BinFileMd5参数", http.StatusBadRequest)
					return
				}
			case 1:
				if args.CfgFileMd5 == "" {
					http.Error(w, "只升级cfg:缺失CfgFileMd5参数", http.StatusBadRequest)
					return
				}
			case 2:
				if args.BinFileMd5 == "" || args.CfgFileMd5 == "" {
					http.Error(w, "升级bin和cfg:缺失BinFileMd5或CfgFileMd5参数", http.StatusBadRequest)
					return
				}
			}

			cache.NewAgentUpgradeArgs = &args
			//显示成功
			RenderDataJson(w, cache.NewAgentUpgradeArgs)
		} else {
			http.Error(w, "no privilege", http.StatusForbidden)
			return
		}
	})
	//获取当前升级的参数
	http.HandleFunc("/agent/upgrade/nowargs", func(w http.ResponseWriter, r *http.Request) {
		RenderDataJson(w, cache.NewAgentUpgradeArgs)
	})
	//取消升级

}
