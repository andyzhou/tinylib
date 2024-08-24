package util

import (
	"encoding/json"
	"errors"
	"log"
)

/*
 * json opt face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */
type BaseJson struct {
}

//construct
func NewBaseJson() *BaseJson {
	this := &BaseJson{}
	return this
}

////for fix redis!!!!
//func (j BaseJson) MarshalBinary() ([]byte, error) {
//	return json.Marshal(j)
//}
//func (j BaseJson) UnmarshalBinary(data []byte) error {
//	return json.Unmarshal(data, j)
//}

//decode map obj to json obj
func (j *BaseJson) DecodeMap2JsonObj(
	mapRec map[string]interface{},
	jsonObj interface{}) error {
	//check
	if mapRec == nil || jsonObj == nil {
		return errors.New("invalid parameter")
	}

	//convert to hash map obj
	jsonBytes, err := json.Marshal(mapRec)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonBytes, &jsonObj)
	return err
}

//encode json obj to map obj
func (j *BaseJson) EncodeJsonObj2Map(
	jsonObj interface{}) (map[string]interface{}, error) {
	//convert to hash map obj
	jsonBytes, err := json.Marshal(jsonObj)
	if err != nil {
		return nil, err
	}
	countMap := make(map[string]interface{})
	err = json.Unmarshal(jsonBytes, &countMap)
	return countMap, nil
}

//encode self
func (j *BaseJson) EncodeSelf() ([]byte, error) {
	//encode json
	resp, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//decode self
func (j *BaseJson) DecodeSelf(data []byte) error {
	err := json.Unmarshal(data, j)
	return err
}

//encode json data
func (j *BaseJson) Encode(i interface{}) ([]byte, error) {
	if i == nil {
		return nil, errors.New("invalid parameter")
	}
	//encode json
	resp, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//decode json data
func (j *BaseJson) Decode(data []byte, i interface{}) error {
	if len(data) <= 0 {
		return errors.New("json data is empty")
	}
	//try decode json data
	err := json.Unmarshal(data, i)
	if err != nil {
		//log.Println("BaseJson::Decode, decode failed, err:", err.Error())
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		//log.Println("BaseJson::Decode, track:", string(debug.Stack()))
		return err
	}
	return nil
}

//encode simple kv data
func (j *BaseJson) EncodeSimple(data map[string]interface{}) ([]byte, error) {
	if data == nil {
		return nil, errors.New("json data is empty")
	}
	//try encode json data
	byte, err := json.Marshal(data)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		return nil, err
	}
	return byte, nil
}

//decode simple kv data
func (j *BaseJson) DecodeSimple(data []byte, kv map[string]interface{}) error {
	if len(data) <= 0 {
		return errors.New("json data is empty")
	}
	//try decode json data
	err := json.Unmarshal(data, &kv)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Printf("sakura response: %q", data)
		return err
	}
	return nil
}

//encode to json string
func (j *BaseJson) Encode2Str(i interface{}) (string, error) {
	jsonByte, err := j.Encode(i)
	if err != nil {
		return "", err
	}
	return string(jsonByte), nil
}