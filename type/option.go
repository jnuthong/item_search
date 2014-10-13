package option

import ("errors"
	// "runtime"
	// "reflect"
	"strconv"
	"github.com/jnuthong/item_search/utils")

type Option map[string] interface {}

func (d Option) IntKey(key string) (int, error) {	
	if t, err := d[key]; err {
		if t == nil{
			return 0, errors.New("[error] Empty value of key: " + key)
		}
		switch t.(type){
		case int:
			return t.(int), nil
		case *int:
			return t.(int), nil
		case string:
			return strconv.Atoi(t.(string))
		default:
			info := utils.CurrentCallerInfo()
			return 0, errors.New("[error] Invalid key:" + key + " parameter type from config " + info)
		} 
	}
	return 0, errors.New("[error] couldnt find key in the map variable")
}

func (d Option) StringKey(key string) (string, error){
	if t, err := d[key]; err {
		if t == nil{
			return "", errors.New("[error] Empty value of key: " + key)
		}
		switch t.(type){
		case int:
			return string(t.(int)), nil
		case float64:
			return strconv.FormatFloat(t.(float64), 'f', -1, 64), nil
		case string:
			return t.(string), nil
		default:
			info := utils.CurrentCallerInfo()
			return "", errors.New("[error] Invalid key:" + key + " parameter type from config " + info)
		} 
	}
	return "", errors.New("[error] couldnt find key in the map variable")
}

func (d Option) BoolKey(key string) (bool, error){
	if t, err := d[key]; err{
		if t == nil {
			return false, errors.New("[error] Empty value of key: " + key)
		}
		switch t.(type){
		case int:
			// TODO case of integer should rewrite
			if t.(int) > 0 {
				return true, nil
			}else{
				return false, nil
			}
		case bool:
			return t.(bool), nil
		case string:
			// accepted 1, t, T, TRUE, true
			// 	    0, f, F, FLASE, false
			return strconv.ParseBool(t.(string))
		default:
			info := utils.CurrentCallerInfo()
			return false, errors.New("[error] Invalid key:" + key + " parameter type from config " + info)
		}
	}
	return false, errors.New("[error] couldnt find key in the map variable")
}
