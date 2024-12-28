package queue

import (
	"container/list"
	"errors"
	"log"
	"math/rand"
	"runtime"
	"runtime/debug"
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
		//listLen int
		//data *list.Element
		force bool
	)
	if forces != nil && len(forces) > 0 {
		force = forces[0]
	}

	//set closed
	f.Lock()
	f.closed = true
	f.Unlock()

	//force clear
	if force {
		//data opt with locker
		f.Lock()
		defer f.Unlock()
		//gc opt and reset list
		f.l.Init()
		atomic.StoreInt64(&f.enumCount, 0)
		runtime.GC()
		return
	}

	//process left list elements
	f.processLeftList()
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
	f.Lock()
	defer f.Unlock()
	f.l.Init()
	atomic.StoreInt64(&f.enumCount, 0)
	runtime.GC()
}

//check closed or not
func (f *List) Closed() bool {
	//data opt with locker
	f.Lock()
	defer f.Unlock()
	return f.closed
}

//get length
func (f *List) Len() int64 {
	//data opt with locker
	f.Lock()
	defer f.Unlock()
	return f.enumCount
}

//get element value
func (f *List) GetVal(e *list.Element) interface{} {
	return e.Value
}

//pop head
func (f *List) Pop() *list.Element {
	//data opt with locker
	f.Lock()
	defer f.Unlock()

	//check
	if f.enumCount <= 0 || f.closed {
		return nil
	}

	//pop data opt
	ele := f.l.Front()
	defer func() {
		f.l.Remove(ele)
		atomic.AddInt64(&f.enumCount, -1)
		if f.enumCount <= 0 {
			atomic.StoreInt64(&f.enumCount, 0)
			//gc opt
			runtime.GC()
		}
	}()

	//setup seed
	rand.Seed(time.Now().UnixNano())

	//rand force gc opt
	randVal := rand.Intn(DefaultTenThousandPercent)
	if randVal <= DefaultGcRate {
		runtime.GC()
	}

	return ele
}

//pop tail
func (f *List) Tail() *list.Element {
	//data opt with locker
	f.Lock()
	defer f.Unlock()

	//check
	if f.enumCount <= 0 || f.closed {
		return nil
	}

	//tail data opt
	ele := f.l.Back()
	defer func() {
		f.l.Remove(ele)
		atomic.AddInt64(&f.enumCount, -1)
		if f.enumCount <= 0 {
			atomic.StoreInt64(&f.enumCount, 0)
			//gc opt
			runtime.GC()
		}
	}()

	//setup seed
	rand.Seed(time.Now().UnixNano())

	//rand force gc opt
	randVal := rand.Intn(DefaultTenThousandPercent)
	if randVal <= DefaultGcRate {
		runtime.GC()
	}
	return ele
}

//join head
func (f *List) Join(val interface{}) error {
	if val == nil {
		return errors.New("invalid parameter")
	}

	//data opt with locker
	f.Lock()
	defer f.Unlock()
	if f.closed {
		return errors.New("list has closed")
	}

	//push data to front
	f.l.PushFront(val)
	atomic.AddInt64(&f.enumCount, 1)
	return nil
}

//push back
func (f *List) Push(val interface{}) error {
	if val == nil {
		return errors.New("invalid parameter")
	}

	//data opt with locker
	f.Lock()
	defer f.Unlock()
	if f.closed {
		return errors.New("list has closed")
	}

	//push data back
	f.l.PushBack(val)
	atomic.AddInt64(&f.enumCount, 1)
	return nil
}

///////////////
//private func
///////////////

//process left list
func (f *List) processLeftList()  {
	var (
		listLen int
		data *list.Element
	)

	//process left list elements with locker
	f.Lock()
	defer f.Unlock()
	for {
		listLen = f.l.Len()
		if listLen <= 0 {
			break
		}
		//pop front element
		data = f.l.Front()
		if data != nil && data.Value != nil {
			f.cbForConsumer(data.Value)
			f.l.Remove(data)
			atomic.AddInt64(&f.enumCount, -1)
			if f.enumCount <= 0 {
				atomic.StoreInt64(&f.enumCount, 0)
			}
		}
	}

	//gc opt and reset list
	f.l.Init()
	atomic.StoreInt64(&f.enumCount, 0)
	runtime.GC()
}

//run consume process
func (f *List) runConsumeProcess(rates ...float64) {
	var (
		rate float64
		m any = nil
	)

	//check
	if rates != nil && len(rates) > 0 {
		rate = rates[0]
	}
	if rate <= 0 {
		rate = DefaultListConsumeRate
	}

	//setup rate
	rateDuration := time.Duration(int64(rate * float64(time.Second)))

	//defer panic
	defer func() {
		if err := recover(); err != m {
			log.Printf("list.runConsumeProcess panic, err:%v, trace:%v\n",
				err, string(debug.Stack()))
		}
		//process left elements
		f.processLeftList()
	}()

	//setup seed
	rand.Seed(time.Now().UnixNano())

	//loop
	for {
		//check
		if f.closed {
			return
		}

		//pop front element and consume
		ele := f.Pop()
		if ele != nil && ele.Value != nil {
			f.cbForConsumer(ele.Value)
		}
		time.Sleep(rateDuration)
	}
}

