
package option

import ("errors"
	"runtime"
	"utils/utils")

type Option map[string] interface{}

func (d Options) IntKey(key string) (int, err errors) {
	if val, ok := d[key]; ok{
		switch vv := val.(type){
		case  float64:
				return int(vv), nil
		default:
			var info = utils.CurrentCallerInfo() 
			return nil, errors.New("[error] Invalid key:" + key + " parameter type from config " + info)
		}
	}
}

func (d Options) StringKey(key string) (string, err errors){
	if val, ok := d[key]; ok{
		switch vv := val.(type){
		case string:
			return vv, nil
		default:
			var info = utils.CurrentCallerInfo()
			return nil, errors.New("[error] Invalid key:" + key + " parameter type from config " + info)	
		}
	}
}

func (d Options) BoolKey(key string) (bool, err errors){
	if val, ok := d[key]; ok{
		switch vv := val.(type){
		case bool:
			return vv, nil
		default:
			var info = utils.CurrentCallerInfo()
			return nil, errors.New("[error] Invalid key:" + key + " parameter type from config " + info)
		}
	}
}
