package util

import (
	"bufio"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

/*
 * system signal catch face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - watch system signal
 * - notify relate refer chan when shutdown
 * use steps
 * shutDownChan := make(chan bool, 1)
 * s := NewSignal()
 * s.RegisterShutDownChan(shutDownChan)
 * s.MonSignal()
 */

//inter macro define
const (
	ShutDownWaitSeconds = 5 //default 5 seconds
	SigUserShutDown = syscall.Signal(0xa)
)

//face info
type Signal struct {
	SIGTERM  int32
	SIGINT   int32
	waitSeconds int
	shutDownChan chan bool //refer chan slice
	ch chan os.Signal
	stopSig chan bool
	cbForQuit func()
	initDone bool
}

//construct, step-1
func NewSignal(waitSeconds ...int) *Signal {
	//check
	waitSecond := ShutDownWaitSeconds
	if waitSeconds != nil && len(waitSeconds) > 0 {
		waitSecond = waitSeconds[0]
	}
	//self init
	this := &Signal{
		waitSeconds:waitSecond,
		shutDownChan:make(chan bool, 1),
		ch:make(chan os.Signal, 1),
		stopSig:make(chan bool, 1),
	}
	return this
}

//register shut down chan, step-2
func (f *Signal) RegisterShutDownChan(ch chan bool, cb func()) error {
	//check
	if ch == nil || cb == nil {
		return errors.New("invalid parameter")
	}

	//sync
	f.shutDownChan = ch
	f.cbForQuit = cb

	if !f.initDone {
		//spawn son process to receive message
		go f.receiveMsg()
		f.initDone = true
	}
	return nil
}

//monitor signal, step-3
func (f *Signal) MonSignal() {
	//signal notify
	signal.Notify(
		f.ch,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGKILL,
	)

	//spawn win32 signal check
	go f.checkSignalOfWin32(f.ch)

	go func(f *Signal) {
		//receive signal
		for {
			msg := <- f.ch
			switch msg {
			case syscall.SIGHUP:
				fallthrough
			case syscall.SIGTERM:
				atomic.StoreInt32(&f.SIGTERM, 1)
				f.onExit(msg)
			case syscall.SIGINT:
				atomic.StoreInt32(&f.SIGINT, 1)
				f.onExit(msg)
			case syscall.SIGQUIT:
				f.onExit(msg)
			case SigUserShutDown:
				f.forceQuit()
			default:
				log.Printf("get signal of %v\n", msg.String())
			}
		}
	}(f)
}

//force quit
func (f *Signal) ForceNotify() {
	//check
	if f.shutDownChan == nil {
		return
	}
	//send notify to chan
	f.shutDownChan <- true
}

///////////////
//private func
///////////////

//receive shut down message
func (f *Signal) receiveMsg() {
	select {
	case <- f.shutDownChan:
		{
			if f.cbForQuit != nil {
				f.cbForQuit()
			}
			return
		}
	}
}

//force quit
func (f *Signal) forceQuit() {
	f.notifyShutDownChan()
	os.Exit(-1)
}

//exit
func (f *Signal) onExit(msg os.Signal) {
	f.notifyShutDownChan()
	<- f.stopSig
	os.Exit(-1)
}

//notify shutdown chan
func (f *Signal) notifyShutDownChan() {
	//check
	if f.shutDownChan == nil {
		return
	}

	//send notify to relate chan
	f.shutDownChan <- true

	//sleep for a while
	duration := time.Duration(f.waitSeconds) * time.Second
	time.Sleep(duration)

	//stop notify
	f.stopSig <- true
}

//check signal of win32
func (f *Signal) checkSignalOfWin32(c chan <- os.Signal) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmd := scanner.Text()
		cmd, _ = f.parseCmd(cmd)
		if cmd == "quit" {
			c <- syscall.SIGINT
		}
	}
	c <- syscall.SIGINT
}

//parse input command
//return command and args
func (f *Signal) parseCmd(str string) (string, string) {
	for i, c := range str {
		if c == ' ' {
			return strings.TrimSpace(str[:i]), strings.TrimSpace(str[i+1:])
		}
	}
	return str, ""
}
