package funcs

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"strconv"

	"strings"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"github.com/toolkits/file"
)

func SysUptimeMetric() (L []*model.MetricValue) {
	procUptime := "/proc/uptime"
	if !file.IsExist(procUptime) {
		return nil
	}

	contents, err := ioutil.ReadFile(procUptime)
	if err != nil {
		return nil
	}

	reader := bufio.NewReader(bytes.NewBuffer(contents))
	for {
		line, err := file.ReadLine(reader)
		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return nil
		}
		fields := strings.Fields(string(line))
		if len(fields) <= 0 {
			return nil
		}
		if uptime, err := strconv.ParseFloat(fields[0], 64); err != nil {
			return nil
		} else {
			L = append(L, GaugeValue(g.UPTIME, uptime))
			return
		}

	}
	return nil
}
