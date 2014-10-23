package multiChannel
// package utils

import (// "fmt"
	"errors"
	"os/exec"
	"math"
	"strings"
	"strconv"

	"github.com/jnuthong/item_search/utils"
)

type function func(inputPath string, outputPath string) error

var (
	MAX_CHANNEL = 64
	MAX_LINE = 100000	
)

func MultipleChannel_ProcessFile(name string, inputFile string, outputFile string, tmpInDir string, tmpOutDir string, call function) error{
	if ok := utils.Exists(inputFile); !ok{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Unrecognized inputFile name: " + inputFile + info)
	}

	if ok := utils.Exists(outputFile); !ok {
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Unrecognized outputFile name: " + outputFile + info)
	}

	if ok := utils.Exists(tmpInDir);!ok {
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Unrecognized tmpIndir direcotry: " + tmpInDir + info)
	}
	
	if ok := utils.Exists(tmpOutDir); !ok{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Unrecognized tmpOutDir directory: " + tmpOutDir + info)
	}
	// remove all files in tmpInDir
       	cmd := exec.Command("rm", tmpInDir + "*")
	err := cmd.Run()
	if err != nil{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Couldnot remove the file in the tmpInDir directory: " + tmpInDir + info)
	}	
	// remove all files in tmpOutDir
	cmd = exec.Command("rm", tmpOutDir + "*")
	err = cmd.Run()
	if err != nil{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Couldnt Remove the file in the tmpOutDir Directory: " + tmpOutDir + info)
	}
	
	lines, err := utils.CountFileLines(inputFile)
	if err != nil{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Couldnt Count inputFile lines with File: " + inputFile + info)
	}
	pieces_num := math.Ceil(float64(lines) / float64(MAX_LINE)); 
	if pieces_num > float64(MAX_CHANNEL){
		pieces_num = float64(MAX_CHANNEL)
	}
	
	// split big file input small piece file
	cmd = exec.Command("split", "-l", strconv.Itoa(int(pieces_num)), "-a4", "-d", inputFile, tmpInDir + name)
	err = cmd.Run()
	if err != nil{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Could not split inputFile " + info)
	}
	
     	out, err := exec.Command("ls", tmpInDir).Output()
	if err != nil{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Could not get relevant small file" + info)
	}
	s := string(out[:]) 
	file_names := strings.Split(strings.Trim(s, "\n\r"), "\n")
	
	message := make(chan string, len(file_names))
	for i := 0; i < len(file_names); i++ {
		go func (){
			call(tmpInDir + file_names[i], tmpOutDir + file_names[i])
			message <- "done"
		}()
	}

	for i := 0; i < len(file_names); i++{
		<- message
	}
	return nil
}
