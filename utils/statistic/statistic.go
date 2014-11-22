package statistic

import (
	"fmt"
	"gopkg.in/mgo.v2"

	// "github.com/jnuthong/item_search/utils/iterator"		
	"github.com/jnuthong/item_search/utils"
)

// compute the histogram graph for the given combined-fields list
func Hist(iter *mgo.Iter, fields [][]string){
	// fields in the following format: [["field01", "field02"], ["field03"], ["field04"]]
	var result map[string]interface{}
	var hist = make([]map[string]int, len(fields))

	for iter.Next(&result) {
		for i := range fields {

			switch len(fields[i]){
			case 0: continue
			case 1:
				tmp, count := utils.PathParser(fields[i][0], ".") 
				if count != 0{
					continue
				}
				v := utils.PathLook(tmp, result)
				key := utils.ToString(v)
				hist[i][key] += 1
			default:
				var key string
				for j := range fields[i]{
					tmp, count := utils.PathParser(fields[i][j], ".")
					if count != 0{
						continue
					}
					v := utils.PathLook(tmp, result)
					key += utils.ToString(v)	   
				}
				hist[i][key] += 1
			}
		}
	}

	fmt.Println(hist)
}
