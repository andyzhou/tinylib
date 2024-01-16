package crypt

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

/*
 * Simple encrypt algorithm
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

// simple encrypt info
type SimpleEncrypt struct {
	key string `base security key`
}

// construct
func NewSimpleEncrypt(securityKeys ...string) *SimpleEncrypt {
	securityKey := SimpleEncryptKeyDefault
	if securityKeys != nil && len(securityKeys) > 0 {
		securityKey = securityKeys[0]
	}
	this := &SimpleEncrypt{
		key: securityKey,
	}
	return this
}

// decrypt string
func (e *SimpleEncrypt) Decrypt(encStr string) (string, error) {
	var (
		encryptStr = bytes.NewBuffer(nil)
		tmpByte    byte
	)
	//check
	if encStr == "" || e.key == "" {
		return "", errors.New("invalid parameter")
	}
	//try decode
	decodeByte, err := base64.StdEncoding.DecodeString(encStr)
	if err != nil {
		return "", err
	}
	decodeStr := string(decodeByte)
	bindKey := e.bindKey(decodeStr)
	bindKeyLen := len(bindKey)
	for i := 0; i < bindKeyLen; i++ {
		if i >= (bindKeyLen - 1) {
			continue
		}
		tmpByte = byte(bindKey[i] ^ bindKey[i+1])
		encryptStr.WriteByte(tmpByte)
		i++
	}
	return encryptStr.String(), nil
}

// encrypt string
func (e *SimpleEncrypt) Encrypt(orgStr string) (string, error) {
	var (
		encryptStr = bytes.NewBuffer(nil)
		tmpByte    uint8
		j          int
	)

	//check
	if orgStr == "" || e.key == "" {
		return "", errors.New("invalid parameter")
	}
	mixStr := e.genMixString()
	orgStrLen := len(orgStr)
	for i := 0; i < orgStrLen; i++ {
		if j == SimpleEncryptKeySize {
			j = 0
		}
		tmpByte = byte(orgStr[i] ^ mixStr[j])
		encryptStr.WriteByte(byte(mixStr[j]))
		encryptStr.WriteByte(tmpByte)
		j++
	}
	bindKey := e.bindKey(encryptStr.String())
	return base64.StdEncoding.EncodeToString([]byte(bindKey)), nil
}

// set key
func (e *SimpleEncrypt) SetKey(key string) error {
	if key == "" {
		return errors.New("invalid parameter")
	}
	e.key = key
	return nil
}

///////////////
//private func
///////////////

// bind key
func (e *SimpleEncrypt) bindKey(orgStr string) string {
	var (
		tmpByte byte
		buffStr = bytes.NewBuffer(nil)
		i, j    int
	)
	entryKey := e.genMd5(e.key)
	orgStrLen := len(orgStr)
	for i = 0; i < orgStrLen; i++ {
		if j == SimpleEncryptKeySize {
			j = 0
		}
		tmpByte = byte(orgStr[i] ^ entryKey[j])
		buffStr.WriteByte(tmpByte)
		j++
	}
	return buffStr.String()
}

// generate mix string
func (e *SimpleEncrypt) genMixString() string {
	now := int(time.Now().Unix())
	randStr := fmt.Sprintf("%v", now)
	md5Str := e.genMd5(randStr)
	return md5Str
}

//gen md5 value
func (e *SimpleEncrypt) genMd5(input string) string {
	if len(input) <= 0 {
		return ""
	}
	m := md5.New()
	m.Write([]byte(input))
	return hex.EncodeToString(m.Sum(nil))
}
