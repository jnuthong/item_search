package terminal

import (
	"os"
	"io"
	"os/signal"
	"strings"
	"time"
	"fmt"
	"path/filepath"

	"github.com/peterh/liner"
	"github.com/robertkrimen/otto"
	"github.com/imdario/mergo"

	"github.com/jnuthong/item_search/utils"
	"github.com/jnuthong/item_search/utils/log"
	"github.com/jnuthong/item_search/db/mongo"
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
	ps1 = "item >"
	ps2 = " ... >"	

	file_dir, err =filepath.Abs(filepath.Dir(os.Args[0]))
       	history = file_dir + "/" + ".history"
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
				instance[field] = map[string]string{"$regex": "/*" + value  + "*/"}	
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
	
	// .with function filter the relevant field
	// TODO list
	vm.Set("with", func(call otto.FunctionCall) otto.Value{		
			return otto.NullValue()
	})

	// .without function filter the unwanted field
	vm.Set("without", func(call otto.FunctionCall) otto.Value{
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
		log.Log("error", fmt.Sprintf("%s", err) + "\n" + info)
	}

	return value
}

func CmdParser(cmd string, vm *otto.Otto) error{
	cmd_list, cmd_num := utils.PathParser(cmd, ".")
	hasCommand := func(cmd string) bool {
		return strings.Contains(cmd, "has")
	}
	var has_list []string
	has_list = utils.FilterString(hasCommand, cmd_list, cmd_num, has_list)

	// package all the has function parameter into VARIABLE:acc
	acc := make(map[string]interface{})
	for index := range has_list{
		value, err := runUnsafe(has_list[index], vm).Export()
		if err != nil{
			log.Log("error", "[error] " + fmt.Sprintf("%s", err) + "\n")
			continue
		}
		mergo.Merge(&acc, value)
	}
	// fmt.Println("---- result ----\n", acc)		
	return acc
}

func Repl(histPath string, *mongo.MongoIndex) error {
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
				// do another thing here
				code = strings.Trim(code, ";")
				fmt.Println("------ Command ------\n" + code + "\n")
				// call function here
				err := CmdParser(code, vm)
				fmt.Println(err)
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
