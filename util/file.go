package util

import (
	"errors"
	"io/ioutil"
	"os"
)

/*
 * gen file opt face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type File struct {
}

//check file stat and last modify time
func (f *File) GetFileModifyTime(
	filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	modifyTime := fileInfo.ModTime().Unix()
	return modifyTime, nil
}

//get file info
func (f *File) GetFileInfo(
	filePath string) (os.FileInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	return fileInfo, nil
}

//read byte file
func (f *File) ReadBinFile(
	filePath string,
	needRemoves ...bool) ([]byte, error) {
	//check
	if filePath == "" {
		return nil, errors.New("invalid file path")
	}
	//try read file
	byteData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	needRemove := false
	if needRemoves != nil && len(needRemoves) > 0 {
		needRemove = needRemoves[0]
		if needRemove {
			os.Remove(filePath)
		}
	}
	return byteData, nil
}

//check or create dir
func (f *File) CheckOrCreateDir(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		return err
	}
	bRet := os.IsExist(err)
	if bRet {
		return nil
	}
	err = os.Mkdir(dir, 0777)
	return err
}