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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/open-falcon/falcon-plus/modules/api/app/utils"
	"github.com/open-falcon/falcon-plus/modules/api/config"
)

var db config.DBPool

const badstatus = http.StatusBadRequest
const expecstatus = http.StatusExpectationFailed

func Routes(r *gin.Engine) {

	db = config.Con()
	authapi := r.Group("/api/v1")
	noauthapi := r.Group("/api/v1")
	authapi.Use(utils.AuthSessionMidd)

	noauthapi.GET("/graph/endpointobj", EndpointObjGet)
	noauthapi.GET("/graph/endpoint", EndpointRegexpQuery)
	noauthapi.GET("/graph/endpoint_counter", EndpointCounterRegexpQuery)
	noauthapi.POST("/graph/history", QueryGraphDrawData)
	noauthapi.POST("/graph/lastpoint", QueryGraphLastPoint)
	noauthapi.POST("/graph/info", QueryGraphItemPosition)

	//authapi.DELETE("/graph/endpoint", DeleteGraphEndpoint)
	//authapi.DELETE("/graph/counter", DeleteGraphCounter)

	grfanaapi := r.Group("/api/v1")
	grfanaapi.GET("/grafana", GrafanaMainQuery)
	grfanaapi.GET("/grafana/metrics/find", GrafanaMainQuery)
	grfanaapi.POST("/grafana/render", GrafanaRender)
	grfanaapi.GET("/grafana/render", GrafanaRender)
}
