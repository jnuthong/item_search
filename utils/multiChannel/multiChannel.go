package multiChannel
// package utils

import ("fmt"
	"errors"
	// "os"
	"os/exec"
	"math"
	"strings"
	"strconv"
	"time"

	"github.com/jnuthong/item_search/utils"
	// "github.com/jnuthong/item_search/utils/log"
)

type function func(inputPath string, outputPath string) error

var (
	MAX_CHANNEL = 64
	MAX_LINE = 100000	
)

func MultipleChannel_ProcessFile(name string, inputFile string, outputFile string, tmpInDir string, tmpOutDir string, call function) error{
	if ok := utils.Exists(inputFile); !ok{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Unrecognized inputFile name: " + inputFile + "\n" + info)
	}
	/*
	if ok := utils.Exists(outputFile); !ok {
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Unrecognized outputFile name: " + outputFile + info)
	}
	*/
	if ok := utils.Exists(tmpInDir);!ok {
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Unrecognized tmpIndir direcotry: " + tmpInDir + "\n" + info)
	}
	
	if ok := utils.Exists(tmpOutDir); !ok{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Unrecognized tmpOutDir directory: " + tmpOutDir + "\n" + info)
	}

	// remove all files in tmpInDir
	var out []byte
	err := utils.DelAll_DirFiles(tmpInDir)
	if err != nil{
		info := utils.CurrentCallerInfo()
		fmt.Println("[warn] Couldnt remove the file in the tmpInDir directory: " + tmpInDir + "\n" + 
					fmt.Sprintf("%s", err) + info)
	}	

	// remove all files in tmpOutDir
	err = utils.DelAll_DirFiles(tmpOutDir)
	if err != nil{
		info := utils.CurrentCallerInfo()
		fmt.Println("[warn] Couldnt Remove the file in the tmpOutDir Directory: " + tmpOutDir + "\n" + 
					fmt.Sprintf("%s", err) + info)

	}
	
	lines, err := utils.CountFileLines(inputFile)
	if err != nil{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Couldnt Count inputFile lines with File: " + inputFile + fmt.Sprintf("%s", err) + "\n" + info)
	}
	fmt.Println("InputFile total lines : ", lines)
	pieces_num := math.Ceil(float64(lines) / float64(MAX_LINE)); 
	per_line := MAX_LINE
	if pieces_num > float64(MAX_CHANNEL){
		pieces_num = float64(MAX_CHANNEL)
		per_line = int(math.Ceil(float64(lines) / float64(pieces_num)))
	}
	
	// split big file input small piece file
	cmd := exec.Command("split", "-l", strconv.Itoa(int(per_line)), "-a4", "-d", inputFile, tmpInDir + name)
	out, err = cmd.CombinedOutput()
	if err != nil{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Could not split inputFile " + fmt.Sprintf("%s", err) + "\n" + info)
	}
	
     	out, err = exec.Command("ls", tmpInDir).Output()
	if err != nil{
		info := utils.CurrentCallerInfo()
		return errors.New("[error] Could not get relevant small file" + info)
	}
	s := string(out[:]) 
	file_names := strings.Split(strings.Trim(s, "\n\r"), "\n")

	message := make(chan string, len(file_names))
	for i := 0; i < len(file_names); i++ {
		go func (){
			err := call(tmpInDir + file_names[i], tmpOutDir + file_names[i])
			if err != nil{
				fmt.Println(err)
			}
			message <- "done"
		}()
		time.Sleep(1 * time.Second)
	}

	for i := 0; i < len(file_names); i++{
		<- message
	}
	return nil
}
