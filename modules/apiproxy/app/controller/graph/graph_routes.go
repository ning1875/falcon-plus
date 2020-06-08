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
)

const badstatus = http.StatusBadRequest
const expecstatus = http.StatusExpectationFailed

func Routes(r *gin.Engine) {

	authapi := r.Group("/api/v1")
	authapi.GET("/graph/endpointobj", EndpointObjProxy)
	authapi.GET("/graph/endpoint", EndpointProxy)
	authapi.GET("/graph/endpoint_counter", EndpointCounterProxy)
	authapi.POST("/graph/history", QueryHistoryProxy)
	authapi.POST("/graph/lastpoint", QueryLastPointProxy)
	authapi.POST("/graph/info", QueryGraphInfoProxy)
	grfanaapi := r.Group("/api")
	grfanaapi.GET("/v1/grafana", GrafanaMainQueryProxy)
	grfanaapi.GET("/v1/grafana/metrics/find", GrafanaMainQueryProxy)
	grfanaapi.POST("/v1/grafana/render", GrafanaProxy)
	grfanaapi.GET("/v1/grafana/render", GrafanaProxy)
}
