package mysql

import (
	"errors"
	"sync"
)

//face info
type Mysql struct {
	connectMap map[string]*Connect //dbTag -> *Connect
	JsonData
	sync.RWMutex
}

//construct
func NewMysql() *Mysql {
	this := &Mysql{
		connectMap: map[string]*Connect{},
	}
	return this
}

//quit
func (f *Mysql) Quit() {
	f.Lock()
	defer f.Unlock()
	for k, v := range f.connectMap {
		v.Quit()
		delete(f.connectMap, k)
	}
}

//get connect
func (f *Mysql) GetConnect(tag string) *Connect {
	f.Lock()
	defer f.Unlock()
	v, ok := f.connectMap[tag]
	if ok && v != nil {
		return v
	}
	return nil
}

//create connect
func (f *Mysql) CreateConnect(tag string, conf *Config) (*Connect, error) {
	//check
	if tag == "" || conf == nil {
		return nil, errors.New("invalid parameter")
	}

	//init new connect
	conn := NewConnect(conf)
	err := conn.Ping()
	if err != nil {
		return nil, err
	}

	//sync into map
	f.Lock()
	defer f.Unlock()
	f.connectMap[tag] = conn
	return conn, nil
}

//gen new config
func (f *Mysql)  GenNewConfig() *Config {
	return &Config{}
}