package testing

import (
	"errors"
	"github.com/andyzhou/tinylib/queue"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var (
	worker *queue.Worker
	workerTickRate = float64(1/5)
	maxWorkers = 5
	workerId int32
)

//init
func init() {
	//init worker
	worker = queue.NewWorker()
	worker.SetCBForBindObjTickerOpt(cbForBindObjTicker)

	//create batch son workers
	for i := 0; i < maxWorkers; i++ {
		worker.CreateWorkers(i, workerTickRate)
	}
}

//cb for bind obj ticker
func cbForBindObjTicker(workerId int32, objs ...interface{}) error {
	//check
	if workerId <= 0 || objs == nil || len(objs) <= 0 {
		return errors.New("invalid parameter")
	}
	log.Printf("cbForBindObjTicker, workerId:%v, objs:%v\n", workerId, objs)
	return nil
}

//bind obj access
func bindObjAccess(val int64) (interface{}, error) {
	//get rand worker id
	workerId := rand.Intn(maxWorkers) + 1

	//get son worker
	sonWorker, err := worker.GetWorker(int32(workerId))
	if err != nil {
		return nil, err
	}

	//get obj
	obj := sonWorker.GetBindObj()

	//check or init obj
	if obj == nil {
		obj = make(map[int64]interface{})
	}
	objMap, ok := obj.(map[int64]interface{})
	if !ok || objMap == nil {
		return nil, errors.New("invalid obj type")
	}

	//update obj
	objMap[val] = val
	err = sonWorker.UpdateBindObj(objMap)
	return obj, err
}

//test bind obj
func TestBindObj(t *testing.T) {
	//obj access
	obj, err := bindObjAccess(time.Now().Unix())
	if err != nil {
		t.Errorf("test bind obj failed, err:%v\n", err.Error())
		return
	}
	t.Logf("test bind obj, obj:%v\n", obj)
}

//test bind obj ticker
func TestBindObjTicker(t *testing.T) {
	var (
		wg sync.WaitGroup
	)
	//obj access
	_, err := bindObjAccess(time.Now().Unix())
	if err != nil {
		t.Errorf("test bind obj failed, err:%v\n", err.Error())
		return
	}
	sf := func() {
		wg.Done()
	}
	wg.Add(1)
	time.AfterFunc(time.Second * 3, sf)
	wg.Wait()
	t.Logf("test bind obj ticker succeed\n")
}

//benchmark bind obj
func BenchmarkBindObj(b *testing.B) {
	succeed := 0
	failed := 0
	for i := 0; i < b.N; i++ {
		//create son worker
		worker.CreateWorkers(1, workerTickRate)

		//obj access
		_, err := bindObjAccess(int64(i))
		if err != nil {
			failed++
		}else{
			succeed++
		}
	}
	b.Logf("benchmark bind obj, succeed:%v, failed:%v\n", succeed, failed)
}