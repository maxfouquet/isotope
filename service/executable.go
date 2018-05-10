package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Tahler/service-grapher/pkg/graph"

	multierror "github.com/hashicorp/go-multierror"
)

// Executable is the top-level interface for commands.
type Executable interface {
	Execute() error
}

// SleepExecutable makes graph.SleepCommand a receiver for Execute.
type SleepExecutable graph.SleepCommand

// Execute sleeps for exe.Duration.
func (exe SleepExecutable) Execute() error {
	time.Sleep(exe.Duration)
	return nil
}

// RequestExecutable makes graph.RequestCommand a receiver for Execute.
type RequestExecutable struct {
	graph.RequestCommand
	http.Header
}

// Execute sends an HTTP request to another service. Assumes DNS is available
// which maps exe.ServiceName to the relevant URL to reach the service.
func (exe RequestExecutable) Execute() (err error) {
	url := fmt.Sprintf("http://%s:%v", exe.ServiceName, port)
	request, err := buildRequest(
		exe.HTTPMethod, url, exe.RequestSettings.Size, exe.Header)
	if err != nil {
		return
	}
	log.Printf(
		"Sending %s request to %s (%s)", exe.HTTPMethod, exe.ServiceName, url)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	log.Printf("%s responded with %s", exe.ServiceName, response.Status)
	if response.StatusCode == http.StatusInternalServerError {
		err = fmt.Errorf(
			"service %s responded with %s", exe.ServiceName, response.Status)
	}
	return
}

func buildRequest(
	method graph.HTTPMethod, url string, size int64, requestHeader http.Header) (
	request *http.Request, err error) {
	payload := make([]byte, size, size)
	request, err = http.NewRequest(string(method), url, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	copyHeader(request, requestHeader)
	return
}

func copyHeader(request *http.Request, header http.Header) {
	for key, values := range header {
		request.Header[key] = values
	}
}

// ConcurrentExecutable makes graph.ConcurrentCommand a receiver for Execute.
type ConcurrentExecutable struct {
	Executables []Executable
}

// Execute calls each command in exe.Commands asynchronously and waits for each
// to complete.
func (exe ConcurrentExecutable) Execute() error {
	wg := sync.WaitGroup{}
	wg.Add(len(exe.Executables))
	var errs *multierror.Error
	for _, subExe := range exe.Executables {
		go func(subExe Executable) {
			err := subExe.Execute()
			errs = multierror.Append(errs, err)
			wg.Done()
		}(subExe)
	}
	wg.Wait()
	return errs
}

func toConcurrentExecutable(
	cmd graph.ConcurrentCommand, requestHeader http.Header) (
	ConcurrentExecutable, error) {
	exe := ConcurrentExecutable{
		Executables: make([]Executable, 0, len(cmd.Commands)),
	}
	for subCmd := range cmd.Commands {
		subExe, err := toExecutable(subCmd, requestHeader)
		if err != nil {
			return exe, err
		}
		exe.Executables = append(exe.Executables, subExe)
	}
	return exe, nil
}

func toExecutable(
	step interface{}, requestHeader http.Header) (exe Executable, err error) {
	switch cmd := step.(type) {
	case graph.SleepCommand:
		exe = SleepExecutable(cmd)
	case graph.RequestCommand:
		exe = RequestExecutable{
			RequestCommand: cmd,
			Header:         requestHeader,
		}
	case graph.ConcurrentCommand:
		exe, err = toConcurrentExecutable(cmd, requestHeader)
	default:
		err = fmt.Errorf("unknown type %T", cmd)
	}
	return
}
