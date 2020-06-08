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

package g

import (
	"time"
)

// changelog:
// 3.1.3: code refactor
// 3.1.4: bugfix ignore configuration
// 5.0.0: 支持通过配置控制是否开启/run接口；收集udp流量数据；du某个目录的大小
// 5.1.0: 同步插件的时候不再使用checksum机制
// 5.1.1: 修复往多个transfer发送数据的时候crash的问题
// 5.1.2: ignore mount point when blocks=0
// 6.0.0: agent自升级,新增一些监控项
// 6.0.1: agent collect level
// 6.0.2: 添加单核监控开关默认不打开,单核监控tag变更为core=core0x ,添加mem.available.percent
// 6.0.3: 增加sys.uptime
// 6.0.4: 修复cpu.iowait>100的bug
// 6.0.5: 添加进程采集监控，间隔30s
// 6.0.6: 调整内置的采集func间隔 disk io相关和tcp 10s-->30s,agent_version 整数代表当前版本,去掉动态监控方法
// 6.0.7: ntp 支持chronyc ，服务监控rpc call 间隔调整为一分钟
// 6.0.8: 修改监控项抓取时间间隔， 10s只保留cpu，解决断点问题
// 6.0.9: 修复dfa dfb块设备采集,修复不同版本ss-s的bug
// 6.1.0: 修复机器上主机名被改case,使ip转化为nxx-xx-xx的形式
const (
	VERSION          = "6.1.0"
	COLLECT_INTERVAL = time.Second
	URL_CHECK_HEALTH = "url.check.health"
	NET_PORT_LISTEN  = "net.port.listen"
	DU_BS            = "du.bs"
	PROC_NUM         = "proc.num"
	UPTIME           = "sys.uptime"
)
