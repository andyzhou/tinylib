package main

import (
	"errors"
	"fmt"
	"github.com/andyzhou/tinylib/algorithm"
	"github.com/andyzhou/tinylib/queue"
	"github.com/andyzhou/tinylib/util"
	"github.com/andyzhou/tinylib/web"
	"log"
	"math/rand"
	"os"
	"runtime"
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
	objMap := sonWorker.GetAllBindObjs()

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
	var (
		wg sync.WaitGroup
	)
	//init queue
	q := queue.NewQueue()

	//set callback
	cbForQuit := func() {
		fmt.Printf("cbForQuit\n")
	}
	cbForOpt := func(data interface{}) (interface{}, error){
		//log.Printf("cbForOpt, data:%v\n", data)
		return nil, nil
	}
	q.SetCallback(cbForOpt)
	q.SetQuitCallback(cbForQuit)

	//add group
	wg.Add(1)

	//loop send func
	sendFunc := func() {
		sendRate := 0.01 //xxx seconds
		for {
			if q == nil {
				break
			}else{
				q.SendData(fmt.Sprintf("test-%v", time.Now().UnixNano()))
				time.Sleep(time.Duration(sendRate * float64(time.Second)))
			}
		}
	}
	go sendFunc()

	//delay quit
	delayQuit := func() {
		q.Quit()
		runtime.GC()
		wg.Done()
	}
	time.AfterFunc(time.Second * 300, delayQuit)

	//group wait
	wg.Wait()
}

//test list
func cbForListConsumer(data interface{}) error {
	return nil
}
func testList() {
	var (
		wg sync.WaitGroup
	)
	//init list
	l := queue.NewList()
	l.SetConsumer(cbForListConsumer, 0.001)

	wg.Add(1)

	//quit func
	qf := func() {
		l.Quit()
		wg.Done()
	}

	//test fill data
	sf := func() {
		sendRate := 0.001 //xx seconds
		for {
			if l.Closed() {
				break
			}
			l.Push(time.Now().UnixNano())
			time.Sleep(time.Duration(sendRate * float64(time.Second)))
		}
	}
	go sf()
	time.AfterFunc(time.Second * 60, qf)
	wg.Wait()
	fmt.Println("list test finish")
}

//test tick, pass
func testTick() {
	//init tick
	duration := 0.001 //N second
	t := queue.NewTicker(duration)

	//set callback
	cbForQuit := func() {
		fmt.Printf("cbForQuit\n")
	}
	cbForCheckOpt := func() error {
		//fmt.Printf("cbForCheckOpt, now:%v\n", time.Now().UnixNano())
		return nil
	}

	//set callback
	t.SetCheckerCallback(cbForCheckOpt)
	t.SetQuitCallback(cbForQuit)

	t.UpdateDuration(0.03)
	durationVal := t.GetDuration()
	fmt.Printf("durationVal:%v\n", durationVal)

	//delay opt
	delayOpt := func() {
		t.Quit()
	}
	time.AfterFunc(time.Second * 5, delayOpt)
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

//test page
func testPage() {
	//init obj
	page := web.NewPage()

	//setup config
	pageCfg := &web.PageConfig{
		TplPath: ".",
		StaticPath: ".",
	}
	page.SetConfig(pageCfg)

	//parse main tpl
	tpl, err := page.ParseTpl("test.tpl")
	if err != nil {
		log.Printf("err:%v\n", err.Error())
		os.Exit(1)
	}

	//get tpl page
	content, subErr := page.GetTplContent(tpl, nil)
	log.Printf("content:%v, err:%v\n", content, subErr)

	//gen html page
	pagePath := "test.html"
	err = page.GenHtmlPage(tpl, pagePath)
	log.Printf("err:%v\n", err)
}

//test c-hash
func testCHash()  {
	hashRing := algorithm.NewHashRingDefault()

	//add nodes
	hashRing.Add("0", "1")

	////get ring len
	//ringLen := hashRing.GetRingLen()
	//log.Printf("ringLen:%v\n", ringLen)
	//
	////get all rings
	//allRings := hashRing.GetRings()
	//log.Printf("allRings:%v\n", allRings)
	//
	////get ring by idx
	//targetRing := hashRing.GetByIdx(0)
	//log.Printf("targetRing:%v\n", targetRing)
	//
	////get idx by ring
	//ringIdx := hashRing.GetIdxByRing("0")
	//log.Printf("ringIdx:%v\n", ringIdx)
	//
	////get next ring
	//nextRing := hashRing.GetNextRing("4")
	//log.Printf("nextRing:%v\n", nextRing)
	//
	////get first ring
	//firstRing := hashRing.GetFirstRing()
	//log.Printf("firstRing:%v\n", firstRing)


	log.Printf("\n====1====\n")
	for i := 0; i < 30; i++ {
		tr := hashRing.Get(fmt.Sprintf("%v", i))
		log.Printf("i:%v, tr:%v\n", i, tr)
	}

	//log.Printf("===2===\n")
	//hashRing.Add("3")
	//for i := 0; i < 10; i++ {
	//	tr := hashRing.Get(fmt.Sprintf("%v", i))
	//	log.Printf("i:%v, tr:%v\n", i, tr)
	//}
}

//test phash
func testPHash() {
	//init hash
	hash := algorithm.NewConsistentHash()

	//add node
	hash.Add("0")
	hash.Add("1")

	log.Printf("\n====1====\n")
	for i := 0; i < 10; i++ {
		node, _ := hash.Get(i)
		log.Printf("i:%v, node:%v\n", i, node)
	}

	hash.Add("2")
	log.Printf("\n====2====\n")
	for i := 0; i < 10; i++ {
		node, _ := hash.Get(i)
		log.Printf("i:%v, node:%v\n", i, node)
	}
}

//test ring
func testRing() {
	ring := algorithm.NewRing()
	ring.AddNodes("1:7700", "2:7701", "3:7703", "4:7704")
	log.Printf("\n====1====\n")
	for i := 0; i < 30; i++ {
		v, _ := ring.GetNode(fmt.Sprintf("%v", i))
		log.Printf("i:%v, v:%v\n", i, v)
	}

	//ring.AddNodeWithWeight("3:7703", 1)
	//log.Printf("====2====\n")
	//for i := 0; i < 50; i++ {
	//	v, _ := ring.GetNode(fmt.Sprintf("%v", i))
	//	log.Printf("i:%v, v:%v\n", i, v)
	//}
}

//test consistent
func testConsistent() {
	// Create a new consistent instance
	cfg := algorithm.Config{
		PartitionCount:    7,
		ReplicationFactor: 20,
		Load:              1.25,
		Hasher:            algorithm.MyHasher{},
	}

	//init object
	c := algorithm.NewConsistent(nil, cfg)

	// Add some members to the consistent hash table.
	// Add function calculates average load and distributes partitions over members
	node1 := algorithm.NewNode("1")
	c.Add(node1)

	node2 := algorithm.NewNode("3")
	c.Add(node2)

	log.Printf("====1====\n")
	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("%v", i))
		owner := c.LocateKey(key)
		log.Printf("i:%v, owner:%v\n", i, owner.String())
	}

	node3 := algorithm.NewNode("2")
	c.Add(node3)

	//node4 := myMember("3")
	//c.Add(node4)

	log.Printf("====2====\n")
	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("%v", i))
		owner := c.LocateKey(key)
		log.Printf("i:%v, owner:%v\n", i, owner.String())
	}

	members := c.GetMembers()
	log.Printf("members:%v\n", members)
	for _, member := range members {
		log.Printf("member:%v\n", member.String())
	}

	average := c.AverageLoad()
	log.Printf("average:%v\n", average)

	disMap := c.LoadDistribution()
	log.Printf("disMap:%v\n", disMap)
}

//test x hash, stable!!
func testXHash() {
	//obj init
	consistentHash := algorithm.NewXConsistent()

	//add node with virual nodes count
	consistentHash.Add("0")
	consistentHash.Add("1")
	consistentHash.Add("2")

	//get all nodes
	nodes := consistentHash.GetOrgNodes()
	log.Printf("nodes:%v\n", nodes)

	//get target nodes
	log.Printf("====1====\n")
	for i := 1; i <= 10; i++ {
		node := consistentHash.GetNode(fmt.Sprintf("%v", i))
		log.Printf("i:%v, node:%v\n", i, node)
	}

	consistentHash.Add("3")
	log.Printf("====2====\n")
	for i := 11; i <= 50; i++ {
		node := consistentHash.GetNode(fmt.Sprintf("%v", i))
		log.Printf("i:%v, node:%v\n", i, node)
	}
}

//test date time
func testDateTime() {
	//t := util.Time{}
	now := time.Now().Unix()
	t := time.Unix(now, 0)
	//dayFormat := t.TimeStampToDayStr(now, 2)
	dd := t.Format(time.ANSIC)
	log.Printf("dd:%v\n", dd)
}

func main() {
	var (
		wg sync.WaitGroup
	)
	wg.Add(1)

	//test code
	//testChanIsClosed()
	//testQueue()
	//testList()
	testTick()
	//testWorker()
	//testPage()
	//testCHash()
	//testPHash()
	//testConsistent()
	//testRing()
	//testXHash()
	//testDateTime()
	wg.Wait()
}
