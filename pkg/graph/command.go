package graph

import (
	"fmt"
	"strings"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
)

// Executable is the top-level interface for commands.
type Executable interface {
	Execute() error
}

// HTTPMethod is an enum-like string type describing any of the HTTP request
// method: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods.
type HTTPMethod string

// Each constant below is a possible HTTP request method.
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

// HTTPMethodFromString case-insensitively converts a string to a HTTPMethod.
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

// ConcurrentCommand describes a set of commands that should be executed
// simultaneously.
type ConcurrentCommand struct {
	Commands []Executable
}

// Execute calls each command in c.Commands asynchronously and waits for each to
// complete.
func (c ConcurrentCommand) Execute() error {
	wg := sync.WaitGroup{}
	wg.Add(len(c.Commands))
	var errs *multierror.Error
	for _, cmd := range c.Commands {
		go func(exe Executable) {
			err := exe.Execute()
			errs = multierror.Append(errs, err)
			wg.Done()
		}(cmd)
	}
	wg.Wait()
	return nil
}

// RequestCommand describes a command to send an HTTP request to another
// service.
type RequestCommand struct {
	RequestSettings
	ServiceName string
	HTTPMethod  HTTPMethod
}

// Execute sends an HTTP request to another service.
func (c RequestCommand) Execute() error {
	// TODO: Send c.HTTPMethod request to c.ServiceName with c.Settings.
	return nil
}

// SleepCommand describes a command to pause for a duration.
type SleepCommand struct {
	Duration time.Duration
}

// Execute sleeps for c.Duration.
func (c SleepCommand) Execute() error {
	time.Sleep(c.Duration)
	return nil
}
