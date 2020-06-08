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
	"bytes"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/toolkits/file"
	"github.com/toolkits/nux"
)

func MemMetrics() []*model.MetricValue {
	m, err := nux.MemInfo()
	if err != nil {
		log.Println(err)
		return nil
	}

	memFree := m.MemFree + m.Buffers + m.Cached
	memUsed := m.MemTotal - memFree

	pmemFree := 0.0
	pmemUsed := 0.0
	pmemAvai := 0.0
	if m.MemTotal != 0 {
		pmemFree = float64(memFree) * 100.0 / float64(m.MemTotal)
		pmemUsed = float64(memUsed) * 100.0 / float64(m.MemTotal)
		pmemAvai = float64(m.MemAvailable) * 100.0 / float64(m.MemTotal)
	}

	pswapFree := 0.0
	pswapUsed := 0.0
	if m.SwapTotal != 0 {
		pswapFree = float64(m.SwapFree) * 100.0 / float64(m.SwapTotal)
		pswapUsed = float64(m.SwapUsed) * 100.0 / float64(m.SwapTotal)
	}

	return []*model.MetricValue{
		GaugeValue("mem.memtotal", m.MemTotal),
		GaugeValue("mem.memused", memUsed),
		GaugeValue("mem.memfree", memFree),
		GaugeValue("mem.cached", m.Cached),
		GaugeValue("mem.buffers", m.Buffers),
		GaugeValue("mem.swaptotal", m.SwapTotal),
		GaugeValue("mem.swapused", m.SwapUsed),
		GaugeValue("mem.swapfree", m.SwapFree),
		GaugeValue("mem.memfree.percent", pmemFree),
		GaugeValue("mem.memused.percent", pmemUsed),
		GaugeValue("mem.memavaiable.percent", pmemAvai),
		GaugeValue("mem.swapfree.percent", pswapFree),
		GaugeValue("mem.swapused.percent", pswapUsed),
		GaugeValue("mem.shmem", m.Shmem),
		GaugeValue("mem.memavailable", m.MemAvailable),
	}
}

//proc mem collect
type SingleMemInfo struct {
	Peak uint64
	Size uint64
	Lck  uint64
	Hwm  uint64
	Rss  uint64
	Data uint64
	Stk  uint64
	Exe  uint64
	Lib  uint64
	Pte  uint64
	Swap uint64
}

func GetOnePidMemInfo(pid string) *SingleMemInfo {
	if pid == "" {
		return nil
	}

	fName := "/proc/" + pid + "/status"
	bs, err := ioutil.ReadFile(fName)
	if err != nil {
		return nil
	}
	reader := bufio.NewReader(bytes.NewBuffer(bs))

	pmem := &SingleMemInfo{}
	for {
		line, err := file.ReadLine(reader)
		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return nil
		}

		ParseMemLine(line, pmem)

	}
	return pmem
}

func ParseMemLine(line []byte, pmem *SingleMemInfo) {
	fields := strings.Fields(string(line))
	if len(fields) < 2 {
		return
	}

	fieldName := fields[0]
	if fieldName == "VmPeak:" {
		ret, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return
		}
		pmem.Peak = ret
		return
	}

	if fieldName == "VmSize:" {
		ret, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return
		}
		pmem.Size = ret
		return
	}

	if fieldName == "VmLck:" {
		ret, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return
		}
		pmem.Lck = ret
		return
	}

	if fieldName == "VmHWM:" {
		ret, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return
		}
		pmem.Hwm = ret
		return
	}

	if fieldName == "VmRSS:" {
		ret, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return
		}
		pmem.Rss = ret
		return
	}

	if fieldName == "VmData:" {
		ret, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return
		}
		pmem.Data = ret
		return
	}

	if fieldName == "VmStk:" {
		ret, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return
		}
		pmem.Stk = ret
		return
	}

	if fieldName == "VmExe:" {
		ret, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return
		}
		pmem.Exe = ret
		return
	}

	if fieldName == "VmLib:" {
		ret, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return
		}
		pmem.Lib = ret
		return
	}

	if fieldName == "VmPTE:" {
		ret, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return
		}
		pmem.Pte = ret
		return
	}

	if fieldName == "VmSwap:" {
		ret, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return
		}
		pmem.Swap = ret
		return
	}
}

func ProcMemMetrics(name string, pids []string) []*model.MetricValue {
	if pids == nil || len(pids) == 0 {
		return []*model.MetricValue{}
	}

	var (
		peakSum uint64
		sizeSum uint64
		lckSum  uint64
		hwmSum  uint64
		rssSum  uint64
		dataSum uint64
		stkSum  uint64
		exeSum  uint64
		libSum  uint64
		pteSum  uint64
		swapSum uint64
	)
	tag := "service=" + name

	for _, pid := range pids {
		onePid := GetOnePidMemInfo(pid)
		if onePid == nil {
			log.Debugln("Mem proc info get fail:", pid)
			continue
		}

		peakSum += onePid.Peak
		sizeSum += onePid.Size
		lckSum += onePid.Lck
		hwmSum += onePid.Hwm
		rssSum += onePid.Rss
		dataSum += onePid.Data
		stkSum += onePid.Stk
		exeSum += onePid.Exe
		libSum += onePid.Lib
		pteSum += onePid.Pte
		swapSum += onePid.Swap
	}

	return []*model.MetricValue{
		GaugeValue("proc.mem.peak", peakSum, tag),
		GaugeValue("proc.mem.size", sizeSum, tag),
		GaugeValue("proc.mem.lck", lckSum, tag),
		GaugeValue("proc.mem.hwm", hwmSum, tag),
		GaugeValue("proc.mem.rss", rssSum, tag),
		GaugeValue("proc.mem.data", dataSum, tag),
		GaugeValue("proc.mem.stk", stkSum, tag),
		GaugeValue("proc.mem.exe", exeSum, tag),
		GaugeValue("proc.mem.lib", libSum, tag),
		GaugeValue("proc.mem.pte", pteSum, tag),
		GaugeValue("proc.mem.swap", swapSum, tag),
	}
}
