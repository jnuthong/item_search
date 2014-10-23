package iterator

import (
	"errors"		
)

type Element struct{
	id string
	key string
	value interface{}
	datatype string
	label string	// maintain for further using
	parent string
	path string
}

type Child struct{
	id string
	name string
	label string
} 

type Iterator struct{
	Next() *Element
	Previous() *Element
}

func GeneElement(key string, value interface{}, datatype string, label string) *Element{
	i := Element{key: key, value: value, datatype: datatype, label: label}
	return &i
}

func (opt *Element) GetElementAttribute(key string) interface{} {
	if key == "key"{
		return opt.key
	}
	if key == "value"{
		return opt.value
	}
	if key == "datatype"{
		return opt.datatype
	}
	if key == "label"{
		return opt.label
	}
	return errors.News("[error] Couldnt find the relevant key: " + key)
}
