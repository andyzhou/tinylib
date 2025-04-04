package queue

import (
	"errors"
	"fmt"
	"github.com/andyzhou/tinylib/util"
	"math/rand"
	"runtime"
	"strconv"
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
		queue    *Queue
		ticker   *Ticker
		//cb for ticker
		cbForGenTicker     func(int32) error
		cbForBindObjTicker func(int32, ...interface{}) error
		sync.RWMutex
	}
)

//face info
type Worker struct {
	//basic
	workerMap   map[int32]*SonWorker //workerId -> *SonWorker
	workerIdMap map[int64]int32      //dataId -> workerId, for bind obj
	workers     int32

	//cb func
	cbForQueueOpt         func(interface{}) (interface{}, error)
	cbForGenTickerOpt     func(int32) error
	cbForBindObjTickerOpt func(int32, ...interface{}) error

	sync.RWMutex
	util.Util
}

//construct
func NewWorker() *Worker {
	this := &Worker{
		workerMap: map[int32]*SonWorker{},
		workerIdMap: map[int64]int32{},
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
//if setup, will open queue
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
//if setup, will open ticker
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
//if setup, used for bind obj ticker opt
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
	num int, tickerRates ...float64) error {
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
//objIds used for hash calculate value
func (f *Worker) SendData(
		data interface{},
		objIds []int64,
		needResponses ...bool,
	) (map[int64]interface{}, error) {
	//check
	if data == nil || objIds == nil {
		return nil, errors.New("invalid parameter")
	}
	if f.workers <= 0 {
		return nil, errors.New("no any workers")
	}

	//loop process
	result := map[int64]interface{}{}
	for _, objId := range objIds {
		//get son worker
		sonWorker, err := f.GetTargetWorker(objId)
		if err != nil || sonWorker == nil {
			continue
		}
		//send data to queue
		resp, subErr := sonWorker.queue.SendData(data, needResponses...)
		if subErr != nil {
			continue
		}
		result[objId] = resp
	}
	return result, nil
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

//get all objs
func (f *Worker) GetAllBindObj(workerId int32) (map[int64]interface{}, error) {
	//get target worker by id
	sonWorker, err := f.GetWorker(workerId)
	if err != nil || sonWorker == nil {
		return nil, err
	}
	//get all objs
	objs := sonWorker.GetAllBindObjs()
	return objs, nil
}

//get one obj by id
func (f *Worker) GetBindObj(objId int64) (interface{}, error) {
	//check
	if objId <= 0 {
		return nil, errors.New("invalid parameter")
	}

	//get target worker
	sonWorker, err := f.GetTargetWorker(objId)
	if err != nil || sonWorker == nil {
		return nil, err
	}

	//get from son worker
	obj, subErr := sonWorker.GetBindObj(objId)
	return obj, subErr
}

//remove bind obj
func (f *Worker) RemoveBindObj(objId int64) error {
	//check
	if objId <= 0 {
		return errors.New("invalid parameter")
	}

	//get target worker
	sonWorker, err := f.GetTargetWorker(objId)
	if err != nil || sonWorker == nil {
		return err
	}

	//remove from son worker
	err = sonWorker.RemoveBindObj(objId)
	if err != nil {
		return err
	}

	//remove obj id from cached map
	f.Lock()
	defer f.Unlock()
	delete(f.workerIdMap, objId)
	return nil
}

//update bind obj
func (f *Worker) UpdateBindObj(objId int64, obj interface{}) error {
	//check
	if objId <= 0 || obj == nil {
		return errors.New("invalid parameter")
	}

	//get target worker
	sonWorker, err := f.GetTargetWorker(objId)
	if err != nil || sonWorker == nil {
		return err
	}

	//update into son worker
	err = sonWorker.UpdateBindObj(objId, obj)
	return err
}

//get son worker
//extParas -> dataId(int64), needBind(bool)
func (f *Worker) GetTargetWorker(extParas ...interface{}) (*SonWorker, error) {
	var (
		objId int64
		targetWorkerId int32
		needBind bool
	)
	//check
	if extParas != nil {
		extParaLen := len(extParas)
		switch extParaLen {
		case 1:
			{
				objId = f.Str2Int(fmt.Sprintf("%v", extParas[0]))
			}
		case 2:
			{
				objId = f.Str2Int(fmt.Sprintf("%v", extParas[0]))
				needBind, _ = strconv.ParseBool(fmt.Sprintf("%v", extParas[0]))
			}
		}
	}

	//gen hashed worker id
	f.Lock()
	defer f.Unlock()
	if objId <= 0 {
		//hashed by rand
		now := time.Now().UnixNano()
		rand.Seed(now)
		targetWorkerId = int32(rand.Int63n(now) % int64(f.workers)) + 1
	}else{
		//hashed by data id
		if needBind {
			//get from cached map
			v, ok := f.workerIdMap[objId]
			if !ok || v <= 0 {
				//hashed by data id
				targetWorkerId = int32(rand.Int63n(objId) % int64(f.workers)) + 1
				//sync into cache map
				f.workerIdMap[objId] = targetWorkerId
			}else{
				targetWorkerId = v
			}
		}else{
			//hashed by data id
			targetWorkerId = int32(rand.Int63n(objId) % int64(f.workers)) + 1
		}
	}

	//get target son worker
	v, ok := f.workerMap[targetWorkerId]
	if ok && v != nil {
		return v, nil
	}
	return nil, errors.New("can't get son worker")
}

func (f *Worker) GetWorker(workerId int32) (*SonWorker, error) {
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
	//gc opt
	runtime.GC()
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

//get all bind objs
func (f *SonWorker) GetAllBindObjs() map[int64]interface{} {
	//get with locker
	f.Lock()
	defer f.Unlock()
	return f.bindObjs
}

//get one bind obj
func (f *SonWorker) GetBindObj(objId int64) (interface{}, error) {
	//check
	if objId <= 0 {
		return nil, errors.New("invalid parameter")
	}
	//get with locker
	f.Lock()
	defer f.Unlock()
	v, ok := f.bindObjs[objId]
	if ok && v != nil {
		return v, nil
	}
	return nil, errors.New("no such obj by id")
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
	if len(f.bindObjs) <= 0 {
		//init new and gc memory
		f.bindObjs = map[int64]interface{}{}
		runtime.GC()
	}
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
func (f *SonWorker) interCBForGenTicker(inputs ...interface{}) error {
	if f.cbForGenTicker == nil {
		return errors.New("inter cb for gen opt is nil")
	}
	err := f.cbForGenTicker(f.workerId)
	return err
}

//inter cb opt for bind obj ticker
func (f *SonWorker) interCBForBindObjTicker(inputs ...interface{}) error {
	if f.cbForBindObjTicker == nil {
		return errors.New("inter cb for bind obj opt is nil")
	}
	err := f.cbForBindObjTicker(f.workerId, f.bindObjs)
	return err
}