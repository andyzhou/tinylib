package testing

import (
	"github.com/andyzhou/tinylib/queue"
	"math/rand"
	"testing"
	"time"
)

var (
	l *queue.List
)

func init() {
	l = queue.NewList()
	l.SetConsumer(cbForConsume, 1/10)
}

func cbForConsume(data interface{}) error {
	//log.Printf("cbForConsume, data:%v\n", data)
	return nil
}

func getLen() int {
	return l.Len()
}

func pushEle() {
	now := time.Now().UnixNano()
	rand.Seed(now)
	randVal := rand.Int63n(now)
	l.Push(randVal)
}

//testing
func TestList(t *testing.T) {
	listLen := l.Len()
	t.Logf("begin listLen:%v\n", listLen)

	pushEle()
	pushEle()
	listLen = l.Len()
	t.Logf("push ele listLen:%v\n", listLen)

	l.Pop()
	listLen = l.Len()
	t.Logf("pop listLen:%v\n", listLen)

	l.Quit()
	listLen = l.Len()
	t.Logf("quit listLen:%v\n", listLen)
}

//benchmark
func BenchmarkPush(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pushEle()
	}
	b.Logf("benchmark push N:%v\n", b.N)
	b.Logf("list len:%v\n", getLen())
}

