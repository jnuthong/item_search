package main

import (
	"encoding/json"
	"flag"
	"os"
	"errors"
	"fmt"

	"io/ioutil"
	
	"github.com/fatih/structs"
	"github.com/barakmich/glog"
	"github.com/jnuthong/item_search/logic"
	"github.com/jnuthong/item_search/utils"	
	"github.com/jnuthong/item_search/db/mongo"
	"github.com/jnuthong/item_search/utils/log"
)

var (
	config =  flag.String("config",  "/home/users/hongjianbin/.jumbo/lib/go/site/src/github.com/jnuthong/item_search/" + "conf.go", "config file path")	
	inputFile = flag.String("inputFile", "", "file to load, the data format in the file should following the FILELOAD.README file")
	logDir = flag.String("logDir", "", "log file directory")
)

type Configuration struct {
	DefaultDBName string
	DefaultMongIndexKey string
	DefaultCollection string
	User string
	Address string
	Port string
	Password string

	Index []map[string]interface{}
	Other string
}

func LoadConfigurationFile(path string) (*Configuration, error) {
	// fmt.Println(path)
	// file, err := os.Open(path)
	file, e := ioutil.ReadFile(path)
	if e != nil{
		info := utils.CurrentCallerInfo()
		conf := new(Configuration)
		return conf, errors.New("[error] Couldnt Load Configuration File " + info)	
	}

	conf := new(Configuration)
	err := json.Unmarshal(file, &conf)
	if err != nil{
		info := utils.CurrentCallerInfo()
		glog.Fatalln(err)
		return conf, errors.New("[error] Couldnt Load Configuration Json " + info)
	}
	return conf, nil
}

func main(){
	flag.Parse()

	/* TODO should setting up the default configuration file	
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil{
		glog.Fatalln(err)
	}
	*/

	var config_path string // config file path
	if *config != ""{
		config_path = *config
	}

	/*
	var input_path string // input file path

	if err != nil{
		glog.Fatalln(err)
	}
	*/
	
	conf, err := LoadConfigurationFile(config_path)
	x := structs.New(conf).Map()
	if err != nil{
		fmt.Println(err)
		os.Exit(0)
	}
	if *inputFile == ""{
		fmt.Println("[error] Expect Parameter --inputFile been set")
		os.Exit(-1)
	}

	// TODO should consider another way to config the index list
	mongodb_path := "mongodb://" + conf.User + ":" + conf.Password + "@" +  conf.Address + ":" + conf.Port + "/" + conf.DefaultDBName
	log.Log("info", "[info] Connect Mongo Address: " + mongodb_path)
	db_instance, err := mongo.CreateMongoDBCollection(mongodb_path, x)	
	fmt.Println(db_instance)
	c := db_instance.GetCollection()
	fmt.Println(*inputFile)
	err = logic.InsertDocWithFile(*inputFile, "", c)
	if err != nil {
		fmt.Println(err)
	}
}
