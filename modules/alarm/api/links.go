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

package api

import (
	"encoding/json"
	"fmt"
	"time"

	"errors"

	"io/ioutil"
	"net/http"

	"bytes"

	"github.com/open-falcon/falcon-plus/modules/alarm/g"
)

func LinkToSMS(content string) (string, error) {
	return HttpLink(content)
}

func HttpLink(content string) (string, error) {
	retryLimit := 3
	p := 0
	for p < retryLimit {
		uri := fmt.Sprintf("%s/api/v1/alarm/link", g.Config().Api.MainApi)
		data := make(map[string]interface{})
		data["content"] = content
		bytesData, err := json.Marshal(data)
		if err != nil {
			continue
		}
		reader := bytes.NewReader(bytesData)
		request, err := http.NewRequest("POST", uri, reader)
		if err != nil {
			continue
		}
		request.Header.Set("Content-Type", "application/json;charset=UTF-8")
		token, _ := json.Marshal(map[string]string{
			"name": "falcon-alarm",
			"sig":  g.Config().Api.PlusApiToken,
		})
		request.Header.Set("Apitoken", string(token))

		client := http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			continue
		}
		respBytes, err := ioutil.ReadAll(resp.Body)
		type res struct {
			Message string `json:"message"`
		}
		var rEs res
		err = json.Unmarshal(respBytes, &rEs)
		if err != nil {
			p += 1
			time.Sleep(time.Second)
		} else {
			return rEs.Message, nil
		}
	}

	return "", errors.New("HttpLink_finally_failed")
}
