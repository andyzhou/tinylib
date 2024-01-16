package queue

import (
	"errors"
	"log"
	"sync"
)

/*
 * general queue worker
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//inter macro define
const (
	DefaultQueueSize = 1024
)

//inter type
type (
	interReq struct {
		req interface{} //origin input request
		resp chan []byte //json byte data
		needResp bool
	}
)

//face info
type Queue struct {
	queueSize int
	//reqChan chan interface{}
	reqChan chan interReq
	closeChan chan bool
	cbForReq func(data interface{}) ([]byte, error)
	closed bool
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
		close(f.closeChan)
	}
}

//get run queue size
func (f *Queue) GetQueueSize() int {
	return len(f.reqChan)
}

//send data, STEP-2
func (f *Queue) SendData(
	data interface{},
	needResponses...bool) ([]byte, error) {
	var (
		resp []byte //json bytes
		needResponse bool
	)
	//check
	if data == nil || data == "" {
		return nil, errors.New("invalid parameter")
	}
	if f.closed {
		return nil, errors.New("inter chan has closed")
	}
	if f.reqChan == nil || f.GetQueueSize() >= f.queueSize {
		return nil, errors.New("inter chan invalid or full")
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
		req.resp = make(chan []byte, 1)
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

//set callback, STEP-1
func (f *Queue) SetCallback(
	cb func(data interface{}) ([]byte, error)) bool {
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
		resp []byte //json bytes
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
		//close request chan
		close(f.reqChan)
		f.reqChan = nil
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
				//update closed switcher
				f.Lock()
				f.closed = true
				f.Unlock()
			}
			return
		}
	}
}
