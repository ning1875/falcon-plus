package utils

import (
	"github.com/op/go-logging"
)

func InitLogger() *logging.Logger {

	var format = logging.MustStringFormatter(
		//`%{time:2006-01-02 15:04:05.000} %{shortfile} %{longfunc} %{pid} %{level:.4s} %{message}`,
		`%{shortfile} %{longfunc} %{pid} %{level:.4s} %{message}`,
	)
	logging.SetFormatter(format)
	var log = logging.MustGetLogger("example")
	return log
}
