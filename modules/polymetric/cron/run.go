package cron

import (
	"time"
)

func RunAllPoly() {
	for {
		// 同步group策略
		InitGroupStrategy()
		//time.Sleep(time.Duration(time.Second * 120))
		time.Sleep(time.Duration(time.Second * RunPolyInterval))
	}

}
