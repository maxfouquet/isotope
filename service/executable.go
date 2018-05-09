package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

// Execute sends an HTTP request to another service.
func (exe RequestExecutable) Execute() (err error) {
	url, err := getServiceURL(exe.ServiceName)
	if err != nil {
		return
	}
	request, err := http.NewRequest(string(exe.HTTPMethod), url, nil)
	// TODO: set payload size
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
	return
}

// getServiceURL builds a URL using environment variables to reach the service.
// https://kubernetes.io/docs/concepts/services-networking/service/#environment-variables
func getServiceURL(serviceName string) (url string, err error) {
	serviceEnvName := strings.ToUpper(strings.Replace(serviceName, "-", "_", -1))

	hostKey := fmt.Sprintf("%s_SERVICE_HOST", serviceEnvName)
	host, ok := os.LookupEnv(hostKey)
	if !ok {
		err = fmt.Errorf(
			"no environment variable for host of service %s exists: %s=%s",
			serviceName, hostKey, host)
		return
	}

	portKey := fmt.Sprintf("%s_SERVICE_PORT", serviceEnvName)
	port, ok := os.LookupEnv(portKey)
	if !ok {
		err = fmt.Errorf(
			"no environment variable for port of service %s exists: %s=%s",
			serviceName, portKey, port)
		return
	}

	url = fmt.Sprintf("http://%s:%s", host, port)
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
