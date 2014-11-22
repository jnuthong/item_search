package logic

import (
	"fmt"
	"strings"
	"bufio"
	"errors"
	"encoding/json"
	"gopkg.in/mgo.v2"
	"os"
	"runtime"
	"path"

	"github.com/barakmich/glog"

	"github.com/fatih/structs"
	"github.com/jnuthong/item_search/db/mongo"
	"github.com/jnuthong/item_search/utils"
	"github.com/jnuthong/item_search/utils/log"
	"github.com/jnuthong/item_search/utils/multiChannel"
	"github.com/jnuthong/item_search/conf"
	// "github.com/jnuthong/item_search/utils/iterator"
)

var (
	subject = 0
	predicate = 1
	object = 2

	// path length constraint
	SUB_LENGTH = 3
	PRE_LENGTH = 3

	// temporarily
	_, filename, _, _ = runtime.Caller(0)
	dir = path.Dir(filename)
	tmpIn = dir + "/tmpIn/" 
	tmpOut = dir + "/tmpOut/"
)

// MultipleChannel_ProcessFile(name string, inputFile string, outputFile string, tmpInDir string, tmpOutDir string, call function)

func InsertDocLineHanler(c *mgo.Collection, line string) error {
	line = strings.Trim(line, "\n")
	tuple, count := utils.PathParser(line, "\t")
	if (count != 3){
		info := utils.CurrentCallerInfo()
		glog.Fatalln("[error] Line Element is not equal to 3, when split the line with delimiter-tab" + info)
	}
	// handle three entity in a single line, #1 tuple represent index entity $subject
	//					#2 tuple standfor predicate
	// 					#3 tuple process as main content 
	
	// process the subject path
	subject_list, sub_length := utils.PathParser(tuple[subject], "/")
	if (sub_length != SUB_LENGTH) {
		return errors.New("[error] Error Line ELEMENT:subject is not the specify path lenght, subject value: " + tuple[subject])
	}

	// process the predicate path
	predicate_list, pre_length := utils.PathParser(tuple[predicate], "/")
	if(pre_length != PRE_LENGTH){
		return errors.New("[error] Error Line ELEMENT:predicate is not satisfied the specify path length, predicate value: " + tuple[predicate])
	}
	
	// TODO list - more elegant way to handle the path
	// 1) find before insert
	// 2) if could find any object with the specify path, then insert new document	
	switch predicate_list[pre_length - 1]{
	case "property":
		doc := make(map[string]interface{})
		var data []utils.Element
		err := json.Unmarshal([]byte(tuple[object]), &data)
		if err != nil{
			info := utils.CurrentCallerInfo()	
			log.Log("error", info + fmt.Sprintf("%s", err))
			return nil
		}
		for index := range data{
			utils.UpdateMap(data[index], &doc)	
		}
		
		path := utils.GeneElement("path", "hello")
		utils.UpdateMap(path, &doc)	
		utils.UpdateMap(utils.GeneElement("id", subject_list[sub_length - 1]), &doc)
		// info, err := mongo.UpdateOrInsert_DocByDocID(c, subject_list[sub_length - 1], doc)
		// log.Log("info", "[Info] " + fmt.Sprintf("%s", info))
		err = mongo.InsertDoc(c, doc)
		if err != nil{
			log.Log("error", "[error] " + fmt.Sprintf("%s", err))
		}
		return nil
	//case "tag":
	//	var data []utils.Comment_tag	
	//	err := json.Unmarshal([]byte(tuple[object]), &data)
	//	if err != nil{
	//		log.Log("error", fmt.Sprintf("%s", err))
	//		return nil
	//	}
	//	info, err := mongo.UpdateOrInsert_FieldByDocID(c, subject_list[sub_length - 1], predicate_list[pre_length - 1], data)
	//	log.Log("info", "[Info] " + fmt.Sprintf("%s", err) + "\n")
	default:
		return nil
	}			
	return nil
}

func InsertDocWithFile(inputPath string, outputPath string, c *mgo.Collection) error{
	inputFile, err := os.Open(inputPath)
	if err != nil{
		log.Log("error", fmt.Sprintf("%s", err))
	}
	defer inputFile.Close()
	
	scanner := bufio.NewReader(inputFile)
	for {
		text, err := scanner.ReadString('\n')
		if err == nil{
			// TODO do something to each line of the file
			info := InsertDocLineHanler(c, text)
			if info != nil{
				log.Log("error", fmt.Sprintf("%s", err))
			}
		}else{
			break
		}
	}	
	return nil
}

// multi-chan for insertDocWithFile with a single connected client
func MultiChan_InsertDocWithFile(inputPath string, outputPath string, c *mgo.Collection) error {
	insertDoc := func(input string, output string) error {
		   	return InsertDocWithFile(input, output, c)
		   }
	if !utils.Exists(tmpIn){
		err := os.Mkdir(tmpIn, 0777)
		if err != nil{
			info := utils.CurrentCallerInfo()
			log.Log("error", "[error] " + fmt.Sprintf("%s", err) + "\n" + info)
			os.Exit(1)
		}
	}

	if !utils.Exists(tmpOut){
		err := os.Mkdir(tmpOut, 0777)
		if err != nil{
			info := utils.CurrentCallerInfo()
			log.Log("error", "[error] " + fmt.Sprintf("%s", err) + "\n" + info)
			os.Exit(1)
		}
	}	

	err := multiChannel.MultipleChannel_ProcessFile("part", inputPath, outputPath, tmpIn, tmpOut, insertDoc)	
	return err
}

func MultiClient_InsertDocWithFile(inputPath string, outputPath string, c conf.Configuration) error{
	mongodb_path := "mongodb://" + c.User + ":" + c.Password + "@" +  c.Address + ":" + c.Port + "/" + c.DefaultDBName
	insertDoc := func(input string, output string) error{
		x := structs.New(c).Map()
		db, err := mongo.ConnectMongoDBCollection(mongodb_path, x)
		if err != nil{
			info := utils.CurrentCallerInfo() 
			log.Log("error", "[error] " + fmt.Sprintf("%s", err) + "\n" + info)
		}
		c := db.GetCollection()
		return InsertDocWithFile(input, output, c)
	}
	err := multiChannel.MultipleChannel_ProcessFile("part", inputPath, outputPath, tmpIn, tmpOut, insertDoc)	
	return err
}
