package util

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

/**
 * simple dynamic call instance method
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 *
 * for example:
 	face := base.NewFaceMap()
 	players := player.NewPlayers()
	face.Bind("players", players)
	result, err := face.Call("players", "Test", "this is string test")
	fmt.Println(result[0].String()) //print string type result
*/

//face info
type FaceMap struct {
	faceMap sync.Map //name -> interface{}
}

//construct
func NewFaceMap() *FaceMap {
	this := &FaceMap{
		faceMap: sync.Map{},
	}
	return this
}

//cleanup
func (f *FaceMap) CleanUp() {
	f.faceMap = sync.Map{}
}

//get face instance
func (f *FaceMap) GetFace(
	name string) interface{} {
	face, ok := f.faceMap.Load(name)
	if !ok && face != nil {
		return nil
	}
	return face
}

//bind face with name
func (f *FaceMap) Bind(
	name string,
	face interface{}) error {
	//check
	if name == "" || face == nil {
		return errors.New("invalid parameter")
	}
	//check is exists or not
	v := f.GetFace(name)
	if v != nil {
		return nil
	}
	f.faceMap.Store(name, reflect.ValueOf(face))
	return nil
}

//unbind face
func (f *FaceMap) UnBind(name string) error {
	//check
	if name == "" {
		return errors.New("invalid parameter")
	}
	f.faceMap.Delete(name)
	return nil
}

//call method on all faces
func (f *FaceMap) Cast(
	method string,
	params ...interface{}) error {
	//check
	if method == "" {
		return errors.New("invalid parameter")
	}

	//init parameters
	inParam := make([]reflect.Value, 0)
	if params != nil {
		for _, para := range params {
			inParam = append(inParam, reflect.ValueOf(para))
		}
	}
	//call method on each face
	subFunc := func(key interface{}, face interface{}) bool {
		faceObj, ok := face.(reflect.Value)
		if ok && &faceObj != nil {
			faceObj.MethodByName(method).Call(inParam)
		}
		return true
	}
	f.faceMap.Range(subFunc)
	return nil
}

//dynamic call method with parameters support
func (f *FaceMap) Call(
		name, method string,
		params ...interface{},
	) ([]reflect.Value, error) {
	//check
	if name == "" || method == "" {
		return nil, errors.New("invalid parameter")
	}

	//check instance
	face, isOk := f.faceMap.Load(name)
	if !isOk && &face == nil {
		tips := fmt.Sprintf("No face instance for name %s", name)
		return nil, errors.New(tips)
	}

	subFace, ok := face.(reflect.Value)
	if !ok || &subFace == nil {
		tips := fmt.Sprintf("Invalid face instance for name %s", name)
		return nil, errors.New(tips)
	}

	//init parameters
	inParam := make([]reflect.Value, 0)
	//f.params = len(params)
	for _, para := range params {
		inParam = append(inParam, reflect.ValueOf(para))
	}

	//dynamic call method with parameter
	callResult := subFace.MethodByName(method).Call(inParam)
	return callResult, nil
}