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
	MethodGet     HTTPMethod = "GET"
	MethodHead    HTTPMethod = "HEAD"
	MethodPost    HTTPMethod = "POST"
	MethodPut     HTTPMethod = "PUT"
	MethodPatch   HTTPMethod = "PATCH"
	MethodDelete  HTTPMethod = "DELETE"
	MethodConnect HTTPMethod = "CONNECT"
	MethodOptions HTTPMethod = "OPTIONS"
	MethodTrace   HTTPMethod = "TRACE"
)

func HTTPMethodFromString(s string) (m HTTPMethod, err error) {
	switch upper := strings.ToUpper(s); upper {
	case string(MethodGet):
		m = MethodGet
	case string(MethodHead):
		m = MethodHead
	case string(MethodPost):
		m = MethodPost
	case string(MethodPut):
		m = MethodPut
	case string(MethodPatch):
		m = MethodPatch
	case string(MethodDelete):
		m = MethodDelete
	case string(MethodConnect):
		m = MethodConnect
	case string(MethodOptions):
		m = MethodOptions
	case string(MethodTrace):
		m = MethodTrace
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
	Method      HTTPMethod
}

func (c RequestCommand) Execute() error {
	// TODO: Send c.Method request to c.ServiceName with c.Settings.
	return nil
}

type SleepCommand struct {
	Duration time.Duration
}

func (c SleepCommand) Execute() error {
	time.Sleep(c.Duration)
	return nil
}
