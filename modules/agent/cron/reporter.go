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

package cron

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
)

var (
	Services    = make(map[string][]string)
	ServiceLock = new(sync.RWMutex)
)

func ReportAgentStatus() {
	if g.Config().Heartbeat.Enabled && g.Config().Heartbeat.Addr != "" {
		go reportAgentStatus(time.Duration(g.Config().Heartbeat.Interval) * time.Second)
	}
}

func SyncDynamicMonitoringConfig() {
	for {
		hostname, err := g.Hostname()
		if err != nil {
			hostname = fmt.Sprintf("error:%s", err.Error())
		}

		req := model.AgentReportRequest{
			Hostname: hostname,
			IP:       g.IP(),
		}
		var resp model.AgentDynamicMonitoringConfigRpcResponse
		err = g.HbsClient.Call("Agent.DynamicMonitoring", req, &resp)
		if err != nil {
			log.Infoln("RPC Response Error:", err)
		}
		g.SetDynamicMonitoringConfig(resp.Args)
		time.Sleep(time.Second * 10)
	}
}

//同步service name列表
func SyncServiceConfig() {
	//从hbs同步service name即可,立即生成pid映射
	for {
		hostname, err := g.Hostname()
		if err != nil {
			hostname = fmt.Sprintf("error:%s", err.Error())
		}

		req := model.AgentReportRequest{
			Hostname: hostname,
			IP:       g.IP(),
		}

		var resp model.AgentServiceConfigRpcResponse
		err = g.HbsClient.Call("Agent.SyncServiceNames", req, &resp)
		if err != nil {
			log.Errorln("RPC Response Error: ", err)
			//time.Sleep(time.Second)
			//continue
		}

		if resp.Code == 1 {
			log.Errorln("HostName report error")
			//time.Sleep(time.Second)
			//continue
		}

		if resp.Code == 2 {
			log.Debugln("Service getted from hbs is null")
		}

		//service为空也得传下去判断，用于删除老的service
		DynamicServiceWatcher(trimStr(resp.Args))
		time.Sleep(time.Second * 60)
	}
}

func trimStr(str string) []string {
	if str == "" || str == "," {
		return []string{}
	}

	strs := strings.Split(str, ",")
	if strs[len(strs)-1] == "" {
		strs = strs[:len(strs)-1]
	}

	rets := []string{}
	strmap := make(map[string]string, len(strs))

	for _, val := range strs {
		if _, ok := strmap[val]; !ok {
			rets = append(rets, val)
			strmap[val] = val
		}
	}

	return rets
}

func DeleteService(name string) {
	g.ServiceChangeSigns[name] <- 1
	delete(g.ServiceChangeSigns, name)
	delete(Services, name)
}

func DynamicServiceWatcher(sNames []string) {
	//切割service，并获取pid，生成映射
	ServiceLock.Lock()
	defer ServiceLock.Unlock()

	log.Debugln("Services from hbs: ", sNames)
	log.Debugln("Services from mem: ", Services)
	if len(sNames) == 0 && len(Services) == 0 {
		log.Debugln("Service from hbs and Service in memory is both null")
		return
	}

	//寻找砍掉的service从services删除并发送停止信号
	for name, _ := range Services {
		var hasValue bool
		for _, tmpname := range sNames {
			if name == tmpname {
				hasValue = true
				break
			}
		}
		if !hasValue {
			log.Infoln("Delete service: ", name)
			DeleteService(name)
		}
	}

	if len(sNames) == 0 {
		log.Debugln("No service to collect and exit")
		return
	}

	var newServices = make(map[string][]string)
	currentServices := GetPidsForAllSysService()
	//如果没有获取到系统service:pids 映射，不启动新采集并直接退出
	if currentServices == nil || len(currentServices) == 0 {
		log.Errorln("GetPidsForAllSysService Error")
		return
	}

	//[1,2,3]   old	services
	//[3,4,5]   new	snames
	//遍历从hbs同步的service列表查看是否跟内存中的service一致
	for _, name := range sNames {
		if pids, ok := Services[name]; ok {
			//service已有，查看pids是否变化
			newpids, ok := currentServices[name]
			if !ok || newpids == nil || len(newpids) == 0 {
				//service stop
				log.Infoln("Changed and delete service:", name)
				DeleteService(name)
			} else {
				//service restart
				if !g.StrSliceEqualBCE(pids, newpids) { //pid changed
					log.Infoln("Changed service:", name)
					g.ServiceChangeSigns[name] <- 1 //发送信号停掉老service
					newServices[name] = newpids     //add new service
					Services[name] = newpids
				}
			}
		} else {
			//service无,直接新增
			if name == "" {
				continue
			}
			pids, ok := currentServices[name]
			if ok && pids != nil && len(pids) != 0 {
				//pid 有效，增加此service
				log.Infoln("Add new service:", name)
				g.ServiceChangeSigns[name] = make(chan int)
				newServices[name] = pids
				Services[name] = pids
			} else {
				log.Debugln("Error in get pids for service:", name)
			}
		}
	}

	if len(newServices) != 0 {
		//服务和进程保存到本机
		finishSig := make(chan int)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		go writeFile(finishSig, Services)
		defer cancel()

		select {
		case <-ctx.Done():
			log.Debugln("Sync service and pids to disk timeout")
		case <-finishSig:
			log.Debugln("Sync service and pids to disk successfully")
		}

		//创建采集进程
		CreateProcMapperWatcher(newServices)
	}
}

func FindService(line string) string {
	reg := regexp.MustCompile(`^([a-zA-Z].*).(service|scope)$`)
	return reg.FindString(line)
}
func FindPid(line string) string {
	reg := regexp.MustCompile(`^(\d+)`)
	return reg.FindString(line)
}

//获取系统所有service和pids 映射表
func GetPidsForAllSysService() map[string][]string {
	var (
		name string
		rets = make(map[string][]string)
	)

	cmd := "/bin/systemctl status --no-page|sed 's/^ .*─//g'|sed 's/^[ \t]*//g'"
	stdout, stderr, err := g.ShellCmdTimeout(5, "/bin/bash", "-c", cmd)
	if stdout == "" || err != nil {
		log.Debugln("Exe error:", cmd, err, stderr)
		return nil
	}

	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if ret := FindPid(line); ret != "" {
			rets[name] = append(rets[name], ret)
		} else if ret := FindService(line); ret != "" {
			name = strings.Replace(ret, ".service", "", -1)
			name = strings.Replace(name, ".scope", "", -1)
		} else {
			name = line
		}
	}
	return rets
}

func reportAgentStatus(interval time.Duration) {
	for {
		hostname, err := g.Hostname()
		if err != nil {
			hostname = fmt.Sprintf("error:%s", err.Error())
		}

		req := model.AgentReportRequest{
			Hostname:      hostname,
			IP:            g.IP(),
			AgentVersion:  g.VERSION,
			PluginVersion: g.GetCurrPluginVersion(),
			InUpgrading:   g.InUpgrading,
		}
		var resp model.AgentUpgradeRpcResponse
		err = g.HbsClient.Call("Agent.ReportStatus", req, &resp)
		if err != nil || resp.Code != 0 {
			log.Infoln("call Agent.ReportStatus fail:", err, "Request:", req, "Response:", resp)
		}
		if resp.Args != nil && g.InUpgrading == false && g.Config().SelfUpgrade == true {
			log.Infoln("got upgrade command from HBS server %+v", resp)
			g.InUpgrading = true
			g.UpgradeChannel <- resp.Args
		}
		time.Sleep(interval)
	}
}

func mvFile(from string, to string) error {
	if exist := g.CheckFileExist(from); !exist {
		log.Infoln("Not exist, skip to mv", from)
		return nil
	}

	cmd := exec.Command("mv", from, to)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// service/pids 保存到本机
func writeFile(finishSig chan int, data map[string][]string) error {
	path := g.Config().AppBaseDir
	err := g.CheckPathAndMkdir(path)
	if err != nil {
		log.Errorln("Path not exist and create fail", path, err)
		return err
	}

	name := path + "/service_pids"
	backup := name + ".backup"
	if err = mvFile(name, backup); err != nil {
		log.Errorf("Backup %s fail: %s", name, err)
		return err
	}

	//使用io.WriteString()函数进行数据的写入
	fileObj, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Errorln("Failed to open the file and rollback", err)
		mvFile(backup, name)
		return err
	}
	defer fileObj.Close()

	for svs, pids := range data {
		stringpids := strings.Join(pids, ",")
		stringToApp := svs + ":" + stringpids + "\n"
		if err != nil {
			log.Errorln("Fail to printf")
			mvFile(backup, name)
			return err
		}
		if _, err := io.WriteString(fileObj, stringToApp); err != nil {
			log.Errorln("Fail to append to the file with os.OpenFile and io.WriteString.", stringpids)
			mvFile(backup, name)
			return err
		}
	}

	finishSig <- 1
	return nil
}
