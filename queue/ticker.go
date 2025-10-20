package queue

import (
	"errors"
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
	inputs       []interface{}
	tickDuration time.Duration
	tickChan     chan struct{}
	closeChan    chan bool
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
	//close closeChan first
	if f.closeChan != nil {
		select {
		case f.closeChan <- true:
		default: //ignore block
		}
		close(f.closeChan)
	}

	//wait awhile to let goroutine quit
	time.Sleep(10 * time.Millisecond)

	//close tickChan
	if f.tickChan != nil {
		close(f.tickChan)
	}
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

		//clean channels
		if f.tickChan != nil {
			isClosed, _ := f.IsChanClosed(f.tickChan)
			if !isClosed {
				close(f.tickChan)
				f.tickChan = nil
			}
		}
		if f.closeChan != nil {
			isClosed, _ := f.IsChanClosed(f.closeChan)
			if !isClosed {
				close(f.closeChan)
				f.closeChan = nil
			}
		}
	}()

	//start first ticker
	if f.tickChan != nil && !f.QueueClosed() {
		f.tickChan <- struct{}{}
	}

	//loop
	for {
		select {
		case <- f.tickChan:
			{
				if isQuitting {
					//quiting, just return
					return
				}
				if f.cbForChecker != nil {
					//call cb
					f.cbForChecker(f.inputs...)
				}
				//send next tick
				time.Sleep(f.tickDuration)
				if f.tickChan != nil && !f.QueueClosed() && !isQuitting {
					select {
					case f.tickChan <- struct{}{}:
					default: //ignore block
					}
				}
			}
		case <- f.closeChan:
			isQuitting = true
			return
		}
	}
}