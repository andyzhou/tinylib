package util

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

/*
 * util tools
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//inter macro define
const (
	DefaultAsciiSize = 2
)

//face info
type Util struct {
}

//gen md5 string
func (f *Util) GenMd5(orgString string) string {
	if len(orgString) <= 0 {
		return ""
	}
	m := md5.New()
	m.Write([]byte(orgString))
	return hex.EncodeToString(m.Sum(nil))
}

//check chan is closed or not
//true:closed, false:opening
func (f *Util) IsChanClosed(ch interface{}) (bool, error) {
	//check
	if reflect.TypeOf(ch).Kind() != reflect.Chan {
		return false, errors.New("input value not channel type")
	}

	// get interface value pointer, from cgo_export
	// typedef struct { void *t; void *v; } GoInterface;
	// then get channel real pointer
	cPtr := *(*uintptr)(unsafe.Pointer(
		unsafe.Pointer(uintptr(unsafe.Pointer(&ch)) + unsafe.Sizeof(uint(0))),
	))

	// this function will return true if chan.closed > 0
	// see hchan on https://github.com/golang/go/blob/master/src/runtime/chan.go
	// type hchan struct {
	// qcount   uint           // total data in the queue
	// dataqsiz uint           // size of the circular queue
	// buf      unsafe.Pointer // points to an array of dataqsiz elements
	// elemsize uint16
	// closed   uint32
	// **

	cPtr += unsafe.Sizeof(uint(0))*2
	cPtr += unsafe.Sizeof(unsafe.Pointer(uintptr(0)))
	cPtr += unsafe.Sizeof(uint16(0))
	return *(*uint32)(unsafe.Pointer(cPtr)) > 0, nil
}

//string to integer
func (f *Util) Str2Int(input string) int64 {
	out, _ := strconv.ParseInt(input, 10, 64)
	return out
}
func (f *Util) Int2Str(input int64) string {
	return fmt.Sprintf("%v", input)
}

//slice convert
func (f *Util) GenSlice2StrSlice(input ...interface{}) []string {
	if input == nil {
		return nil
	}
	result := make([]string, 0)
	for _, val := range input {
		if val == nil {
			continue
		}
		result = append(result, fmt.Sprintf("%v", val))
	}
	return result
}
func (f *Util) IntSlice2StrSlice(input ...int64) []string {
	if input == nil {
		return nil
	}
	result := make([]string, 0)
	for _, val := range input {
		if val <= 0 {
			continue
		}
		result = append(result, fmt.Sprintf("%v", val))
	}
	return result
}
func (f *Util) StrSlice2IntSlice(input ...string) []int64 {
	if input == nil {
		return nil
	}
	result := make([]int64, 0)
	for _, val := range input {
		if val == "" {
			continue
		}
		intVal, _ := strconv.ParseInt(val, 10, 64)
		result = append(result, intVal)
	}
	return result
}
func (f *Util) StrSlice2GenSlice(input ...string) []interface{} {
	if input == nil {
		return nil
	}
	result := make([]interface{}, 0)
	for _, val := range input {
		result = append(result, val)
	}
	return result
}
func (f *Util) IntSlice2GenSlice(input ...int64) []interface{} {
	if input == nil {
		return nil
	}
	result := make([]interface{}, 0)
	for _, v := range input {
		result = append(result, v)
	}
	return result
}

//get ascii value
func (f *Util) GetAsciiValue(
	input string,
	sizes ...int) (int, error) {
	var (
		size int
	)
	//check
	if input == "" {
		return 0, errors.New("invalid input parameter")
	}

	//detect assigned size
	if sizes != nil && len(sizes) > 0 {
		size = sizes[0]
	}
	if size <= 0 {
		size = DefaultAsciiSize
	}
	inputLen := len(input)
	if inputLen < size {
		size = inputLen
	}

	//loop check
	finalVal := 1
	for idx, val := range input {
		if idx >= size {
			break
		}
		factor := 10 << idx
		asciiVal := int(val)
		finalVal += asciiVal * factor
	}
	return finalVal, nil
}

//deep copy object
func (f *Util) DeepCopy(src, dist interface{}) (err error){
	buf := bytes.Buffer{}
	if err = gob.NewEncoder(&buf).Encode(src); err != nil {
		return
	}
	return gob.NewDecoder(&buf).Decode(dist)
}

//get rand number
func (f *Util) GetRandomVal(maxVal int) int {
	randSand := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSand)
	return r.Intn(maxVal)
}
func (f *Util) GetRealRandomVal(maxVal int) int {
	return int(rand.Float64() * 1000) % (maxVal)
}

//shuffle objs
func (f *Util) Shuffle(slice interface{}) {
	rand.Seed(time.Now().UnixNano())
	rv := reflect.ValueOf(slice)
	swap := reflect.Swapper(slice)
	length := rv.Len()
	for i := length - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		swap(i, j)
	}
}