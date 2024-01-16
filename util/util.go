package util

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"
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

//string to integer
func (f *Util) Str2Int(input string) int64 {
	out, _ := strconv.ParseInt(input, 10, 64)
	return out
}
func (f *Util) Int2Str(input int64) string {
	return fmt.Sprintf("%v", input)
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