package graph

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Executable interface {
	Execute() error
}

type HTTPMethod string

const (
	HTTPGet     HTTPMethod = "GET"
	HTTPHead    HTTPMethod = "HEAD"
	HTTPPost    HTTPMethod = "POST"
	HTTPPut     HTTPMethod = "PUT"
	HTTPPatch   HTTPMethod = "PATCH"
	HTTPDelete  HTTPMethod = "DELETE"
	HTTPConnect HTTPMethod = "CONNECT"
	HTTPOptions HTTPMethod = "OPTIONS"
	HTTPTrace   HTTPMethod = "TRACE"
)

func HTTPMethodFromString(s string) (m HTTPMethod, err error) {
	switch upper := strings.ToUpper(s); upper {
	case string(HTTPGet):
		m = HTTPGet
	case string(HTTPHead):
		m = HTTPHead
	case string(HTTPPost):
		m = HTTPPost
	case string(HTTPPut):
		m = HTTPPut
	case string(HTTPPatch):
		m = HTTPPatch
	case string(HTTPDelete):
		m = HTTPDelete
	case string(HTTPConnect):
		m = HTTPConnect
	case string(HTTPOptions):
		m = HTTPOptions
	case string(HTTPTrace):
		m = HTTPTrace
	default:
		err = fmt.Errorf("%s is not a valid HTTP method", s)
	}
	return
}

type ConcurrentCommand struct {
	Commands []Executable
}

func (c ConcurrentCommand) Execute() error {
	wg := sync.WaitGroup{}
	wg.Add(len(c.Commands))
	for _, cmd := range c.Commands {
		go func(exe Executable) {
			// TODO: Handle error.
			exe.Execute()
			wg.Done()
		}(cmd)
	}
	wg.Wait()
	return nil
}

type RequestCommand struct {
	RequestSettings
	ServiceName string
	HTTPMethod  HTTPMethod
}

func (c RequestCommand) Execute() error {
	// TODO: Send c.HTTPMethod request to c.ServiceName with c.Settings.
	return nil
}

type SleepCommand struct {
	Duration time.Duration
}

func (c SleepCommand) Execute() error {
	time.Sleep(c.Duration)
	return nil
}
