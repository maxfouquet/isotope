package graph

import (
	"fmt"
	"strings"
	"time"
)

// Command is the top level interface for commands.
type Command interface{}

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
	Commands []Command
}

// RequestCommand describes a command to send an HTTP request to another
// service.
type RequestCommand struct {
	RequestSettings
	ServiceName string
	HTTPMethod  HTTPMethod
}

// SleepCommand describes a command to pause for a duration.
type SleepCommand struct {
	Duration time.Duration
}
