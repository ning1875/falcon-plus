package g

import (
	"strconv"
)

func JudgeIsLongInt(input string) (res bool) {
	// True 代表是group_chat False 是私聊
	//input = strings.Split(input, "@")[0]
	_, error := strconv.ParseInt(input, 10, 64)
	if error == nil {
		res = true
	}
	return
}
