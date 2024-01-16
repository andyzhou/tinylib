package redis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"log"
	"sync"
	"time"
)

/*
 * redis pub sub face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//inter data
type (
	PubSubCallback func(msg *redis.Message) error
)

//face info
type PubSub struct {
	conn *Connection //reference
	chanMap map[string]chan struct{} //channel -> chan struct{}
	timeout time.Duration
	sync.RWMutex
}

//construct
func NewPubSub() *PubSub {
	this := &PubSub{
		timeout: DefaultTimeOut,
		chanMap: map[string]chan struct{}{},
	}
	return this
}

//close sub process
func (f *PubSub) Close() {
	if f.chanMap == nil || len(f.chanMap) <= 0 {
		return
	}
	f.Lock()
	defer f.Unlock()
	for _, v := range f.chanMap {
		close(v)
	}
	f.chanMap = map[string]chan struct{}{}
}

//close channel
func (f *PubSub) CloseChannel(channelName string) error {
	if channelName == "" {
		return errors.New("invalid parameter")
	}
	f.Lock()
	defer f.Unlock()
	v, ok := f.chanMap[channelName]
	if !ok || v == nil {
		return errors.New("no such channel")
	}
	//close chan
	close(v)
	delete(f.chanMap, channelName)
	return nil
}

//publish message
func (f *PubSub) Publish(channelName string, message interface{}) error {
	//check
	if channelName == "" {
		return errors.New("invalid parameter")
	}
	if f.conn == nil {
		return errors.New("inter conn not init")
	}
	//key opt
	ctx, cancel := f.CreateContext()
	defer cancel()
	c := f.conn.GetConnect()
	_, err := c.Publish(ctx, channelName, message).Result()
	return err
}

//subscript channel
func (f *PubSub) Subscript(channelName string, cb PubSubCallback) error {
	var (
		m any = nil
	)
	//check
	if channelName == "" || cb == nil {
		return errors.New("invalid parameter")
	}
	if f.conn == nil {
		return errors.New("inter conn not init")
	}
	f.Lock()
	defer f.Unlock()
	_, ok := f.chanMap[channelName]
	if ok {
		return errors.New("channel had subscript")
	}
	closeChan := make(chan struct{}, 1)
	f.chanMap[channelName] = closeChan

	//run sub process
	sf := func(channelName string, closeChan chan struct{}, cb PubSubCallback) {
		defer func() {
			if err := recover(); err != m {
				log.Printf("PubSub:Subscript channel %v panic, err %v", channelName, err)
			}
			f.Lock()
			defer f.Unlock()
			close(closeChan)
			delete(f.chanMap, channelName)
		}()

		//key opt
		ctx, cancel := f.CreateContext()
		defer cancel()
		c := f.conn.GetClient()
		ps := c.Subscribe(ctx, channelName)
		dataChan := ps.Channel()

		//loop
		for {
			select {
			case data, ok := <- dataChan:
				if ok && cb != nil{
					cb(data)
				}
			case <- closeChan:
				return
			}
		}
	}
	go sf(channelName, closeChan, cb)
	return nil
}

//set base redis connect
func (f *PubSub) SetConn(conn *Connection) error {
	//check
	if conn == nil {
		return errors.New("invalid parameter")
	}
	f.conn = conn
	return nil
}

//create context
func (f *PubSub) CreateContext() (context.Context, context.CancelFunc){
	return context.WithTimeout(context.Background(), f.timeout*time.Second)
}
