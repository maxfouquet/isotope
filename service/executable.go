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
type RequestExecutable graph.RequestCommand

// Execute sends an HTTP request to another service. Assumes DNS is available
// which maps exe.ServiceName to the relevant URL to reach the service.
func (exe RequestExecutable) Execute() (err error) {
	url := fmt.Sprintf("http://%s:%v", exe.ServiceName, port)
	payload := make([]byte, exe.RequestSettings.Size, exe.RequestSettings.Size)
	request, err := http.NewRequest(
		string(exe.HTTPMethod), url, bytes.NewBuffer(payload))
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

// ConcurrentExecutable makes graph.ConcurrentCommand a receiver for Execute.
type ConcurrentExecutable graph.ConcurrentCommand

// Execute calls each command in exe.Commands asynchronously and waits for each
// to complete.
func (exe ConcurrentExecutable) Execute() error {
	wg := sync.WaitGroup{}
	wg.Add(len(exe.Commands))
	var errs *multierror.Error
	for _, cmd := range exe.Commands {
		subExe, err := toExecutable(cmd)
		if err != nil {
			return err
		}
		go func(subExe Executable) {
			err := subExe.Execute()
			errs = multierror.Append(errs, err)
			wg.Done()
		}(subExe)
	}
	wg.Wait()
	return errs
}

func toExecutable(step interface{}) (exe Executable, err error) {
	switch cmd := step.(type) {
	case graph.SleepCommand:
		exe = SleepExecutable(cmd)
	case graph.RequestCommand:
		exe = RequestExecutable(cmd)
	case graph.ConcurrentCommand:
		exe = ConcurrentExecutable(cmd)
	default:
		err = fmt.Errorf("unknown type %T", cmd)
	}
	return
}
