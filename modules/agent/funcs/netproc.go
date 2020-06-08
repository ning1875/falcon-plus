package funcs

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
)

type PidNet struct {
	TcpSend int64
	TcpRecv int64
	UdpSend int64
	UdpRecv int64
}

func (this *PidNet) String() string {
	return fmt.Sprintf("<TcpSend: %d, TcpRecv: %d, UdpSend: %d, UdpRecv: %d>",
		this.TcpSend,
		this.TcpRecv,
		this.UdpSend,
		this.UdpRecv)
}

func getCmd(pids []string) string {
	pidstring := strings.Join(pids, "|")
	cmd := fmt.Sprintf("/usr/bin/timeout --signal=KILL 2s /usr/bin/pmval bcc.proc.network_length -s 1 |grep -P '%v'", pidstring)
	return cmd
}

func stringToInt64(str string) int64 {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

func getAllPidNet(pids []string) map[string]*PidNet {
	pidNets := make(map[string]*PidNet)

	cmd := getCmd(pids)
	//stdout, _, err := g.ShellCmdTimeout(2, cmd)
	stdout, err := g.ExeShellCommand(cmd)
	if err != nil || stdout == "" {
		log.Debugln("Exe cmd err:", cmd, err)
		return pidNets
	}

	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		fileds := strings.Fields(line)
		if len(fileds) < 5 {
			continue
		}
		pidNets[fileds[1]] = &PidNet{
			TcpRecv: stringToInt64(fileds[2]),
			TcpSend: stringToInt64(fileds[3]),
			UdpRecv: stringToInt64(fileds[4]),
			UdpSend: stringToInt64(fileds[5]),
		}
	}
	return pidNets
}

func ProcNetMetrics(service string, pids []string) []*model.MetricValue {
	if service == "" || len(pids) == 0 {
		return []*model.MetricValue{}
	}

	var tcpSendSum, tcpRecvSum, udpSendSum, udpRecvSum int64

	pidNets := getAllPidNet(pids)
	if len(pidNets) == 0 {
		return []*model.MetricValue{}
	}

	for pid, _ := range pidNets {
		tcpSendSum += pidNets[pid].TcpSend
		tcpRecvSum += pidNets[pid].TcpRecv
		udpSendSum += pidNets[pid].UdpSend
		udpRecvSum += pidNets[pid].UdpRecv
	}

	tag := fmt.Sprintf("service=%s", service)
	return []*model.MetricValue{
		CounterValue("proc.net.tcp_send", tcpSendSum, tag),
		CounterValue("proc.net.tcp_recv", tcpRecvSum, tag),
		CounterValue("proc.net.udp_send", udpSendSum, tag),
		CounterValue("proc.net.udp_recv", udpRecvSum, tag),
	}
}

type SSInfo struct {
	Timestamp int64
	RTT       float64
	LocalIp   string
	LocalPort int64
	DestIp    string
	DestPort  int64
}

func (this *SSInfo) String() string {
	return fmt.Sprintf("<TcpSend: %d, TcpRecv: %d, UdpSend: %d, UdpRecv: %d>",
		this.Timestamp,
		this.RTT,
		this.LocalIp,
		this.LocalPort,
		this.DestIp,
		this.DestPort,
	)
}

func getAllPidSs(pids []string) map[string][]*SSInfo {
	pidSS := make(map[string][]*SSInfo)
	pidstring := strings.Join(pids, ",")
	cmd := fmt.Sprintf("/usr/bin/timeout --signal=KILL 2s /var/lib/pcp/pmdas/ss/toplat -p %v", pidstring)
	stdout, err := g.ExeShellCommand(cmd)

	if err != nil || stdout == "" {
		log.Println("Exe cmd err:", cmd, err)
		return pidSS
	}

	jsonObj := make(map[string][][]interface{})
	if err := json.Unmarshal([]byte(stdout), &jsonObj); err != nil {
		return pidSS
	}

	for pid, pval := range jsonObj {
		if len(pval) == 0 {
			continue
		}

		pidinfo := make([]*SSInfo, 0)
		for _, sval := range pval {
			if len(sval) == 0 {
				continue
			}
			pidinfo = append(pidinfo, &SSInfo{
				Timestamp: int64(sval[0].(float64)),
				RTT:       sval[1].(float64),
				LocalIp:   sval[2].(string),
				LocalPort: int64(sval[3].(float64)),
				DestIp:    sval[4].(string),
				DestPort:  int64(sval[5].(float64)),
			})
		}
		pidSS[pid] = pidinfo
	}
	return pidSS
}

func ProcSsMetrics(service string, pids []string) []*model.MetricValue {
	metrics := []*model.MetricValue{}
	if service == "" || len(pids) == 0 {
		return metrics
	}

	pidSs := getAllPidSs(pids)
	if len(pidSs) == 0 {
		return metrics
	}

	for pid, _ := range pidSs {
		if len(pidSs[pid]) == 0 {
			continue
		}

		tag := fmt.Sprintf("service=%s,spid=%s", service, pid)
		for key, val := range pidSs[pid] {
			subtag := fmt.Sprintf("%s,top=%d,uipport=%s:%d-%s:%d",
				tag, key+1, val.LocalIp, val.LocalPort, val.DestIp, val.DestPort)
			tmpval := GaugeValue("proc.ss.toplat", val.RTT, subtag)
			metrics = append(metrics, tmpval)
		}
	}
	return metrics
}
