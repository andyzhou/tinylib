package queue

import (
	"log"
	"time"
)

/*
 * general ticker queue worker
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - used for run auto ticker queue data
 */

//inter macro define
const (
	DefaultTickTimer = time.Second
)

//face info
type Ticker struct {
	tickDuration time.Duration
	tickChan chan struct{}
	closeChan chan bool
	cbForChecker func() error
}

//construct
func NewTicker(tickDurations ...time.Duration) *Ticker {
	//set ticker rate
	defaultTicker := DefaultTickTimer
	if tickDurations != nil && len(tickDurations) > 0 {
		if tickDurations[0] > 0 {
			defaultTicker = tickDurations[0]
		}
	}
	//self init
	this := &Ticker{
		tickDuration: defaultTicker,
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
		close(f.closeChan)
	}
}

//set callback, STEP-1
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
		//close relate chan
		close(f.tickChan)
		close(f.closeChan)
	}()

	//start first ticker
	if f.tickChan != nil {
		sf := func() {
			f.tickChan <- struct{}{}
		}
		time.AfterFunc(f.tickDuration, sf)
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
						f.tickChan <- struct{}{}
					}
				}
				time.AfterFunc(f.tickDuration, sf)
			}
		case <- f.closeChan:
			return
		}
	}
}