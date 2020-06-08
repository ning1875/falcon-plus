// liangyuntao
// 动态配置结构体.

package model

type DynamicConfig struct {
	Id       int
	EndPoint string
	Metrics  string
	Tag      string
	Interval int
}
