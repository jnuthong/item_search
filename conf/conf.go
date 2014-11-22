package conf

import (
	"encoding/json"
	"errors"
	"fmt"

	"io/ioutil"

	"github.com/jnuthong/item_search/utils"
	"github.com/barakmich/glog"
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

/*
func (c Configuration) User() string{
	return c.User
}

func (c Configuration) Port() string{
	return c.Port
}

// More elegant way should be consider
func (c Configuration) Password() string{
	return c.Password
}

func (c Configuration) DefaultDBName() string{
	return c.DefaultDBName
}

func (c Configuration) Address() string{
	return c.Address()
}
*/

func LoadConfigurationFile(path string) (*Configuration, error) {
	// fmt.Println(path)
	// file, err := os.Open(path)
	file, e := ioutil.ReadFile(path)
	if e != nil{
		info := utils.CurrentCallerInfo()
		conf := new(Configuration)
		glog.Fatalln(e)
		return conf, errors.New("[error] Couldnt Load Configuration Json " + fmt.Sprintf("%s", e) + "\n" + info)
	}

	conf := new(Configuration)
	err := json.Unmarshal(file, &conf)
	if err != nil{
		info := utils.CurrentCallerInfo()
		glog.Fatalln(err)
		return conf, errors.New("[error] Couldnt Load Configuration Json " + fmt.Sprintf("%s", err) + "\n" + info)
	}
	return conf, nil
}

