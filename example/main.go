package main

import (
	"github.com/andyzhou/tinylib/queue"
	"github.com/andyzhou/tinylib/util"
	"log"
	"sync"
	"time"
)

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
func testQueue()  {
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

func main() {
	var (
		wg sync.WaitGroup
	)
	wg.Add(1)

	//test code
	//testChanIsClosed()
	testQueue()

	wg.Wait()
}
