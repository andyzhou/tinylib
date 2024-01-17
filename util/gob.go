package util

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"sync"
)

/*
 * gob file opt face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//inter macro define
const (
	FilePerm     = 0755
)

//face info
type Gob struct {
	rootPath string
	sync.RWMutex
}

//construct
func NewGob() *Gob {
	this := &Gob{}
	return this
}

//load gob file
func (f *Gob) Load(fileName string, outVal interface{}) error {
	//check
	if fileName == "" || outVal == nil {
		return errors.New("invalid parameter")
	}

	//format gob file path
	filePath := fmt.Sprintf("%v/%v", f.rootPath, fileName)

	//try open file
	file, err := os.OpenFile(filePath, os.O_RDONLY, FilePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	//try decode gob file
	f.Lock()
	defer f.Unlock()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(outVal)
	return err
}

//store gob file
func (f *Gob) Store(fileName string, inputVal interface{}) error {
	//check
	if fileName == "" || inputVal == nil {
		return errors.New("invalid parameter")
	}

	//format gob file path
	filePath := fmt.Sprintf("%v/%v", f.rootPath, fileName)

	//open file
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, FilePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	//try encode gob file
	f.Lock()
	defer f.Unlock()
	enc := gob.NewEncoder(file)
	err = enc.Encode(inputVal)
	return err
}

//register assigned data type
//include map, struct, etc.
func (f *Gob) Register(val any) {
	gob.Register(val)
}
func (f *Gob) RegisterName(name string, val any) {
	gob.RegisterName(name, val)
}

//set root path
func (f *Gob) SetRootPath(path string) {
	f.rootPath = path
}