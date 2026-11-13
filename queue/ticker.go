package queue

import (
	"errors"
	"github.com/andyzhou/tinylib/util"
	"log"
	"time"
	"sync"
)

/*
 * general ticker queue worker
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - used for run auto ticker queue data
 */

//face info
type Ticker struct {
	inputs       []interface{}
	tickDuration time.Duration
	tickChan     chan struct{}
	closeChan    chan bool
	closeOnce    sync.Once
	cbForChecker func(inputs ...interface{}) error
	cbForQuit    func()
	util.Util
}

//construct
//tickDurations -> seconds value, default 1 second
func NewTicker(tickDurations ...float64) *Ticker {
	//check and set ticker rate
	defaultDuration := float64(1) //default 1 second
	if tickDurations != nil && len(tickDurations) > 0 {
		if tickDurations[0] > 0 {
			defaultDuration = tickDurations[0]
		}
	}
	durationTicker := time.Duration(int64(defaultDuration * float64(time.Second)))

	//self init
	this := &Ticker{
		inputs: []interface{}{},
		tickDuration: durationTicker,
		tickChan: make(chan struct{}, 1),
		closeChan: make(chan bool, 1),
	}

	//spawn main process
	go this.runMainProcess()
	return this
}

//quit
func (f *Ticker) Quit() {
	f.closeOnce.Do(func() {
		close(f.closeChan)   // 只关一次
	})
}

//check ticker is closed
func (f *Ticker) QueueClosed() bool {
	closed, _ := f.IsChanClosed(f.tickChan)
	return closed
}

//get current ticker duration
func (f *Ticker) GetDuration() float64 {
	return f.tickDuration.Seconds()
}

//update ticker duration
func (f *Ticker) UpdateDuration(tickDuration float64) error {
	//check
	if tickDuration <= 0 {
		return errors.New("invalid parameter")
	}

	//set new duration
	newDuration := time.Duration(int64(tickDuration * float64(time.Second)))
	f.tickDuration = newDuration
	return nil
}

//set callback for process quit
func (f *Ticker) SetQuitCallback(cb func()) bool {
	if cb == nil {
		return false
	}
	f.cbForQuit = cb
	return true
}

//set callback for data opt, STEP-1
func (f *Ticker) SetCheckerCallback(cb func(inputs ...interface{}) error, inputs ...interface{}) bool {
	if cb == nil {
		return false
	}
	f.cbForChecker = cb
	f.inputs = inputs
	return true
}

///////////////
//private func
///////////////

//run main process
func (f *Ticker) runMainProcess() {
	var (
		m any = nil
		isQuitting bool
	)

	defer func() {
		if err := recover(); err != m {
			log.Printf("ticker.runMainProcess panic, err:%v\n", err)
		}

		//only normal quit to execute
		if isQuitting {
			if f.cbForChecker != nil {
				f.cbForChecker(f.inputs...)
			}
			if f.cbForQuit != nil {
				f.cbForQuit()
			}
		}

		// 只有 runMainProcess 关闭 tickChan（唯一所有者）
		close(f.tickChan)
	}()

	//start first ticker
	select {
	case f.tickChan <- struct{}{}:
	case <-f.closeChan:
		isQuitting = true
		return
	}

	//loop
	for {
		select {
		case <-f.tickChan:
			if f.cbForChecker != nil {
				f.cbForChecker(f.inputs...)
			}
			time.Sleep(f.tickDuration)
			// 用 select 同时监听 closeChan，避免 send on closed
			select {
			case f.tickChan <- struct{}{}:
			case <-f.closeChan:
				isQuitting = true
				return
			}
		case <- f.closeChan:
			isQuitting = true
			return
		}
	}
}