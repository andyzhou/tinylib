package main

import (
	"errors"
	"github.com/andyzhou/tinylib/queue"
	"github.com/andyzhou/tinylib/util"
	"log"
	"math/rand"
	"sync"
	"time"
)

var (
	worker *queue.Worker
	workerTickRate = float64(1)/float64(10)
	maxWorkers = 5
)

//init
func init() {
	//init worker
	worker = queue.NewWorker()
	worker.SetCBForBindObjTickerOpt(cbForBindObjTicker)
	//worker.SetCBForGenTickerOpt(cbForGenTicker)
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

//cb for gen ticker
func cbForGenTicker(workerId int32) error {
	log.Printf("cbForGenTicker, workerId:%v\n", workerId)
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
	objMap := sonWorker.GetBindObj()

	//check or init obj
	if objMap == nil {
		objMap = make(map[int64]interface{})
	}

	//update obj
	objMap[val] = val
	err = sonWorker.UpdateBindObj(val, val)
	return objMap, err
}

//test chan
func testChanIsClosed() {
	ch := make(chan bool, 1)
	//close(ch)
	sf := func() {
		u := util.Util{}
		isClosed, err := u.IsChanClosed(ch)
		log.Printf("isClosed:%v, err:%v\n", isClosed, err)
	}
	time.AfterFunc(time.Second * 3, sf)
}

//test queue
func testQueue() {
	//init queue
	q := queue.NewQueue()

	//set callback
	cbForQuit := func() {
		log.Printf("cbForQuit\n")
	}
	cbForOpt := func(data interface{}) (interface{}, error){
		log.Printf("cbForOpt, data:%v\n", data)
		return nil, nil
	}

	q.SetCallback(cbForOpt)
	q.SetQuitCallback(cbForQuit)

	//delay opt
	delayOpt := func() {
		q.SendData("test")
		q.Quit()
	}
	time.AfterFunc(time.Second * 2, delayOpt)
}

//test tick
func testTick() {
	//init tick
	t := queue.NewTicker()

	//set callback
	cbForQuit := func() {
		log.Printf("cbForQuit\n")
	}
	cbForCheckOpt := func() error {
		log.Printf("cbForCheckOpt\n")
		return nil
	}

	t.SetCheckerCallback(cbForCheckOpt)
	t.SetQuitCallback(cbForQuit)

	//delay opt
	delayOpt := func() {
		t.Quit()
	}
	time.AfterFunc(time.Second * 2, delayOpt)
}

//test worker
func testWorker() {
	var (
		wg sync.WaitGroup
	)
	//create son worker
	worker.CreateWorkers(maxWorkers, workerTickRate)

	//fill batch son worker
	for i := 0; i < 10; i++ {
		//obj access
		obj, err := bindObjAccess(time.Now().Unix())
		if err != nil {
			log.Printf("test worker, err:%v\n", err.Error())
			return
		}
		log.Printf("test bind obj ticker, obj:%v\n", obj)
	}

	sf := func() {
		wg.Done()
	}
	wg.Add(1)
	time.AfterFunc(time.Second * 30, sf)
	wg.Wait()
	log.Printf("test bind obj ticker succeed\n")
}

func main() {
	var (
		//wg sync.WaitGroup
	)
	//wg.Add(1)

	//test code
	//testChanIsClosed()
	//testQueue()
	//testTick()
	testWorker()

	//wg.Wait()
}
