
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
type function_env func(value interface{}, env interface{}) interface{}
type filter func(value interface{}) bool
type fold func(value interface{}, acc map[string]interface{}) map[string]interface{}
type Element struct {
	Key string
	Value interface{}
}
type Comment_tag struct{
	Emotion string 		`json:"emotion"`
	End interface{} 	`json:"end"`
	Keyword string 		`json:"keyword"`
	Start int 		`json:"start"`
	Num int 		`json:"num"`
	Classify string 	`json:"classify"` 
}

func GeneElement(key string, value interface{}) Element{
	return Element{Key: key, Value : value}
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

func UpdateMap(value Element, acc *map[string]interface{}){
	(*acc)[value.Key] = value.Value
}

// function lang
// map function, take each element in the list and call the function return a list
// PARAM: length - should be length of the xs
func Mapping(f function, xs []interface{}, length int, result *[]interface{}){
	if length == 0{
		return
	}
	*result = append(*result, f(xs[0]))
	Mapping(f, xs[1:length], length - 1, result) 
}

func Map(f func(string, interface{}) interface{}, xs []string, env interface{}, length int, result *[]interface{}){
	if length == 0{
		return
	}
	*result = append(*result, f(xs[0], env))
	Map(f, xs[1:length], env, length - 1, result)
}


func FilterString(f func(string) bool, xs []string, length int, result []string) []string{
	if length == 0{
		return result
	}
	if f(xs[0]){
		result = append(result, xs[0])
	}
	return FilterString(f, xs[1:length], length - 1, result)
}

func Filter(f func(value interface{}) bool, xs []interface{}, length int, result []interface{}) []interface{}{
	if length == 0{
		return result
	}
	if f(xs[0]){
		result = append(result, xs[0])
	}
	return Filter(f, xs[1:length], length - 1, result)
}

// fold function, take each element in the list and call the function return acc 
func Folding(f fold, xs []interface{}, length int, result map[string]interface{}) map[string]interface{}{
	if length == 0{
		return result
	}
			
	result = f(xs[0], result) 
	return Folding(f, xs[1:length], length - 1, result)
}

