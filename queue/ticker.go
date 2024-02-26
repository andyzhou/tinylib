package queue

import (
	"github.com/andyzhou/tinylib/util"
	"log"
	"time"
)

/*
 * general ticker queue worker
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - used for run auto ticker queue data
 */

//face info
type Ticker struct {
	tickDuration time.Duration
	tickChan chan struct{}
	closeChan chan bool
	cbForChecker func() error
	cbForQuit func()
	util.Util
}

//construct
//realDur := tickDuration * time.Second
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
	if f.closeChan != nil {
		f.closeChan <- true
	}
}

//check ticker is closed
func (f *Ticker) QueueClosed() bool {
	closed, _ := f.IsChanClosed(f.tickChan)
	return closed
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
func (f *Ticker) SetCheckerCallback(cb func() error) bool {
	if cb == nil {
		return false
	}
	f.cbForChecker = cb
	return true
}

///////////////
//private func
///////////////

//run main process
func (f *Ticker) runMainProcess() {
	var (
		m any = nil
	)

	//defer
	defer func() {
		if err := recover(); err != m {
			log.Printf("ticker.runMainProcess panic, err:%v\n", err)
		}
		//run last opt
		if f.cbForChecker != nil {
			//call cb
			f.cbForChecker()
		}
		//call cb for quit
		if f.cbForQuit != nil {
			f.cbForQuit()
		}
		//close tick chan
		close(f.tickChan)
	}()

	//start first ticker
	if f.tickChan != nil {
		sf := func() {
			chanIsClosed, _ := f.IsChanClosed(f.tickChan)
			if !chanIsClosed {
				f.tickChan <- struct{}{}
			}
		}
		sf()
	}

	//loop
	for {
		select {
		case <- f.tickChan:
			{
				if f.cbForChecker != nil {
					//call cb
					f.cbForChecker()
				}
				//send next ticker
				sf := func() {
					if f.tickChan != nil {
						chanIsClosed, _ := f.IsChanClosed(f.tickChan)
						if !chanIsClosed {
							f.tickChan <- struct{}{}
						}
					}
				}
				time.AfterFunc(f.tickDuration, sf)
			}
		case <- f.closeChan:
			return
		}
	}
}