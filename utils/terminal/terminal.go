package terminal

import (
	"os"
	"io"
	"os/signal"
	"strings"
	"time"
	"fmt"
	//"bytes"
	"path/filepath"
	"encoding/json"

	"github.com/peterh/liner"
	"github.com/robertkrimen/otto"
	"github.com/imdario/mergo"

	"github.com/jnuthong/item_search/utils"
	"github.com/jnuthong/item_search/utils/log"
	// "github.com/jnuthong/item_search/utils/statistic"
	"github.com/jnuthong/item_search/db/mongo"
	"github.com/jnuthong/item_search/query"
)

// REF: https://github.com/peterh/liner
// REF: https://github.com/google/cayley/blob/master/db/repl.go
type Term struct {
	win *liner.State
}

// VM REF: http://godoc.org/github.com/robertkrimen/otto
type VM struct {
	vm *otto.Otto
}

var (
	ps1 		= "item >"
	ps2 		= " ... >"	

	LIMIT 		= 10
	file_dir, err 	= filepath.Abs(filepath.Dir(os.Args[0]))
       	history 	= file_dir + "/" + ".history"

	support_cmd 	= []string{"has", "with", "without", "out"}		// cmd for mongo query
	inner_cmd 	= []string{"hist"}					// cmd for local process
)

// main function in define the query syntax
func InitVM() *otto.Otto{
	vm := otto.New()
	
	vm.Set("search", "search")
	vm.Set("match", "match")

	// .has function achiev the && operation
	// using compose function to optimize the query 
	vm.Set("has", func(call otto.FunctionCall) otto.Value{
		if call.Argument(0).IsDefined() && call.Argument(1).IsDefined() && call.Argument(2).IsDefined() {
			field, err := call.Argument(0).ToString()
			if err != nil{
				fmt.Println(err)
			}
			method, err := call.Argument(1).ToString()
			if err != nil{
				fmt.Println(err)
			}
			if method != "search" && method != "match" {
				fmt.Println("[Error] Current has function only support method: search & match!")
				return otto.NullValue()
			}
			value, err := call.Argument(2).ToString()	
			if err != nil{
				fmt.Println(err)
			}
			// using operator $regex
			instance := make(map[string]interface{})
			if method == "search"{
				instance[field] = map[string]string{"$regex": value}	
			}else{
				instance[field] = value
			}	
			if result, err := vm.ToValue(instance); err == nil{
				return result
			}
		// range search: smaller < vlaue < bigger
		}else if(call.Argument(0).IsDefined() && call.Argument(1).IsDefined()){	
			field, err := call.Argument(0).ToString()
			if err != nil{
				fmt.Println(err)
			}
			value := call.Argument(3).Object()
			instance := make(map[string]interface{})
			instance[field] = value
			if result, err := otto.ToValue(instance); err == nil{
				return result
			}
		}
		return otto.NullValue()
	})
	
	inner_function := func(value_list []otto.Value, filter int) otto.Value{
		sub_function := func(key string) map[string]int{
			x := make(map[string]int)
			x[key] = filter
			return x
		}

		instance := make(map[string]interface{})
		for i := range value_list {
			if x, err := value_list[i].ToString(); err == nil{
				mergo.Merge(&instance, sub_function(x))
			}
		}
		
		if result, err := otto.ToValue(instance); err == nil{
			return result
		}	
		return otto.NullValue()
	}
	
	// .with function filter the relevant field
	vm.Set("with", func(call otto.FunctionCall) otto.Value{
		return inner_function(call.ArgumentList, 1)
	})

	// .without function filter the unwanted field
	vm.Set("without", func(call otto.FunctionCall) otto.Value{
		return inner_function(call.ArgumentList, 0)
	})

	// .out function 
	// .limit function
	// .count function
	// .freq function
	// .hist function
	// FUNCTION:hist TODO list: & operation and || operation
	// return a list of list
	vm.Set("hist", func(call otto.FunctionCall) otto.Value{
		// example: graph.has(field, search, value)...hist("field1&field2", "field3", ...)

		var instance [][]string
		for i := range call.ArgumentList {
			if call.ArgumentList[i].IsDefined(){
				tmp, err := call.ArgumentList[i].ToString()
				if err != nil{
					info := utils.CurrentCallerInfo()
					log.Log("error", "[error] " + fmt.Sprintf("%s", err) + "\n" + info)
					continue
				}
				str, count := utils.PathParser(tmp, "&")
				if (count == 0){ continue }
				instance = append(instance, str)
			}	
		}

		if result, err := otto.ToValue(instance); err == nil{
			return result
		}
		return otto.NullValue()
	})
	return vm
}

// Orignal Code: http://godoc.org/github.com/robertkrimen/otto
func runUnsafe(script string, vm *otto.Otto) otto.Value{
	start := time.Now()
	defer func(){
		duration := time.Since(start)
		if caught := recover(); caught != nil{
			if caught == "halt" {
				fmt.Fprintf(os.Stderr, "Some code took to long! Stopping after: %v\n", duration)
				return
			}
			panic(caught)
		}
		fmt.Fprintf(os.Stderr, "Ran code successfully: %v\n", duration)
	}()
	
	vm.Interrupt = make(chan func(), 1)
	
	go func(){
		time.Sleep(2 * time.Second) // The buffer prevents blocking
		vm.Interrupt <- func(){
			panic("halt")
		}
	}()

	value, err := vm.Run(script)
	if err != nil{
		info := utils.CurrentCallerInfo() 
		log.Log("error", "[error] " + fmt.Sprintf("%s", err) + "\n" + info)
	}

	return value
}

func CmdParser(cmd string, vm *otto.Otto) (map[string]interface{}, map[string]interface{}, error){
	cmd_list, cmd_num := utils.PathParser(cmd, ".")
	support_cmd 	= append(support_cmd, inner_cmd...)
	
	// package all the has function parameter into VARIABLE:acc as final query send to mongo/database
	// service
	acc := make(map[string]interface{}) 				// query send to the database
	inner_acc := make(map[string]interface{})			// query used in called inner function

	for i := range support_cmd {
		command := func(cmd string) bool {
			return strings.Contains(cmd, support_cmd[i])
		}
	
		// filter out the same command
		var has_list []string
		has_list = utils.FilterString(command, cmd_list, cmd_num, has_list)

		// if the command is inner defined, create a branch and break
		// inner_cmd only support trigger by the first parameters
		new := utils.StringToInterface(inner_cmd)
		if utils.InList(support_cmd[i], new){
			if len(has_list) > 1{
				info := utils.CurrentCallerInfo()
				log.Log("warn", "[warn] " + fmt.Sprintf("%s", err) + "\n" + info)
			}else if len(has_list) == 0{
				continue
			}

			value, err := runUnsafe(has_list[0], vm).Export()
			if err != nil{
				info := utils.CurrentCallerInfo()
				log.Log("error", "[error] " + fmt.Sprintf("%s", err) + "\n" + info)
			}
			inner_acc[support_cmd[i]] = value
			continue	
		}

		for index := range has_list{
			value, err := runUnsafe(has_list[index], vm).Export()
			if err != nil{
				info := utils.CurrentCallerInfo()
				log.Log("error", "[error] " + fmt.Sprintf("%s", err) + "\n" + info)
				continue
			}
			mergo.Merge(&acc, value)
		}
	}

	return acc, inner_acc, nil
}

// main terminal function
func Repl(histPath string, db *mongo.MongoObj) error {
	term, err := InitTerm(histPath)
	if err != nil{
		info := utils.CurrentCallerInfo()
		log.Log("error", "[error] " + fmt.Sprintf("%s", err) + info)
		os.Exit(1)
	}	

	var (
		prompt = ps1
		code = ""
		vm = InitVM()
	)
	
	for {
		if len(code) == 0{
			prompt = ps1
		}else{
			prompt = ps2
		}
		line, err := term.Prompt(prompt)
		if err != nil{
			if err != nil{
				if err == io.EOF{
					return nil
				}
			}
			info := utils.CurrentCallerInfo()
			log.Log("error", "[error] " + fmt.Sprintf("%s", err) + info)
			return nil
		}

		term.AppendHistory(line)
		
		line = strings.TrimSpace(line)
		line = strings.Trim(line, "\n")

		// accept multiple line command input
		if len(line) == 0{
			continue
		}
		code += line
		// accepted the input as function
		if strings.HasSuffix(line, ";") {
			// exit safely 
			switch code {
			case "exit;":
				fmt.Println("bye!")
				os.Exit(0)
			case "quit;":
				fmt.Println("bye!")
				os.Exit(0)
			default:
				code = strings.Trim(code, ";")
				if(len(code) == 0){ continue }
				fmt.Println("------ Command ------\n" + code + "\n")

				x, y, err := CmdParser(code, vm)
				if err != nil{
					log.Log("error", "[error] " + fmt.Sprintln("%s", err))
					fmt.Println(y)
				}

				r, count := mongo.Find(db.GetCollection(), x)
				defer r.Close()
				fmt.Println("++++++ Result ++++++")
				if (count == 0) {
					code = ""
					continue	
				}
				
				result := query.Query(r, count)
				for {
					v := result.Pop()
					if v == nil{ break }
					x, err := json.Marshal(v)
					if err != nil{
						info :=utils.CurrentCallerInfo()
						log.Log("error", "[error] " + fmt.Sprintln("%s", err) + "\n" + info)
					}
					fmt.Println(fmt.Sprintf("%s", x))
				}

				/*
				for key := range y{
					if key == "hist"{
						fmt.Printf("Call histogram function to compute the query result, total match result number : %d", count)
						switch y[key].(type) {
						case [][]string:
							statistic.Hist(r, y[key].([][]string))
						default:
							info := utils.CurrentCallerInfo()
							log.Log("error", "[error] " + "Expect [][]string type, but get " + fmt.Sprintf("%s", y[key]) + "\n" + info)
						}

					}
					code = ""
					continue
				} */

				// no inner function process
				/* code for print result
				for i := 1; i <= LIMIT; i++{
					var x interface{}
					r.Next(&x)
					x, err := json.Marshal(x)
					if err != nil{
						info :=utils.CurrentCallerInfo()
						log.Log("error", "[error] " + fmt.Sprintln("%s", err) + "\n" + info)
					}
					fmt.Println(fmt.Sprintf("Record %v :\n %s", i, x))
				}
				*/
				fmt.Println("... \nGet ", count, " docs")
				code = ""
			}
		}
	}
}

func InitTerm(histPath string) (*liner.State, error) {
	term := liner.NewLiner()
	if !utils.Exists(histPath){
		histPath = history	
	}
	// start a synchronize window waitting for user input
	go func(){
		c := make(chan os.Signal, 1)	
		signal.Notify(c, os.Interrupt, os.Kill)
		s := <-c
		defer close(c)
		fmt.Println("Got signal: ", s)
		
		// record the history command	
		err := persist(term, histPath) 
		if err != nil{
			info := utils.CurrentCallerInfo()
			log.Log("error", "[error]" + fmt.Sprintf("%s", err) + "\n" + info)

			os.Exit(1)
		}

		os.Exit(0)
	}()

	f, err := os.OpenFile(histPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0660)
	if err != nil{
		return term, err
	}
	defer f.Close()
	
	_, err = term.ReadHistory(f)
	return term, err
}

func persist(term *liner.State, path string) error{
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil{
		info := utils.CurrentCallerInfo()
		log.Log("error", "[error]" + fmt.Sprintf("%s", err) + "\n" + info)	
	}
	defer f.Close()

	_, err = term.WriteHistory(f)
	if err != nil{
		info := utils.CurrentCallerInfo()
		log.Log("error", "[error]" + fmt.Sprintf("%s", err) + "\n" + info)	
		return err
	} 
	return term.Close()

}
