package util

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

/*
 * running log service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//internal macro variables
const (
	LogFileCheckRate = 60 //xxx seconds
)

//face info
type LogService struct {
	path      string    //log root path
	prefix    string    //log file prefix
	outFile   *os.File  //file handler
	lastDay   string    //last day for change file
	lastHour  string    //last hour for change file
	closeChan chan bool //chan for close process
	sync.RWMutex
}

//construct
func NewLogService(path, prefix string) *LogService {
	this := &LogService{
		path:path,
		prefix:prefix,
		closeChan:make(chan bool),
	}
	this.lastHour = this.getCurHour()
	this.lastDay = this.getCurDay()
	//register first log file
	this.registerLogFile()
	//spawn son process for sync file handler
	go this.runCheckProcess()
	return this
}

//close
func (l *LogService) Close() bool {
	if l.outFile == nil {
		return false
	}
	if l.closeChan != nil {
		l.closeChan <- true
	}
	//need sleep awhile for internal clean up
	time.Sleep(time.Second/10)
	return true
}

////////////////////
//private function
///////////////////

//get current hour info
func (l *LogService) getCurHour() string {
	now := time.Now()
	allSlice := strings.Split(now.String(), " ")
	day := allSlice[0]
	time := allSlice[1]
	timeSlice := strings.Split(time, ":")
	hour := timeSlice[0]
	return fmt.Sprintf("%s-%s", day, hour)
}

//get current day info
func (l *LogService) getCurDay() string {
	now := time.Now()
	allSlice := strings.Split(now.String(), " ")
	day := allSlice[0]
	return day
}

//register local log file into log.xxx command
func (l *LogService) registerLogFile() bool {
	var err error
	l.Lock()
	defer l.Unlock()
	if l.outFile != nil {
		//close old file
		l.outFile.Close()
		l.outFile = nil
	}
	file := fmt.Sprintf("%s/%s-%s.log", l.path, l.prefix, l.lastDay)
	l.outFile, err = os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Open log file failed, error:", err.Error())
		return false
	}
	//bind on log function
	log.SetOutput(l.outFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	return true
}

//check and sync log handler
func (l *LogService) syncFileHandler() bool {
	curDay := l.getCurDay()
	if curDay == l.lastDay {
		//same day, do nothing
		return false
	}
	//sync last day
	l.lastDay = curDay
	//force register new log file
	l.registerLogFile()
	return true
}

//internal check process
func (l *LogService) runCheckProcess() {
	var (
		tick = time.Tick(time.Second * LogFileCheckRate)
		m any = nil
	)
	//defer
	defer func() {
		if err := recover(); err != m {
			log.Printf("logService:checkProcess panic, err:%v\n", err)
		}
		if l.outFile != nil {
			l.outFile.Close()
			l.outFile = nil
		}
	}()

	//loop
	for {
		select {
		case <- tick:
			//check and sync file handler
			l.syncFileHandler()
		case <- l.closeChan:
			return
		}
	}
}