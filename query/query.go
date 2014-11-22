package query

import (
	"runtime"
	"encoding/json"
	"io/ioutil"
	// "errors"
	"fmt"
	"path"	
	"gopkg.in/mgo.v2"
	"container/heap"
	"strconv"

	"github.com/jnuthong/item_search/utils"
	"github.com/jnuthong/item_search/utils/log"
	"github.com/barakmich/glog"
	"github.com/jnuthong/item_search/module/pheap"
)

type Field_index struct {
	Name string
	Weight float64
	Function string
	High_order string
	High_order_func string
}

type City_conf struct {
	City string
	Field_index []Field_index
}

type Conf struct {
	Catalog []City_conf 
}

var (
	// useful local variable in this current file closure
	_, file_name, _, _ = runtime.Caller(0)
	dir = path.Dir(file_name)	
	config_path = dir + "/conf_query.file"

	// variable setting up some default behaviour 
	load_bool = false
	conf = new(Conf)
	max_length = 100
)

func LoadConfig(file_path string) *Conf{
	if (file_path == ""){
		file_path = config_path
		fmt.Println(file_path)
	}

	file, e:= ioutil.ReadFile(file_path)
	if e != nil{
		info := utils.CurrentCallerInfo()
		glog.Fatalln(e)
		log.Log("error", "[error] " + info + fmt.Sprintf("%s", e))
	}

	conf 	:= new(Conf)
	err 	:= json.Unmarshal(file, &conf)
	if err != nil{
		info := utils.CurrentCallerInfo()
		glog.Fatalln(err)
		log.Log("error", "[error] " + info + fmt.Sprintf("%s", e))
	}
	return conf
}

func GetRelevantFieldIndex(v map[string]interface{}) []Field_index{
	// should consider more elegant way, temporarily strategy
	city := v["city"] 
	field_list := make([]Field_index, 5)
	label_ := false
	for i := range conf.Catalog {
		if conf.Catalog[i].City == city{
			field_list = conf.Catalog[i].Field_index
			label_ = true
			break	
		}
	}

	if !label_ {
		info := utils.CurrentCallerInfo()
		log.Log("error", "[error] " + info + " Couldn't find relevant city configuration info in conf")
		glog.Fatalf(info + " couldn't find city configuration info")
	}

	return field_list
}

func ProcessSingleRecord(v map[string]interface{}, field_list []Field_index) pheap.Element {
	total 	:= 0.0
	for i := range field_list{
		j, err := strconv.ParseFloat(v[field_list[i].Name].(string), 32)
		if err != nil{
			j = 0.0
			log.Log("info", "[info]" + fmt.Sprintln(" find NA value in filed:%s "))
		}
		value := field_list[i].Weight * j
		total += value
	}
	var x pheap.Element 
	x.Value 	= total
	x.Entity 	= v
	return x
}

func Query(r *mgo.Iter, n int) *pheap.Heap{
	// load the configuration file
	if !load_bool{
		conf = LoadConfig("")
	}

	city_label := false 		// base on the city mode
	var field_list []Field_index
	h := new(pheap.Heap)
	heap.Init(h)

	for i := 1; 1 <= n; i++{
		var x interface{}
		r.Next(&x)	
		x, err := json.Marshal(x)
		if err != nil{
			info := utils.CurrentCallerInfo()
			log.Log("error", "[error] " + fmt.Sprintln("%s", err) + "\n" + info)
			continue
		}
		if !city_label{
			field_list = GetRelevantFieldIndex(x.(map[string]interface{}))
		}
		v := ProcessSingleRecord(x.(map[string]interface{}), field_list)
		h.PushHelper(v, max_length)
	}
	return h	    
}
