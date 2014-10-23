
package utils
// package main

import ("runtime"
	// "fmt"
	"os"
	"strconv"
	"os/exec"
	"strings"
)

func CurrentCallerInfo() string {
	var pc, file, line, ok = runtime.Caller(1)
	var info = ""
	if ok {
		info += "[CallerInfo] - "
	}
	// fmt.Println(info)
	info += "PC: " + string(pc) + "||"
	info += "FILE: " + string(file) + "||"
	info += "LINE: " + strconv.Itoa(line) + "||"
	return info
}

func Exists(path string) bool{
	_, err := os.Stat(path)
	if err == nil{ return true }
	if os.IsNotExist(err){ return false }	
	return false
}

func CountFileLines(path string) (int, error){
	out, err := exec.Command(path).Output()
	if err == nil{
		i, err := strconv.Atoi(string(out[:]))
		return i, err
	}
	return 0, err
}

func PathParser(path string, delimiter string) ([]string, int) {
	str := strings.Split(path, delimiter)
	return str, len(str)
}

func JoinPath(xs []string, delimiter string) string{
	return strings.Join(xs, delimiter)	
}

// func main() {
//	CurrentCallerInfo()
// }
