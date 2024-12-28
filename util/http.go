package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

/*
 * http client face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//http request method
const (
	HttpReqGet = iota
	HttpReqPost
)

//inter macro define
const (
	HttpClientTimeOut  = 5 //xx seconds
	HttpReqChanSize    = 1024
	HttpReqRecChanSize = 1
	HttpReqPrefix      = "http://"
	HttpsReqPrefix     = "https://"
)

//http file para
type HttpFilePara struct {
	FilePath string
	FilePara string
}

type HttpResp struct {
	Data []byte
	Err  error
}

//http request face info
type HttpReq struct {
	Kind         int //GET or POST
	Url          string
	Headers      map[string]string
	Params       map[string]interface{}
	FilePara     HttpFilePara
	Body         []byte
	ReceiverChan chan HttpResp //http request receiver chan
	IsAsync      bool
}

//http face
type HttpClient struct {
	client    *http.Client `http client instance`
	reqChan   chan HttpReq `request lazy chan`
	closeChan chan bool
	sync.RWMutex
}

//construct
func NewHttpClient() *HttpClient {
	return NewHttpClientWithChanSize(0)
}

func NewHttpClientWithChanSize(reqChanSize int) *HttpClient {
	//check or init request chan size
	realReqChanSize := reqChanSize
	if realReqChanSize <= 0 {
		realReqChanSize = HttpReqChanSize
	}

	//self init
	this := &HttpClient{
		reqChan:make(chan HttpReq, realReqChanSize),
		closeChan:make(chan bool, 1),
	}

	//inter init
	this.interInit()

	//spawn main process
	go this.runMainProcess()

	return this
}

func NewHttpReq() *HttpReq {
	this := &HttpReq{
		Headers:make(map[string]string),
		Params:make(map[string]interface{}),
		Body:make([]byte, 0),
		ReceiverChan:make(chan HttpResp, HttpReqRecChanSize),
	}
	return this
}

//////////
//api
//////////

//queue quit
func (q *HttpClient) Quit() {
	if q.closeChan != nil {
		q.closeChan <- true
	}
}

//send request
func (q *HttpClient) SendReq(
	req *HttpReq) (*HttpResp, error) {
	var (
		m any = nil
	)
	//check
	if req == nil {
		return nil, errors.New("invalid parameter")
	}

	//init return result
	resp := &HttpResp{
		Data: []byte{},
	}

	//try catch panic
	defer func() {
		if err := recover(); err != m {
			log.Println("httpClient::SendReq, panic happened, err:", err)
			return
		}
	}()

	//send to chan
	q.reqChan <- *req

	//if request is async, just return
	if req.IsAsync {
		return resp, nil
	}

	//sync mode, need wait for response
	*resp, _ = <- req.ReceiverChan
	return resp, nil
}

//gen new request
func (q *HttpClient) GenRequest() *HttpReq {
	return &HttpReq{
		Headers: map[string]string{},
		Params: map[string]interface{}{},
		Body: []byte{},
		ReceiverChan: make(chan HttpResp, HttpReqRecChanSize),
	}
}

//////////////////
//private func
//////////////////

//run main process
func (q *HttpClient) runMainProcess() {
	var (
		req HttpReq
		isOk bool
		m any = nil
	)

	//defer
	defer func() {
		if err := recover(); err != m {
			log.Printf("http")
		}
		//close chan
		close(q.reqChan)
		close(q.closeChan)
	}()

	//loop
	for {
		select {
		case req, isOk = <- q.reqChan:
			//process single request
			if isOk && &req != nil {
				data, err := q.sendHttpReq(&req)
				resp := &HttpResp{
					Data: data,
					Err: err,
				}
				req.ReceiverChan <- *resp
			}
		case <- q.closeChan:
			return
		}
	}
}

//upload file request
func (q *HttpClient) fileUploadReq(reqObj *HttpReq) (*http.Request, error) {
	//try open file
	file, err := os.Open(reqObj.FilePara.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//init multi part
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(
		reqObj.FilePara.FilePara,
		filepath.Base(reqObj.FilePara.FilePath),
	)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	//add extend parameters
	for key, val := range reqObj.Params {
		v2, ok := val.(string)
		if !ok {
			continue
		}
		_ = writer.WriteField(key, v2)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", reqObj.Url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

//general request
func (q *HttpClient) generalReq(reqObj *HttpReq) (*http.Request, error) {
	var (
		tempStr string
		buffer = bytes.NewBuffer(nil)
		req *http.Request
		err error
	)

	//check parameters
	if len(reqObj.Params) > 0 {
		i := 0
		for k, v := range reqObj.Params {
			if i > 0 {
				buffer.WriteString("&")
			}
			tempStr = fmt.Sprintf("%s=%v", k, v)
			buffer.WriteString(tempStr)
			i++
		}
	}

	//get request method
	switch reqObj.Kind {
	case HttpReqPost:
		{
			if reqObj.Body != nil {
				buffer.Write(reqObj.Body)
			}
			//int post req
			req, err = http.NewRequest("POST", reqObj.Url, strings.NewReader(buffer.String()))
			if req.Header == nil || len(req.Header) <= 0 {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		}
	default:
		//init get req
		if buffer != nil && buffer.Len() > 0 {
			reqObj.Url = fmt.Sprintf("%s?%s", reqObj.Url, buffer.String())
		}
		req, err = http.NewRequest("GET", reqObj.Url, nil)
	}

	//format post form
	if reqObj.Params != nil {
		req.Form = make(url.Values)
		for k, v := range reqObj.Params {
			keyVal := fmt.Sprintf("%v", v)
			req.Form.Add(k, keyVal)
		}
	}

	return req, err
}

//send original http request and get response
func (q *HttpClient) sendHttpReq(reqObj *HttpReq) ([]byte, error) {
	var (
		req *http.Request
		err error
	)

	//basic check
	if q.client == nil || reqObj == nil {
		return nil, errors.New("invalid parameter")
	}
	if reqObj.Kind < HttpReqGet || reqObj.Url == "" {
		return nil, errors.New("invalid parameter")
	}

	//check and fix http prefix
	if !strings.HasPrefix(reqObj.Url, HttpReqPrefix) &&
		!strings.HasPrefix(reqObj.Url, HttpsReqPrefix) {
		//fix it
		reqObj.Url = fmt.Sprintf("%v%v", HttpReqPrefix, reqObj.Url)
	}

	if reqObj.FilePara.FilePath != "" &&
		reqObj.FilePara.FilePara != "" {
		//file upload request
		req, err = q.fileUploadReq(reqObj)
	}else{
		//general request
		req, err = q.generalReq(reqObj)
	}
	if err != nil {
		log.Println("HttpClient::sendHttpReq, create request failed, err:", err.Error())
		return nil, err
	}

	//set headers
	if reqObj.Headers != nil {
		for k, v := range reqObj.Headers {
			req.Header.Set(k, v)
		}
	}

	//set http connect close
	req.Header.Set("Connection", "close")
	req.Close = true

	//begin send request
	//c.client.Timeout = time.Second
	resp, err := q.client.Do(req)
	if err != nil {
		log.Println("HttpClient::sendHttpReq, send http request failed, err:", err.Error())
		return nil, err
	}

	//close resp before return
	defer resp.Body.Close()

	//read response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("HttpClient::sendHttpReq, read response body failed, err:", err.Error())
		return nil, err
	}

	//return response
	return respBody, nil
}

//inter init
func (q *HttpClient) interInit() {
	//init http trans
	netTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: HttpClientTimeOut * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: HttpClientTimeOut * time.Second,
		ResponseHeaderTimeout: HttpClientTimeOut * time.Second,
		ExpectContinueTimeout: HttpClientTimeOut* time.Second,
	}

	//init native http client
	q.client = &http.Client{
		Timeout:time.Second * HttpClientTimeOut,
		Transport:netTransport,
	}
}