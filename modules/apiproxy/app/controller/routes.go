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

package controller

import (
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/open-falcon/falcon-plus/modules/apiproxy/app/controller/graph"
	"github.com/spf13/viper"
)

func StartGin(port string, r *gin.Engine) {
	//r.Use(utils.CORS())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, I'm FalconProxy+ (｡A｡)")
	})
	graph.Routes(r)
	//uic.Routes(r)
	//r.Run(port)
	s := &http.Server{
		Addr:           port,
		Handler:        r,
		ReadTimeout:    time.Duration(viper.GetInt("read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(viper.GetInt("write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
