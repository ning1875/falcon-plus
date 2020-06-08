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

package cache

import (
	"sync"
	"time"

	"github.com/open-falcon/falcon-plus/modules/hbs/g"
)

var wg = new(sync.WaitGroup)

func Init() {

	g.Logger.Info("cache begin")

	wg.Add(1)
	go func() {
		g.Logger.Info("ExpressionCache...")
		ExpressionCache.Init()
		g.Logger.Info("ExpressionCache done...")
		wg.Done()
		for {
			time.Sleep(time.Minute)
			ExpressionCache.Init()
		}
	}()

	wg.Add(1)
	go func() {
		g.Logger.Info("ServicesConfigs...")
		Services.Init()
		g.Logger.Info("ServicesConfigs done...")
		wg.Done()
		for {
			time.Sleep(time.Minute)
			Services.Init()
		}
	}()
	wg.Wait()
	g.Logger.Info("cache done")

}

//func Init() {
//
//	log.Println("cache begin")
//
//	wg.Add(1)
//	go func() {
//		log.Println("\n#1 GroupPlugins...")
//		GroupPlugins.Init()
//		log.Println("\n#1 GroupPlugins done...")
//		wg.Done()
//		for {
//			time.Sleep(time.Minute)
//			GroupPlugins.Init()
//		}
//	}()
//
//	wg.Add(1)
//	go func() {
//		log.Println("\n#2 GroupTemplates...")
//		GroupTemplates.Init()
//		log.Println("\n#2 GroupTemplates done...")
//		wg.Done()
//		for {
//			time.Sleep(time.Minute)
//			GroupTemplates.Init()
//		}
//	}()
//
//	wg.Add(1)
//	go func() {
//		log.Println("\n#3 HostGroupsMap...")
//		HostGroupsMap.Init()
//		log.Println("\n#3 HostGroupsMap done...")
//		wg.Done()
//		for {
//			time.Sleep(time.Minute)
//			HostGroupsMap.Init()
//		}
//	}()
//
//	wg.Add(1)
//	go func() {
//		log.Println("\n#4 HostMap...")
//		HostMap.Init()
//		log.Println("\n#4 HostMap done...")
//		wg.Done()
//		for {
//			time.Sleep(time.Minute)
//			HostMap.Init()
//		}
//	}()
//
//	wg.Add(2)
//	go func() {
//		log.Println("\n#5 TemplateCache...")
//		TemplateCache.Init()
//		log.Println("\n#5 TemplateCache done...")
//		wg.Done()
//		log.Println("\n#6 Strategies...")
//		Strategies.Init(TemplateCache.GetMap())
//		//UnionStrategies.Init()
//		log.Println("\n#6 Strategies done...")
//		wg.Done()
//		for {
//			time.Sleep(time.Minute)
//			TemplateCache.Init()
//			Strategies.Init(TemplateCache.GetMap())
//			//UnionStrategies.Init()
//		}
//	}()
//	wg.Add(1)
//	go func() {
//		log.Println("\n#7 HostTemplateIds...")
//		HostTemplateIds.Init()
//		log.Println("\n#7 HostTemplateIds done...")
//		wg.Done()
//		for {
//			time.Sleep(time.Minute)
//			HostTemplateIds.Init()
//		}
//	}()
//	wg.Add(1)
//	go func() {
//		log.Println("\n#8 ExpressionCache...")
//		ExpressionCache.Init()
//		log.Println("\n#8 ExpressionCache done...")
//		wg.Done()
//		for {
//			time.Sleep(time.Minute)
//			ExpressionCache.Init()
//		}
//	}()
//
//	wg.Add(1)
//	go func() {
//		log.Println("\n#9 MonitoredHosts...")
//		MonitoredHosts.Init()
//		log.Println("\n#9 MonitoredHosts done...")
//		wg.Done()
//		for {
//			time.Sleep(time.Minute)
//			MonitoredHosts.Init()
//		}
//	}()
//	wg.Add(1)
//	go func() {
//		log.Println("\n#10 DynamicConfigs...")
//		DynamicConfig.Init()
//		log.Println("\n#10 DynamicConfigs done...")
//		wg.Done()
//		for {
//			time.Sleep(time.Minute)
//			DynamicConfig.Init()
//		}
//	}()
//	wg.Add(1)
//	go func() {
//		log.Println("\n#11 ServicesConfigs...")
//		Services.Init()
//		log.Println("\n#10 ServicesConfigs done...")
//		wg.Done()
//		for {
//			time.Sleep(time.Minute)
//			Services.Init()
//		}
//	}()
//	wg.Wait()
//	log.Println("\n cache done")
//
//	//go LoopInit()
//
//}

//func LoopInit() {
//	for {
//		time.Sleep(time.Minute)
//		GroupPlugins.Init()
//		GroupTemplates.Init()
//		HostGroupsMap.Init()
//		HostMap.Init()
//		TemplateCache.Init()
//		Strategies.Init(TemplateCache.GetMap())
//		HostTemplateIds.Init()
//		ExpressionCache.Init()
//		MonitoredHosts.Init()
//		DynamicConfig.Init()
//		Services.Init()
//	}
//}
