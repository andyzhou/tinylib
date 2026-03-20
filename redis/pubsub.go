package redis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	genRedis "github.com/go-redis/redis/v8"
)

/*
 * redis pub sub face
 * 支持同一 channel 多回调
 * @author AndyZhou
 */
type PubSubCallback func(msg *genRedis.Message) error

type PubSubCallbackWrapper struct {
	Id string
	CB PubSubCallback
}

type channelInfo struct {
	callbacks []PubSubCallbackWrapper
	stopChan  chan struct{}
}

type PubSub struct {
	conn    *genRedis.Client
	chanMap map[string]*channelInfo
	mu      sync.RWMutex
}

func NewPubSub() *PubSub {
	return &PubSub{
		chanMap: make(map[string]*channelInfo),
	}
}

// 设置 Redis 连接
func (ps *PubSub) SetConn(conn *genRedis.Client) error {
	ps.conn = conn
	return nil
}

// Subscribe 订阅 channel，可以多次订阅同一 channel，回调异步执行
func (ps *PubSub) Subscribe(channel string, cb PubSubCallbackWrapper) error {
	if cb.Id == "" || channel == "" {
		return nil
	}
	if ps.conn == nil {
		return fmt.Errorf("redis connection not set")
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()
	info, ok := ps.chanMap[channel]
	if ok {
		// channel 已存在，追加回调
		info.callbacks = append(info.callbacks, cb)
		return nil
	}

	// 新建 channelInfo
	info = &channelInfo{
		callbacks: []PubSubCallbackWrapper{cb},
		stopChan:  make(chan struct{}),
	}
	ps.chanMap[channel] = info

	// 启动订阅 goroutine，传递 channel 和 info 的副本
	go ps.subscribeRoutine(channel, info)

	//go func(info channelInfo) {
	//	pubsub := ps.conn.Subscribe(context.Background(), channel)
	//	defer pubsub.Close()
	//
	//	//wail subscribe succeed
	//	if _, err := pubsub.Receive(context.Background()); err != nil {
	//		log.Printf("subscribe %v failed:", channel, err)
	//		return
	//	}
	//
	//	ch := pubsub.Channel()
	//	for {
	//		select {
	//		case msg, sok := <-ch:
	//			if !sok {
	//				return
	//			}
	//			ps.mu.RLock()
	//			for _, scb := range info.callbacks {
	//				go scb.CB(msg)
	//			}
	//			ps.mu.RUnlock()
	//		case <-info.stopChan:
	//			return
	//		}
	//	}
	//}(*info)

	return nil
}

func (ps *PubSub) subscribeRoutine(channel string, info *channelInfo) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubsub := ps.conn.Subscribe(ctx, channel)
	defer pubsub.Close()

	//wait subscribe succeed
	if _, err := pubsub.Receive(ctx); err != nil {
		log.Printf("subscribe %v failed: %v", channel, err)
		//cleanup
		ps.cleanupChannel(channel)
		return
	}

	ch := pubsub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			ps.handleMessage(info, msg)
		case <-info.stopChan:
			return
		}
	}
}

func (ps *PubSub) cleanupChannel(channel string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if info, ok := ps.chanMap[channel]; ok {
		close(info.stopChan)
		delete(ps.chanMap, channel)
	}
}

func (ps *PubSub) handleMessage(info *channelInfo, msg *genRedis.Message) {
	ps.mu.RLock()
	callbacks := make([]PubSubCallbackWrapper, len(info.callbacks))
	copy(callbacks, info.callbacks)
	ps.mu.RUnlock()

	for _, cb := range callbacks {
		go func(cb PubSubCallbackWrapper) {
			if err := cb.CB(msg); err != nil {
				log.Printf("callback %s error: %v", cb.Id, err)
			}
		}(cb)
	}
}

// Unsubscribe 支持取消单个回调或整个 channel
func (ps *PubSub) Unsubscribe(channel string, cb PubSubCallbackWrapper) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	info, ok := ps.chanMap[channel]
	if !ok {
		return
	}

	if &cb != nil {
		// 删除指定回调
		newCallbacks := []PubSubCallbackWrapper{}
		for _, f := range info.callbacks {
			if f.Id != cb.Id {
				newCallbacks = append(newCallbacks, f)
			}
		}
		info.callbacks = newCallbacks
	}

	// 如果没有回调了，关闭 channel
	if len(info.callbacks) == 0 {
		close(info.stopChan)
		delete(ps.chanMap, channel)
	}
}

// Close 关闭所有订阅
func (ps *PubSub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, info := range ps.chanMap {
		close(info.stopChan)
	}
	ps.chanMap = make(map[string]*channelInfo)
}

// Publish 使用独立连接发送消息
func (ps *PubSub) Publish(channel string, message interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ps.conn.Publish(ctx, channel, message).Err()
}