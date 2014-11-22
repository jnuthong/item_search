// Author: hongjianbin@baidu.com

package pheap

import (
	//"container/heap"
	"fmt"	

	"github.com/jnuthong/item_search/utils/log"
)

var (
	max_length = 100		
)

type Element struct {
	Value float64
	Entity interface{}
}

type Heap []Element

// Heap function part

func (h Heap) Len() int 		{ return len(h) }
func (h Heap) Less(i, j int) bool 	{ return h[i].Value < h[j].Value }
func (h Heap) Swap(i, j int) 		{ h[i], h[j] = h[j], h[i]}

func (h *Heap) PushHelper(x Element, max_length int) {
	max_length = max_length
	h.Push(x)
}

func (h *Heap) Push(x interface{}) {
	n := len(*h)
	if (n < max_length){
		*h = append(*h, x.(Element))
	}else{
		pre_x := (*h)[0]
		if pre_x.Value < x.(Element).Value {
			v := h.Pop()
			*h = append(*h, x.(Element))
			log.Log("info", "[Info] " + fmt.Sprintln("%s", v))
		}
	} 
}

func (h *Heap) Pop() interface{}{
	old 	:= *h
	n 	:= len(old)
	x	:= old[0]
	*h	= old[0:n-1]
	return x
}
