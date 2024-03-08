package queue

import (
	"errors"
	"fmt"
	"github.com/andyzhou/tinylib/util"
	"log"
	"sync"
)

/*
 * general queue worker
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */



//inter type
type (
	interReq struct {
		req interface{} //origin input request
		resp chan interface{}
		needResp bool
	}
)

//face info
type Queue struct {
	queueSize int
	reqChan chan interReq
	closeChan chan bool
	cbForReq func(data interface{}) (interface{}, error)
	cbForQuit func()
	util.Util
	sync.RWMutex
}

//construct
func NewQueue(queueSizes ...int) *Queue {
	//set queue size
	queueSize := DefaultQueueSize
	if queueSizes != nil && len(queueSizes) > 0 {
		if queueSizes[0] > 0 {
			queueSize = queueSizes[0]
		}
	}
	//self init
	this := &Queue{
		queueSize: queueSize,
		reqChan: make(chan interReq, queueSize),
		closeChan: make(chan bool, 1),
	}
	//spawn main process
	go this.runMainProcess()
	return this
}

//quit
func (f *Queue) Quit() {
	if f.closeChan != nil {
		f.closeChan <- true
	}
}

//check queue is closed
func (f *Queue) QueueClosed() bool {
	closed, _ := f.IsChanClosed(f.reqChan)
	return closed
}

//get run queue size
func (f *Queue) GetQueueSize() int {
	return len(f.reqChan)
}

//send data, STEP-2
func (f *Queue) SendData(
	data interface{},
	needResponses...bool) (interface{}, error) {
	var (
		resp interface{}
		needResponse bool
	)
	//check
	if data == nil || data == "" {
		return nil, errors.New("invalid parameter")
	}
	if f.reqChan == nil {
		return nil, errors.New("inter chan is nil")
	}

	//check active queue size
	queueSize := f.GetQueueSize()
	if queueSize >= f.queueSize {
		return nil, fmt.Errorf("inter queue size %v up to limit %v", queueSize, f.queueSize)
	}

	//detect
	if needResponses != nil && len(needResponses) > 0 {
		needResponse = needResponses[0]
	}

	//setup inter request
	req := interReq{
		req: data,
		needResp: needResponse,
	}
	if needResponse {
		req.resp = make(chan interface{}, 1)
	}

	//send to chan with async mode
	select {
	case f.reqChan <- req:
	}

	if needResponse {
		//wait for response
		resp, _ = <- req.resp
	}

	return resp, nil
}

//set callback for process quit
func (f *Queue) SetQuitCallback(cb func()) bool {
	if cb == nil {
		return false
	}
	f.cbForQuit = cb
	return true
}

//set callback for data opt, STEP-1
func (f *Queue) SetCallback(
	cb func(data interface{}) (interface{}, error)) bool {
	if cb == nil {
		return false
	}
	f.cbForReq = cb
	return true
}

///////////////
//private func
///////////////

//process left data in chan
func (f *Queue) processChanLeftData() {
	var (
		data interface{}
		isOk bool
	)
	//check chan
	if f.reqChan == nil || len(f.reqChan) <= 0 {
		return
	}
	//process one by one
	for {
		//pick data from chan
		data, isOk = <- f.reqChan
		if !isOk || data == nil {
			break
		}
		if f.cbForReq != nil {
			f.cbForReq(data)
		}
	}
}

//run main process
func (f *Queue) runMainProcess() {
	var (
		orgReq interReq
		resp interface{}
		isOk bool
		m any = nil
	)

	//defer
	defer func() {
		if err := recover(); err != m {
			log.Printf("queue.runMainProcess panic, err:%v\n", err)
		}

		//process left data in chan
		f.processChanLeftData()

		//call cb for quit
		if f.cbForQuit != nil {
			f.cbForQuit()
		}
	}()

	//loop
	for {
		select {
		case orgReq, isOk = <- f.reqChan:
			{
				if isOk && &orgReq != nil && f.cbForReq != nil {
					resp, _ = f.cbForReq(orgReq.req)
					if orgReq.needResp {
						orgReq.resp <- resp
					}
				}
			}
		case <- f.closeChan:
			{
				return
			}
		}
	}
}
