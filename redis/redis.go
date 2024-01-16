package redis

import (
	"errors"
	"github.com/go-redis/redis/v8"
	"sync"
)

/*
 * redis base face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Redis struct {
	connMap sync.Map //dbTag -> *Connection
	pubSub *PubSub
	sync.RWMutex
}

//construct
func NewRedis() *Redis {
	this := &Redis{
		connMap: sync.Map{},
		pubSub: NewPubSub(),
	}
	return this
}

//quit
func (f *Redis) Quit() {
	sf := func(k, v interface{}) bool {
		conn, ok := v.(*Connection)
		if ok && conn != nil {
			conn.Disconnect()
		}
		f.connMap.Delete(k)
		return true
	}
	f.connMap.Range(sf)
	f.pubSub.Close()
}

//get pub sub instance
func (f *Redis) GetPubSub() *PubSub {
	return f.pubSub
}

//access connect
func (f *Redis) C(dbName string) *Connection {
	if dbName == "" {
		return nil
	}
	v, ok := f.connMap.Load(dbName)
	if ok && v != nil {
		return v.(*Connection)
	}
	return nil
}

//create new connect
func (f *Redis) CreateConn(cfg *Config) (*Connection, error) {
	//check
	if cfg == nil {
		return nil, errors.New("invalid redis db config")
	}
	//check and release old
	v, ok := f.connMap.Load(cfg.DBTag)
	if ok && v != nil {
		v.(*Connection).Disconnect()
	}
	//check config
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = DefaultPoolSize
	}
	//init new connect
	connect := NewConnection()
	connect.config = cfg
	connect.client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DBNum,
		PoolSize: cfg.PoolSize,
	})
	//try connect
	err := connect.Connect()
	if err != nil {
		return nil, err
	}
	//sync into run env
	f.connMap.Store(cfg.DBTag, connect)
	//f.Lock()
	//defer f.Unlock()
	//f.connMap[cfg.DBTag] = connect
	return connect, nil
}

//gen new config
func (f *Redis) GenNewConfig() *Config {
	return &Config{}
}