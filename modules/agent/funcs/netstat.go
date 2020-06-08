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
	"io"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/toolkits/nux"
)

var USES = map[string]struct{}{
	"PruneCalled":        {},
	"LockDroppedIcmps":   {},
	"ArpFilter":          {},
	"TW":                 {},
	"DelayedACKLocked":   {},
	"ListenOverflows":    {},
	"ListenDrops":        {},
	"TCPPrequeueDropped": {},
	"TCPTSReorder":       {},
	"TCPDSACKUndo":       {},
	"TCPLoss":            {},
	"TCPLostRetransmit":  {},
	"TCPLossFailures":    {},
	"TCPFastRetrans":     {},
	"TCPFastRetransRate": {},
	"TCPTimeouts":        {},
	"TCPSchedulerFailed": {},
	"TCPAbortOnMemory":   {},
	"TCPAbortOnTimeout":  {},
	"TCPAbortFailed":     {},
	"TCPMemoryPressures": {},
	"TCPSpuriousRTOs":    {},
	"TCPBacklogDrop":     {},
	"TCPMinTTLDrop":      {},
}

func getTcpInfo() [][]string {
	file, err := os.Open("/proc/net/snmp")
	if err != nil {
		log.Error("Cannot open /proc/net/snmp")
		return nil
	}

	var lines = make([][]string, 0, 2)
	read := bufio.NewReader(file)
	for {
		r, _, err := read.ReadLine()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Error("Read /proc/net/snmp error:", err)
			return nil
		}

		rs := strings.Fields(string(r))
		if rs[0] == "Tcp:" {
			lines = append(lines, rs)
		}
	}

	return lines
}

func parseTcpInfo() map[string]int64 {
	lines := getTcpInfo()
	if lines == nil {
		return nil
	}

	metrics := map[string]int64{
		"InSegs":      0,
		"OutSegs":     0,
		"RetransSegs": 0,
		"InErrs":      0,
	}

	for index, key := range lines[0] {
		if _, ok := metrics[key]; ok {
			val, err := strconv.ParseInt(lines[1][index], 10, 64)
			if err != nil {
				log.Error("Str conv int 64 error:", err, lines[1][index])
				return nil
			}

			metrics[key] = val
		}
	}

	return metrics
}

func NetstatMetrics() (L []*model.MetricValue) {
	tcpExts, err := nux.Netstat("TcpExt")

	if err != nil {
		log.Println(err)
		return
	}

	cnt := len(tcpExts)
	if cnt == 0 {
		return
	}
	net_prex := []string{"eth"}
	netIfs, err := nux.NetIfs(net_prex)

	var netIfOutPacketsEth0 int64
	for _, i := range netIfs {
		if i.Iface == "eth0" {
			netIfOutPacketsEth0 = i.OutPackages
		}
	}
	var netIfOutPacketsEth0Rate float64
	for key, val := range tcpExts {
		if _, ok := USES[key]; !ok {
			continue
		}
		if key == "TCPFastRetrans" {
			if netIfOutPacketsEth0 == 0 {
				netIfOutPacketsEth0Rate = 0.0
			} else {
				netIfOutPacketsEth0Rate = float64(val) * 10000 / float64(netIfOutPacketsEth0)
			}
			var NewKey = "TcpExt.TCPFastRetransRate"
			L = append(L, GaugeValue(NewKey, netIfOutPacketsEth0Rate))
		}
		L = append(L, CounterValue("TcpExt."+key, val))

	}

	// add tcp package send/recv info
	ret := parseTcpInfo()
	if ret == nil {
		return
	}

	for key, val := range ret {
		L = append(L, CounterValue("TcpPkg."+key, val))
	}

	return
}
