package log

import (
	"os"
	// "io/ioutil"
	"time"

	"github.com/jnuthong/item_search/utils"

)

var (
	t = time.Now()
	now = t.Format("20060102 15:04:00")
	time_now, count = utils.PathParser(now, " ")
	time_split, _ = utils.PathParser(time_now[1], ":")
	current_time = time_now[0] + "_" + time_split[0]
	home = "/home/users/hongjianbin/work/log/"	
)

func Log(file_type string, info string) {
	file_path := home + file_type + "." + current_time
	f, err := os.OpenFile(file_path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.WriteString(info + "\n"); err != nil{
		// panic(err)
	}
}
