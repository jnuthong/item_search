package iterator

import (
	"errors"		
)

type Element struct{
	key string
	value interface{}
	type string
	label string
}

func GeneElement(key string, value interface{}, type string, label string) *Element{
	i := Element{key: key, value: value, type: type, label: label}
	return &i
}

func (opt *Element) GetElementAttribute(key string) value interface{} {
	if key == "key"{
		return opt.key
	}
	if key == "value"{
		return opt.value
	}
	if key == "type"{
		return opt.type
	}
	if key == "label"{
		return opt.label
	}
}
