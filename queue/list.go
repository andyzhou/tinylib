package queue

import (
	"container/list"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * general list worker [developing]
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type List struct {
	l             *list.List
	cbForConsumer func(interface{}) error
	enumCount     int64
	closed        bool
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
func (f *List) Quit(forces ...bool) {
	var (
		listLen int
		data *list.Element
		force bool
	)
	if forces != nil && len(forces) > 0 {
		force = forces[0]
	}

	//set closed
	f.closed = true
	time.Sleep(time.Second)

	if force {
		//just clean list
		f.l.Init()
		return
	}

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
//real duration = rate * time.second
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
	atomic.StoreInt64(&f.enumCount, 0)
	runtime.GC()
}

//get length
func (f *List) Len() int64 {
	return f.enumCount
}

//get element value
func (f *List) GetVal(e *list.Element) interface{} {
	return e.Value
}

//pop head
func (f *List) Pop() *list.Element {
	if f.enumCount <= 0 {
		return nil
	}
	ele := f.l.Front()
	defer func() {
		f.l.Remove(ele)
		atomic.AddInt64(&f.enumCount, -1)
		if f.enumCount < 0 {
			atomic.StoreInt64(&f.enumCount, 0)
		}
	}()
	return ele
}

//pop tail
func (f *List) Tail() *list.Element {
	if f.enumCount <= 0 {
		return nil
	}
	ele := f.l.Back()
	defer func() {
		f.l.Remove(ele)
		atomic.AddInt64(&f.enumCount, -1)
		if f.enumCount < 0 {
			atomic.StoreInt64(&f.enumCount, 0)
		}
	}()
	return ele
}

//join head
func (f *List) Join(val interface{}) error {
	if val == nil {
		return errors.New("invalid parameter")
	}
	f.l.PushFront(val)
	atomic.AddInt64(&f.enumCount, 1)
	return nil
}

//push back
func (f *List) Push(val interface{}) error {
	if val == nil {
		return errors.New("invalid parameter")
	}
	f.l.PushBack(val)
	atomic.AddInt64(&f.enumCount, 1)
	return nil
}

///////////////
//private func
///////////////

//run consume process
func (f *List) runConsumeProcess(rates ...float64) {
	var (
		rate float64
		data *list.Element
		needGc bool
	)
	if rates != nil && len(rates) > 0 {
		rate = rates[0]
	}
	if rate <= 0 {
		rate = DefaultListConsumeRate
	}

	rateDuration := time.Duration(int64(rate * float64(time.Second)))

	//loop
	for {
		//check
		if f.closed {
			return
		}

		//list data opt
		if f.enumCount <= 0 {
			if needGc {
				runtime.GC()
				needGc = false
			}
			time.Sleep(rateDuration)
			continue
		}

		//pop front element
		data = f.l.Front()
		if data.Value != nil {
			f.cbForConsumer(data.Value)
			if !needGc {
				needGc = true
			}
		}
		f.l.Remove(data)
		atomic.AddInt64(&f.enumCount, -1)
		if f.enumCount < 0 {
			atomic.StoreInt64(&f.enumCount, 0)
		}
	}
}

