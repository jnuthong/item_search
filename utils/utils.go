
package utils
// package main

import ("runtime"
	"fmt"
	"strconv")

func CurrentCallerInfo() string {
	var pc, file, line, ok = runtime.Caller(1)
	var info = ""
	if ok {
		info += "[CallerInfo] - "
	}
	fmt.Println(info)
	info += "PC: " + string(pc) + "||"
	info += "FILE: " + string(file) + "||"
	info += "LINE: " + strconv.Itoa(line) + "||"
	return info
}

// func main() {
//	CurrentCallerInfo()
// }
