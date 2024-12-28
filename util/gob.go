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
	FilePerm = 0755
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
	fileInfo, err := os.Stat(filePath)
	if err != nil || fileInfo == nil {
		return err
	}

	//opt with locker
	f.Lock()
	defer f.Unlock()

	//try open file
	file, subErr := os.OpenFile(filePath, os.O_RDONLY, FilePerm)
	if subErr != nil {
		return subErr
	}
	defer file.Close()

	//try decode gob file
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

	//opt with locker
	f.Lock()
	defer f.Unlock()

	//format gob file path
	filePath := fmt.Sprintf("%v/%v", f.rootPath, fileName)

	//open file
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, FilePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	//try encode gob file
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