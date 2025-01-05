package util

import (
	"math/rand"
	"reflect"
	"time"
)

//face info
type Object struct {
}

//reset object instance
func (f *Object) RestObject(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}

//shuffle int slice
func (f *Object) ShuffleSlice(data []int) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(data), func(i, j int) { data[i], data[j] = data[j], data[i] })
}

//reverse general slice
func (f *Object) ReverseSlice(args ...interface{}) []interface{}{
	for i := 0; i < len(args)/2; i++ {
		j := len(args) - i - 1
		args[i], args[j] = args[j], args[i]
	}
	return args
}