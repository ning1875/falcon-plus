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

package funcs

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"math"
	"regexp"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
)

func NtpMetricsOld() (L []*model.MetricValue) {
	commandName := "/usr/bin/ntpq"
	params := []string{"-pn"}
	//ntpq -pn 输出
	//remote           refid      st t when poll reach   delay   offset  jitter
	//==============================================================================
	//69.60.114.223   128.227.205.3    2 u 1457 1024    6  212.418   -2.230  19.148
	//193.228.143.26  .STEP.          16 u 110d 1024    0  348.473   -5.384   0.000
	//172.104.71.235  255.254.0.28     2 u  85d 1024    0  118.124  -10.181   0.000
	//120.25.108.11   10.137.53.7      2 u 1312 1024  142   36.175    3.918   8.122
	//*10.4.16.34      85.199.214.100   2 u  825 1024  377    2.405    2.806  13.215

	cmd := exec.Command(commandName, params...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("/usr/bin/ntpq -pn exec  error:%+v", err)
		return nil
	}

	cmd.Start()
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			return nil
		}
		//*当前作为优先同步对象的远程节点
		if strings.HasPrefix(line, "*") {
			fields := strings.Fields(line)
			// 第八行是offset
			ntpOffset, parserror := strconv.ParseFloat(fields[8], 64)
			if parserror != nil {
				log.Printf("parseunit error:%+v", parserror)
				return nil
			}
			L = append(L, GaugeValue("sys.ntp.offset", ntpOffset))
		}

	}
	cmd.Wait()
	return
}

func chronycParseOffset(parseStr string) (error, float64) {
	// parseStr = -4516ns or -11us or -118ns
	s := []byte(parseStr)
	parten := `([\+|\-][\d]+)([a-z]{1,2})`
	reg := regexp.MustCompile(parten)
	// 0: +15us
	// 1: +15
	// 2: us
	all := reg.FindSubmatch(s)
	if len(all) != 3 {
		return errors.New("regexp parse result error"), 0.0
	}
	sec, err := strconv.ParseFloat(string(all[1][:]), 64)
	if err != nil {
		log.Errorf("sec %+v parseFloat error", sec)
	}
	unit := string(all[2][:])
	// 转换成毫秒ms
	switch unit {
	case "ps":
		sec /= math.Pow10(9)
	case "ns":
		sec /= math.Pow10(6)
	case "us":
		sec /= math.Pow10(3)
	case "s":
		sec *= math.Pow10(3)
	}
	log.Debugf("Offset :%+v 毫秒", sec)
	return nil, sec
}

func NtpMetrics() (L []*model.MetricValue) {
	// 优先尝试/usr/bin/ntpq -np 然后 /usr/bin/chronyc tracking
	commandName := "/usr/bin/timeout --signal=KILL 2s /usr/bin/ntpq -np"
	resStr := g.ExeSysCommand(commandName)

	if resStr != "FAILED" {
		resSlice := strings.Split(resStr, "\n")

		for _, line := range resSlice {
			if strings.HasPrefix(line, "*") {
				fields := strings.Fields(line)
				// 第八行是offset 单位毫秒
				ntpOffset, parserror := strconv.ParseFloat(fields[8], 64)
				if parserror != nil {
					log.Errorf("parseunit error:%+v", parserror)
					return nil
				}
				L = append(L, GaugeValue("sys.ntp.offset", ntpOffset))
			}
		}
		return
	} else {
		//commandName = "/usr/bin/timeout --signal=KILL 2s /usr/bin/chronyc tracking"
		commandName = " /usr/bin/chronyc sourcestats |grep  $(/usr/bin/chronyc sources |grep '*' |awk '{print $2}') |awk '{print $(NF-1)}' "
		resStr = g.ExeSysCommand(commandName)
		if resStr == "FAILED" {
			return
		}
		err, ntpOffset := chronycParseOffset(resStr)
		if err != nil {
			log.Errorf("chronycParseOffset error:%+v", err)
			return
		}
		log.Debugf("chronyc sys.ntp.offset:%+v", ntpOffset)
		L = append(L, GaugeValue("sys.ntp.offset", ntpOffset))
		return
	}

	return
}
