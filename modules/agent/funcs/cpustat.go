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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"github.com/toolkits/file"
	"github.com/toolkits/nux"
)

type ProcCpuUsage struct {
	Utime  uint64 // time spent in user mode
	Stime  uint64 // time spent in user mode with low priority (nice)
	Cutime uint64 // 累计的该任务的所有的 waited-for 进程曾经在用户态运行的时间
	Cstime uint64 // 累计的该任务的所有的 waited-for 进程曾经在核心态运行的时间
	Total  uint64 // total of all time fields
}

type PidInfo map[string]*ProcCpuUsage

const (
	historyCount int = 2
)

var (
	procStatHistory [historyCount]*nux.ProcStat
	psLock          = new(sync.RWMutex)
	spcLock         = new(sync.RWMutex)
	processorNum    int
	ServicesProCpus = make([]map[string]PidInfo, 2)
)

func UpdateCpuStat() error {
	ps, err := nux.CurrentProcStat()
	if err != nil {
		return err
	}

	psLock.Lock()
	defer psLock.Unlock()
	for i := historyCount - 1; i > 0; i-- {
		procStatHistory[i] = procStatHistory[i-1]
	}

	procStatHistory[0] = ps
	processorNum = len(ps.Cpus)
	return nil
}
func UpdateProcCpuStat(services map[string][]string) error {

	if services == nil || len(services) == 0 {
		return nil
	}
	spcLock.Lock()
	defer spcLock.Unlock()
	//存最近2次cpu数据
	for i := historyCount - 1; i > 0; i-- {
		ServicesProCpus[i] = ServicesProCpus[i-1]
	}

	ServicesProCpus[0] = make(map[string]PidInfo)
	for name, pids := range services {
		if pids == nil && len(pids) == 0 {
			return errors.New("Pids is null" + name)
		}
		currentProcCpus, err := CurrentProcCpuStat(pids) //[]*ProcCpuUsage
		if err != nil || currentProcCpus == nil {
			return err
		}
		ServicesProCpus[0][name] = currentProcCpus
	}
	return nil
}

func deltaTotal() uint64 {
	if procStatHistory[1] == nil {
		return 0
	}
	return procStatHistory[0].Cpu.Total - procStatHistory[1].Cpu.Total
}

func deltaTotalByCore(index int) uint64 {
	if procStatHistory[1] == nil {
		return 0
	}
	if index == -1 {
		return procStatHistory[0].Cpu.Total - procStatHistory[1].Cpu.Total
	} else if index < -1 {
		return 0
	} else {
		if index < len(procStatHistory[0].Cpus) && index < len(procStatHistory[1].Cpus) {
			return procStatHistory[0].Cpus[index].Total - procStatHistory[1].Cpus[index].Total
		} else {
			return 0
		}
	}
}

func CpuIdle() float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotal()
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	return float64(procStatHistory[0].Cpu.Idle-procStatHistory[1].Cpu.Idle) * invQuotient
}

func CpuIdleByCore(index int) float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotalByCore(index)
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	if index < len(procStatHistory[0].Cpus) && index < len(procStatHistory[1].Cpus) {
		return float64(procStatHistory[0].Cpus[index].Idle-procStatHistory[1].Cpus[index].Idle) * invQuotient
	} else {
		return 0.0
	}
}

func CpuUser() float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotal()
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	return float64(procStatHistory[0].Cpu.User-procStatHistory[1].Cpu.User) * invQuotient
}

func CpuUserByCore(index int) float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotalByCore(index)
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	if index < len(procStatHistory[0].Cpus) && index < len(procStatHistory[1].Cpus) {
		return float64(procStatHistory[0].Cpus[index].User-procStatHistory[1].Cpus[index].User) * invQuotient
	} else {
		return 0.0
	}
}

func CpuNice() float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotal()
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	return float64(procStatHistory[0].Cpu.Nice-procStatHistory[1].Cpu.Nice) * invQuotient
}

func CpuNiceByCore(index int) float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotalByCore(index)
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	if index < len(procStatHistory[0].Cpus) && index < len(procStatHistory[1].Cpus) {
		return float64(procStatHistory[0].Cpus[index].Nice-procStatHistory[1].Cpus[index].Nice) * invQuotient
	} else {
		return 0.0
	}
}

func CpuSystem() float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotal()
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	return float64(procStatHistory[0].Cpu.System-procStatHistory[1].Cpu.System) * invQuotient
}

func CpuSystemByCore(index int) float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotalByCore(index)
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	if index < len(procStatHistory[0].Cpus) && index < len(procStatHistory[1].Cpus) {
		return float64(procStatHistory[0].Cpus[index].System-procStatHistory[1].Cpus[index].System) * invQuotient
	} else {
		return 0.0
	}
}

func CpuValidate(originValue float64) (right float64) {

	if originValue > float64(100) {
		right = 0.0
	} else {
		right = originValue
	}
	return
}

func CpuIowait() float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotal()
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	value := float64(procStatHistory[0].Cpu.Iowait-procStatHistory[1].Cpu.Iowait) * invQuotient

	return CpuValidate(value)
}

func CpuIowaitByCore(index int) float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotalByCore(index)
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	if index < len(procStatHistory[0].Cpus) && index < len(procStatHistory[1].Cpus) {
		return float64(procStatHistory[0].Cpus[index].Iowait-procStatHistory[1].Cpus[index].Iowait) * invQuotient
	} else {
		return 0.0
	}
}

func CpuIrq() float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotal()
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	return float64(procStatHistory[0].Cpu.Irq-procStatHistory[1].Cpu.Irq) * invQuotient
}

func CpuIrqByCore(index int) float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotalByCore(index)
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	if index < len(procStatHistory[0].Cpus) && index < len(procStatHistory[1].Cpus) {
		return float64(procStatHistory[0].Cpus[index].Irq-procStatHistory[1].Cpus[index].Irq) * invQuotient
	} else {
		return 0.0
	}
}

func CpuSoftIrq() float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotal()
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	return float64(procStatHistory[0].Cpu.SoftIrq-procStatHistory[1].Cpu.SoftIrq) * invQuotient
}

func CpuSoftIrqByCore(index int) float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotalByCore(index)
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	if index < len(procStatHistory[0].Cpus) && index < len(procStatHistory[1].Cpus) {
		return float64(procStatHistory[0].Cpus[index].SoftIrq-procStatHistory[1].Cpus[index].SoftIrq) * invQuotient
	} else {
		return 0.0
	}
}

func CpuSteal() float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotal()
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	return float64(procStatHistory[0].Cpu.Steal-procStatHistory[1].Cpu.Steal) * invQuotient
}

func CpuStealByCore(index int) float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotalByCore(index)
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	if index < len(procStatHistory[0].Cpus) && index < len(procStatHistory[1].Cpus) {
		return float64(procStatHistory[0].Cpus[index].Steal-procStatHistory[1].Cpus[index].Steal) * invQuotient
	} else {
		return 0.0
	}
}

func CpuGuest() float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotal()
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	return float64(procStatHistory[0].Cpu.Guest-procStatHistory[1].Cpu.Guest) * invQuotient
}

func CpuGuestByCore(index int) float64 {
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotalByCore(index)
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	if index < len(procStatHistory[0].Cpus) && index < len(procStatHistory[1].Cpus) {
		return float64(procStatHistory[0].Cpus[index].Guest-procStatHistory[1].Cpus[index].Guest) * invQuotient
	} else {
		return 0.0
	}
}

func CurrentCpuSwitches() uint64 {
	psLock.RLock()
	defer psLock.RUnlock()
	return procStatHistory[0].Ctxt
}

func CpuPrepared() bool {
	psLock.RLock()
	defer psLock.RUnlock()
	return procStatHistory[1] != nil
}

func getCpus() int {
	psLock.RLock()
	defer psLock.RUnlock()
	if len(procStatHistory[1].Cpus) > len(procStatHistory[0].Cpus) {
		return len(procStatHistory[0].Cpus)
	} else {
		return len(procStatHistory[1].Cpus)
	}
}

func CpuMetrics() []*model.MetricValue {
	if !CpuPrepared() {
		return []*model.MetricValue{}
	}

	cpuIdleVal := CpuIdle()
	idle := GaugeValue("cpu.idle", cpuIdleVal)
	busy := GaugeValue("cpu.busy", 100.0-cpuIdleVal)
	user := GaugeValue("cpu.user", CpuUser())
	nice := GaugeValue("cpu.nice", CpuNice())
	system := GaugeValue("cpu.system", CpuSystem())
	iowait := GaugeValue("cpu.iowait", CpuIowait())
	irq := GaugeValue("cpu.irq", CpuIrq())
	softirq := GaugeValue("cpu.softirq", CpuSoftIrq())
	steal := GaugeValue("cpu.steal", CpuSteal())
	guest := GaugeValue("cpu.guest", CpuGuest())
	switches := CounterValue("cpu.switches", CurrentCpuSwitches())
	cpuCnt := getCpus()
	cpuNums := GaugeValue("host.cpucores", cpuCnt)
	metrics := []*model.MetricValue{idle, busy, user, nice, system, iowait, irq, softirq, steal, guest, switches, cpuNums}
	// 如果配置中开启了 单核采集的开关
	if g.Config().CpuPerCoreCollect == true {
		// 过程中可能变化由获取函数保护
		for i := 0; i < cpuCnt; i++ {
			tag := "core=core" + strconv.Itoa(i)
			cpuIdleVal := CpuIdleByCore(i)
			idle := GaugeValue("percore.idle", cpuIdleVal, tag)
			busy := GaugeValue("percore.busy", 100.0-cpuIdleVal, tag)
			user := GaugeValue("percore.user", CpuUserByCore(i), tag)
			nice := GaugeValue("percore.nice", CpuNiceByCore(i), tag)
			system := GaugeValue("percore.system", CpuSystemByCore(i), tag)
			iowait := GaugeValue("percore.iowait", CpuIowaitByCore(i), tag)
			irq := GaugeValue("percore.irq", CpuIrqByCore(i), tag)
			softirq := GaugeValue("percore.softirq", CpuSoftIrqByCore(i), tag)
			steal := GaugeValue("percore.steal", CpuStealByCore(i), tag)
			guest := GaugeValue("percore.guest", CpuGuestByCore(i), tag)
			metrics = append(metrics, idle, busy, user, nice, system, iowait, irq, softirq, steal, guest)
		}
	}
	return metrics
}

//proc cpu collect
func (this *ProcCpuUsage) String() string {
	return fmt.Sprintf("<Utime:%d, Stime:%d, Cutime:%d, Cstime:%d, Total:%d>",
		this.Utime,
		this.Stime,
		this.Cutime,
		this.Cstime,
		this.Total)
}

func GetProcCpuStat(fName string) (*ProcCpuUsage, error) {
	bs, err := ioutil.ReadFile(fName)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(bytes.NewBuffer(bs))
	line, err := file.ReadLine(reader)
	if err == io.EOF {
		err = nil
	} else if err != nil {
		return nil, err
	}
	pcu := parseLine(line)
	return pcu, nil
}

func CurrentProcCpuStat(pids []string) (PidInfo, error) {
	var currentProcCpus = make(PidInfo)

	for _, pid := range pids {
		if pid == "" {
			log.Debugln("Skip null pid")
			continue
		}

		fName := "/proc/" + pid + "/stat"
		pcu, err := GetProcCpuStat(fName)
		if pcu == nil || err != nil {
			log.Errorln("GetProcCpuStat fail:", pid, err)
			continue
		}
		currentProcCpus[pid] = pcu
	}
	if len(currentProcCpus) == 0 {
		return nil, errors.New("Cannot get final currentProcCpus")
	}
	return currentProcCpus, nil
}

func parseLine(line []byte) *ProcCpuUsage {
	pcu := &ProcCpuUsage{}
	fields := strings.Fields(string(line))
	if len(fields) < 2 {
		return nil
	}
	pcu.Utime, _ = strconv.ParseUint(fields[13], 10, 64)
	pcu.Stime, _ = strconv.ParseUint(fields[14], 10, 64)
	pcu.Cutime, _ = strconv.ParseUint(fields[15], 10, 64)
	pcu.Cstime, _ = strconv.ParseUint(fields[16], 10, 64)
	pcu.Total = pcu.Utime + pcu.Stime + pcu.Cutime + pcu.Cstime
	return pcu
}

func deltaProcCpu() map[string]PidInfo {
	var deltaServicesProCpus = make(map[string]PidInfo)
	if len(ServicesProCpus[1]) == 0 {
		log.Debugln("ServicesProCpus[1] is null")
		return nil
	}

	spcLock.Lock()
	defer spcLock.Unlock()

	for name, procCpuData := range ServicesProCpus[0] {
		if _, ok := ServicesProCpus[1][name]; !ok {
			//服务改名/删除/重启，直接跳过
			log.Debugln("Service not exist in old", name, ServicesProCpus[1])
			continue
		}
		tmpPidsData := make(PidInfo)
		for pid, _ := range procCpuData {
			if _, ok := ServicesProCpus[1][name][pid]; !ok {
				//服务重启导致pid变化，跳过
				log.Debugln("Pid not exist in old", pid, name)
				continue
			}
			tmpCpu := ProcCpuUsage{
				Utime:  ServicesProCpus[0][name][pid].Utime - ServicesProCpus[1][name][pid].Utime,
				Stime:  ServicesProCpus[0][name][pid].Stime - ServicesProCpus[1][name][pid].Stime,
				Cutime: ServicesProCpus[0][name][pid].Cutime - ServicesProCpus[1][name][pid].Cutime,
				Cstime: ServicesProCpus[0][name][pid].Cstime - ServicesProCpus[1][name][pid].Cstime,
				Total:  ServicesProCpus[0][name][pid].Total - ServicesProCpus[1][name][pid].Total,
			}
			tmpPidsData[pid] = &tmpCpu
		}
		deltaServicesProCpus[name] = tmpPidsData
	}
	return deltaServicesProCpus
}

func CalculateProcCpuRate(num uint64) float64 {
	//procCpuRate = (procCpu2 - procCpu1)/(totalCpu2 - totalCpu1) * logicCpuNum * 100
	psLock.RLock()
	defer psLock.RUnlock()
	dt := deltaTotal()
	if dt == 0 {
		return 0.0
	}
	invQuotient := 100.00 / float64(dt)
	return float64(num) * float64(processorNum) * invQuotient
}

//进程cpu采集函数
func ProcCpuMetrics(name string, pids []string) []*model.MetricValue {
	if !CpuPrepared() || len(ServicesProCpus[1]) == 0 {
		return []*model.MetricValue{}
	}

	var useSum, utimeSum, stimeSum, cutimeSum, cstimeSum uint64
	tag := "service=" + name
	metrics := make([]*model.MetricValue, 0)

	deltaServicesProcCpuData := deltaProcCpu()
	cps, ok := deltaServicesProcCpuData[name] //map[string]PidInfo
	if !ok || cps == nil {
		log.Errorln("Get currentProcCpus error:", cps)
		return []*model.MetricValue{}
	}

	//累加pid数据，求和作为结果数据
	for _, singleProc := range cps {
		useSum = useSum + singleProc.Total
		utimeSum = utimeSum + singleProc.Utime
		stimeSum = stimeSum + singleProc.Stime
		cutimeSum = cutimeSum + singleProc.Cutime
		cstimeSum = cstimeSum + singleProc.Cstime
	}

	use := GaugeValue("proc.cpu.use", CalculateProcCpuRate(useSum), tag)
	utime := GaugeValue("proc.cpu.utime", CalculateProcCpuRate(utimeSum), tag)
	stime := GaugeValue("proc.cpu.stime", CalculateProcCpuRate(stimeSum), tag)
	cutime := GaugeValue("proc.cpu.cutime", CalculateProcCpuRate(cutimeSum), tag)
	cstime := GaugeValue("proc.cpu.cstime", CalculateProcCpuRate(cstimeSum), tag)
	metrics = append(metrics, use, utime, stime, cutime, cstime)
	return metrics
}
