package queue

import (
	"errors"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * general concurrency worker
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//inter type
//gen and bind obj ticker, only one works
type (
	SonWorker struct {
		workerId int32
		bindObjs map[int64]interface{}
		queue *Queue
		ticker *Ticker
		//cb for ticker
		cbForGenTicker func(int32) error
		cbForBindObjTicker func(int32, ...interface{}) error
		sync.RWMutex
	}
)

//face info
type Worker struct {
	//basic
	workerMap map[int32]*SonWorker
	workers int32

	//cb func
	cbForQueueOpt func(interface{})(interface{}, error)
	cbForGenTickerOpt func(int32) error
	cbForBindObjTickerOpt func(int32,...interface{}) error
	sync.RWMutex
}

//construct
func NewWorker() *Worker {
	this := &Worker{
		workerMap: map[int32]*SonWorker{},
	}
	return this
}

//quit
func (f *Worker) Quit() {
	f.Lock()
	defer f.Unlock()
	for k, v := range f.workerMap {
		v.Quit()
		delete(f.workerMap, k)
	}
	atomic.StoreInt32(&f.workers, 0)
	runtime.GC()
}

//set cb for queue opt, STEP-1-1
func (f *Worker) SetCBForQueueOpt(cb func(interface{}) (interface{}, error)) {
	//check
	if cb == nil {
		return
	}
	f.cbForQueueOpt = cb

	//sync into running son workers
	f.Lock()
	defer f.Unlock()
	for _, v := range f.workerMap {
		if v.queue == nil {
			v.queue = NewQueue()
		}
		v.queue.SetCallback(cb)
	}
}

//set cb for gen ticker opt, STEP-1-2
func (f *Worker) SetCBForGenTickerOpt(cb func(int32) error) {
	//check
	if cb == nil {
		return
	}
	f.cbForGenTickerOpt = cb

	//sync into running son workers
	f.Lock()
	defer f.Unlock()
	for _, v := range f.workerMap {
		v.SetCBForGenTicker(cb)
	}
}

//set cb for bind obj ticker opt, STEP-1-3
func (f *Worker) SetCBForBindObjTickerOpt(cb func(int32,...interface{}) error) {
	//check
	if cb == nil {
		return
	}
	f.cbForBindObjTickerOpt = cb

	//sync into running son workers
	f.Lock()
	defer f.Unlock()
	for _, v := range f.workerMap {
		v.SetCBForBindObjTicker(cb)
	}
}

//create workers, STEP-2
//if tickerRates > 0, will create son worker ticker
func (f *Worker) CreateWorkers(
	num int,
	tickerRates ...float64) error {
	//check
	if num <= 0 {
		return errors.New("invalid parameter")
	}

	//init batch son workers with locker
	f.Lock()
	defer f.Unlock()
	for i := 0; i < num; i++ {
		//gen new worker id
		newWorkerId := atomic.AddInt32(&f.workers, 1)

		//init son worker
		sw := NewSonWorker(newWorkerId, tickerRates...)

		//set queue cb
		if f.cbForQueueOpt != nil {
			if sw.queue == nil {
				sw.queue = NewQueue()
			}
			sw.queue.SetCallback(f.cbForQueueOpt)
		}

		//set ticker cb
		if sw.ticker != nil {
			if f.cbForBindObjTickerOpt != nil {
				sw.SetCBForBindObjTicker(f.cbForBindObjTickerOpt)
			}else{
				sw.SetCBForGenTicker(f.cbForGenTickerOpt)
			}
		}

		//sync into run map
		f.workerMap[newWorkerId] = sw
	}
	return nil
}

//send data to one worker, STEP-3
//dataIds used for hash calculate value
func (f *Worker) SendData(
	data interface{},
	dataId int64,
	needResponses ...bool) (interface{}, error) {
	//check
	if data == nil {
		return nil, errors.New("invalid parameter")
	}
	if f.workers <= 0 {
		return nil, errors.New("no any workers")
	}

	//get son worker
	sonWorker, err := f.GetRandWorker(dataId)
	if err != nil || sonWorker == nil {
		return nil, err
	}

	//send data to queue
	resp, subErr := sonWorker.queue.SendData(data, needResponses...)
	return resp, subErr
}

//cast data to all workers
func (f *Worker) CastData(data interface{}) error {
	//check
	if data == nil {
		return errors.New("invalid parameter")
	}
	if f.workers <= 0 {
		return errors.New("no any workers")
	}

	//send data to all workers
	for _, v := range f.workerMap {
		v.queue.SendData(data)
	}
	return nil
}

//get workers
func (f *Worker) GetWorkers() int32 {
	return f.workers
}

//get son worker
func (f *Worker) GetRandWorker(
	dataIds ...int64) (*SonWorker, error) {
	var (
		dataId int64
		hashIdx int32
	)
	//check
	if dataIds != nil && len(dataIds) > 0 {
		dataId = dataIds[0]
	}

	//gen hashed worker id
	if dataId <= 0 {
		//hashed by rand
		now := time.Now().UnixNano()
		rand.Seed(now)
		hashIdx = int32(rand.Int63n(now) % int64(f.workers)) + 1
	}else{
		//hashed by data id
		hashIdx = int32(rand.Int63n(dataId) % int64(f.workers)) + 1
	}

	//get target son worker
	v, ok := f.workerMap[hashIdx]
	if ok && v != nil {
		return v, nil
	}
	return nil, errors.New("can't get son worker")
}

func (f *Worker) GetWorker(
	workerId int32) (*SonWorker, error) {
	//check
	if workerId <= 0 {
		return nil, errors.New("invalid parameter")
	}
	f.Lock()
	defer f.Unlock()
	v, ok := f.workerMap[workerId]
	if ok && v != nil {
		return v, nil
	}
	return nil, errors.New("no such worker")
}

////////////////////
//api for son worker
////////////////////

//construct
func NewSonWorker(id int32, tickerRates ...float64) *SonWorker {
	var (
		tickerRate float64
	)
	if tickerRates != nil && len(tickerRates) > 0 {
		tickerRate = tickerRates[0]
	}

	//self init
	this := &SonWorker{
		workerId: id,
		bindObjs: map[int64]interface{}{},
	}

	//check and start default ticker
	if tickerRate > 0 {
		this.ticker = NewTicker(tickerRate)
		this.ticker.SetCheckerCallback(this.interCBForGenTicker)
	}
	return this
}

//quit
func (f *SonWorker) Quit() {
	if f.queue != nil {
		f.queue.Quit()
	}
	if f.ticker != nil {
		f.ticker.Quit()
	}
}

//send data
func (f *SonWorker) SendData(data interface{}) (interface{}, error) {
	//check
	if data == nil {
		return nil, errors.New("invalid parameter")
	}
	if f.queue == nil {
		return nil, errors.New("inter queue not init")
	}

	//send data to queue
	resp, err := f.queue.SendData(data)
	return resp, err
}

//set cb for gen ticker
func (f *SonWorker) SetCBForGenTicker(cb func(int32) error) error {
	//check
	if cb == nil {
		return errors.New("invalid parameter")
	}
	if f.ticker == nil {
		return errors.New("inter ticker not init")
	}

	//sync cb
	f.cbForGenTicker = cb
	f.ticker.SetCheckerCallback(f.interCBForGenTicker)
	return nil
}

//set cb for bind obj ticker
func (f *SonWorker) SetCBForBindObjTicker(cb func(int32,...interface{}) error) error {
	//check
	if cb == nil {
		return errors.New("invalid parameter")
	}
	if f.ticker == nil {
		return errors.New("inter ticker not init")
	}

	//sync cb
	f.cbForBindObjTicker = cb
	f.ticker.SetCheckerCallback(f.interCBForBindObjTicker)
	return nil
}

//get bind obj
func (f *SonWorker) GetBindObj() map[int64]interface{} {
	//get with locker
	f.Lock()
	defer f.Unlock()
	return f.bindObjs
}

//remove bind obj
func (f *SonWorker) RemoveBindObj(objId int64) error {
	//check
	if objId <= 0 {
		return errors.New("invalid parameter")
	}

	//remove with locker
	f.Lock()
	defer f.Unlock()
	delete(f.bindObjs, objId)
	return nil
}

//update bind obj
func (f *SonWorker) UpdateBindObj(objId int64, obj interface{}) error {
	//check
	if objId <= 0 || obj == nil {
		return errors.New("invalid parameter")
	}

	//sync bind obj with locker
	f.Lock()
	defer f.Unlock()
	f.bindObjs[objId] = obj
	return nil
}

//////////////////
//private func
//////////////////

//inter cb opt for gen ticker
//gen and bind obj ticker, only run one
func (f *SonWorker) interCBForGenTicker() error {
	if f.cbForGenTicker == nil {
		return errors.New("inter cb for gen opt is nil")
	}
	err := f.cbForGenTicker(f.workerId)
	return err
}

//inter cb opt for bind obj ticker
func (f *SonWorker) interCBForBindObjTicker() error {
	if f.cbForBindObjTicker == nil {
		return errors.New("inter cb for bind obj opt is nil")
	}
	err := f.cbForBindObjTicker(f.workerId, f.bindObjs)
	return err
}