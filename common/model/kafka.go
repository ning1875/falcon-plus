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

package model

import (
	"encoding/json"
	"fmt"
	"log"
)

type KafkaData struct {
	Endpoint    string             `json:"endpoint"`
	Timestamp   int64              `json:"timestamp"`
	MetricValue map[string]float64 `json:"metric_values"`
	//Metric      string            `json:"metric"`
	//Step        int64             `json:"step"`
	//Value       float64           `json:"value"`
	//CounterType string            `json:"counterType"`
	//Tags        map[string]string `json:"tags"`
}

func (t *KafkaData) String() string {
	jsonStr, err := json.Marshal(t.MetricValue)
	if err != nil {
		log.Println("err:metrics")
		return ""
	}
	return fmt.Sprintf("{\"endpoint\": \"%s\", \"timestamp\":%d, \"metrics\": %s}",
		t.Endpoint, t.Timestamp, jsonStr)
}
