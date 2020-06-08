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
	"runtime"

	"github.com/open-falcon/falcon-plus/common/utils"
)

// change log:
// 1.0.7: code refactor for open source
// 1.0.8: bugfix loop init cache
// 1.0.9: update host table anyway
// 1.1.0: remove Checksum when query plugins
const (
	VERSION = "1.1.0-xxx"

	REDISAGENTUPGRADESET = "agent_upgrade_set" //redis 中agent升级的set作为控制并发的队列

	REDIS_AGENT_UPGRADE_SET_EXPIRE = 70 //并发队列过期时间，避免异常agent占据升级队列

	MapCleanInterval = 1 //过期agent map清理周期，单位小时

	AgentUpdateQueueLength = 2000 //自升级agent的并发上线
)

var Logger = utils.InitLogger()

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
