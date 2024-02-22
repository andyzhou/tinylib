package queue

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

/*
 * general list worker
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//inter macro define
const (
	listConsumeRate = 1 //xx seconds
)

//face info
type List struct {
	l *list.List
	cbForConsumer func(interface{}) error
	closed bool
	sync.RWMutex
}

//construct
func NewList() *List {
	this := &List{
		l: list.New(),
	}
	return this
}

//quit
func (f *List) Quit() {
	var (
		listLen int
		data *list.Element
	)
	//set closed
	f.closed = true
	time.Sleep(time.Second)

	//process left list elements
	for {
		listLen = f.l.Len()
		if listLen <= 0 {
			break
		}
		//pop front element
		data = f.l.Front()
		f.cbForConsumer(data.Value)
		f.l.Remove(data)
	}

	//reset list
	f.l.Init()
}

//set consumer
func (f *List) SetConsumer(cb func(interface{}) error, rates ...float64) {
	//check
	if cb == nil || f.cbForConsumer != nil {
		return
	}

	//set and run consume process
	f.cbForConsumer = cb
	go f.runConsumeProcess(rates...)
}

//clear
func (f *List) Clear() {
	f.l.Init()
}

//get length
func (f *List) Len() int {
	return f.l.Len()
}

//get element value
func (f *List) GetVal(e *list.Element) interface{} {
	return e.Value
}

//pop head
func (f *List) Pop() *list.Element {
	ele := f.l.Front()
	defer f.l.Remove(ele)
	return ele
}

//pop tail
func (f *List) Tail() *list.Element {
	ele := f.l.Back()
	defer f.l.Remove(ele)
	return ele
}

//join head
func (f *List) Join(val interface{}) error {
	if val == nil {
		return errors.New("invalid parameter")
	}
	f.l.PushFront(val)
	return nil
}

//push back
func (f *List) Push(val interface{}) error {
	if val == nil {
		return errors.New("invalid parameter")
	}
	f.l.PushBack(val)
	return nil
}

///////////////
//private func
///////////////

//run consume process
func (f *List) runConsumeProcess(rates ...float64) {
	var (
		rate float64
		listLen int
		data *list.Element
	)
	if rates != nil && len(rates) > 0 {
		rate = rates[0]
	}
	if rate <= 0 {
		rate = listConsumeRate
	}
	rateDuration := time.Duration(rate) * time.Second

	//loop
	for {
		//check
		if f.closed {
			return
		}

		//list data opt
		listLen = f.l.Len()
		if listLen <= 0 {
			time.Sleep(rateDuration)
			continue
		}

		//pop front element
		data = f.l.Front()
		f.cbForConsumer(data.Value)
		f.l.Remove(data)
	}
}

