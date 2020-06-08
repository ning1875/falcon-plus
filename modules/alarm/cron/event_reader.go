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
	"encoding/json"
	"time"

	"errors"

	log "github.com/Sirupsen/logrus"
	redisc "github.com/chasex/redis-go-cluster"
	cmodel "github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	eventmodel "github.com/open-falcon/falcon-plus/modules/alarm/model/event"
	"github.com/open-falcon/falcon-plus/modules/alarm/redi"
)

func ReadHighEvent() {
	queues := g.Config().Redis.HighQueues
	if len(queues) == 0 {
		return
	}

	for {
		/*brpop 多个队列的1条返回event
		1.传入的是包含多个高优先级的队列的列表比如[p0,p1,p2]
		那么总是先pop完event:p0的队列,然后才是p1 ,p2(这里我进行过实测)
		2.单纯的popevent速度是很快的,但是每次循环里都有下面的consume,如果
		consume速度慢的话会直接影响整体的pop速度,我观察过再没加goroutine之前
		pop速度大概5条/s ,如果报警过多会有堆积现象,之前看到会有4个小时左右的延迟
		*/
		event, err := popEvent(queues)
		if err != nil {
			log.Errorf("[popEvent_high_error],event:%+v,error:%+v", event, err)
			time.Sleep(time.Second)
			continue
		}
		//这里的consume其实和popevent已经没关系了,所以异步执行,但是可能会产生过多的goroutine
		go consume(event, true)
	}
}

func ReadLowEvent() {
	queues := g.Config().Redis.LowQueues
	if len(queues) == 0 {
		return
	}

	for {
		event, err := popEvent(queues)
		if err != nil {
			log.Errorf("[popEvent_low_error],event:%+v,error:%+v", event, err)
			time.Sleep(time.Second)
			continue
		}
		go consume(event, false)
	}
}

func popEvent(queues []string) (*cmodel.Event, error) {
	count := len(queues)

	params := make([]interface{}, count+1)
	for i := 0; i < count; i++ {
		params[i] = queues[i]
	}
	// set timeout 0
	params[count] = 0

	//rc := g.RedisConnPool.Get()
	rc := redi.RedisCluster
	//defer rc.Close()

	var eventStr string
	Nilp := errors.New("nil reply")
	GetFlag := true
	for {

		for _, q := range queues {
			reply, err := redisc.String(rc.Do("RPOP", q))
			if err != Nilp && reply != "" {
				eventStr = reply
				GetFlag = false
				break
			}
		}
		if GetFlag == false {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	//reply, err := redis.Strings(rc.Do("BRPOP", params...))
	//reply, err := redisc.Strings(rc.Do("BRPOP", params...))
	//if err != nil {
	//	log.Errorf("get alarm event from redis fail: %v", err)
	//	return nil, err
	//}

	var event cmodel.Event
	err := json.Unmarshal([]byte(eventStr), &event)
	if err != nil {
		log.Errorf("parse alarm event fail: %v", err)
		return nil, err
	}
	//log.Println("pop event: %s", event.String())
	log.Infof("[popEvent]end_from_redis event_id:%s,event:%+v", event.Id, event)
	//log.Debugf("pop event: %s", event.String())

	//insert event into database
	//eventmodel.InsertEvent(&event)
	go eventmodel.InsertEvent(&event)
	// events no longer saved in memory
	return &event, nil
}
