package cron

import (
	"fmt"

	cmodel "github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/api/app/model/uic"
)

func CommonFilterBlock(event *cmodel.Event, userMap map[string]*uic.User) map[string]*uic.User {
	// 可以按metric 屏蔽
	// 可以按endpoint+metric 屏蔽
	NewMap := make(map[string]*uic.User)

	for userName, user := range userMap {
		counter := fmt.Sprintf("%s_%s", event.Endpoint, event.Metric())
		euKey := fmt.Sprintf("%s%s_%s", g.BLOCK_MONITOR_KEY_PREFIX, userName, counter)
		mUKey := fmt.Sprintf("%s%s_%s", g.BLOCK_MONITOR_KEY_PREFIX, userName, event.Metric())

		_, euE := BlockMonitorCounter.Load(euKey)
		_, muE := BlockMonitorCounter.Load(mUKey)
		if euE == false && muE == false {
			NewMap[userName] = user
		}

	}
	return NewMap
}
