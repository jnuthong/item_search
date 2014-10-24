
package utils
// package main

import ("runtime"
	// "fmt"
	"os"
	"strconv"
	"os/exec"
	"strings"
)

type function func(value interface{}) interface{}
type fold func(value interface{}, acc map[string]interface{}) map[string]interface{}
type Element struct {
	key string
	value interface{}
}

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

func UpdateMap(value Element, acc map[string]interface{}) map[string]interface{}{
	acc[value.key] = value.value
	return acc
}

// function lang
// map function, take each element in the list and call the function return a list
// PARAM: length - should be length of the xs
func Mapping(f function, xs []interface{}, length int, result []interface{}) []interface{}{
	if length == 0{
		return result
	}
	result = append(result, f(xs[0]))
	return Mapping(f, xs[1:length], length - 1, result) 
}

// fold function, take each element in the list and call the function return acc 
func Folding(f fold, xs []interface{}, length int, result map[string]interface{}) map[string]interface{}{
	if length == 0{
		return result
	}
			
	result = f(xs[0], result) 
	return Folding(f, xs[1:length], length - 1, result)
}

