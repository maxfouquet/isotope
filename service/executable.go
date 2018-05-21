package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Tahler/service-grapher/pkg/graph/script"
	multierror "github.com/hashicorp/go-multierror"
)

func execute(step interface{}, forwardableHeader http.Header) (
	paths []string, err error) {
	switch cmd := step.(type) {
	case script.SleepCommand:
		executeSleepCommand(cmd)
	case script.RequestCommand:
		paths, err = executeRequestCommand(cmd, forwardableHeader)
	case script.ConcurrentCommand:
		paths, err = executeConcurrentCommand(cmd, forwardableHeader)
	default:
		log.Fatalf("unknown command type in script: %T", cmd)
	}
	return
}

func executeSleepCommand(cmd script.SleepCommand) {
	time.Sleep(time.Duration(cmd))
}

// Execute sends an HTTP request to another service. Assumes DNS is available
// which maps exe.ServiceName to the relevant URL to reach the service.
func executeRequestCommand(
	cmd script.RequestCommand, forwardableHeader http.Header) (
	paths []string, err error) {
	destName := cmd.ServiceName
	response, err := sendRequest(destName, cmd.Size, forwardableHeader)
	if err != nil {
		return
	}
	paths = response.Header[pathTracesHeaderKey]
	log.Printf("%s responded with %s", destName, response.Status)
	if response.StatusCode == http.StatusInternalServerError {
		err = fmt.Errorf("service %s responded with %s", destName, response.Status)
	}
	return
}

// executeConcurrentCommand calls each command in exe.Commands asynchronously
// and waits for each to complete.
func executeConcurrentCommand(
	cmd script.ConcurrentCommand, forwardableHeader http.Header) (
	paths []string, errs error) {
	numSubCmds := len(cmd)
	wg := sync.WaitGroup{}
	wg.Add(numSubCmds)
	pathsChan := make(chan []string, numSubCmds)
	for _, subCmd := range cmd {
		go func(step interface{}) {
			defer wg.Done()

			// TODO: Split err into actual error and random errorRate-caused error.
			stepPaths, err := execute(step, forwardableHeader)
			pathsChan <- stepPaths
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}(subCmd)
	}
	wg.Wait()
	close(pathsChan)
	for returnedPaths := range pathsChan {
		for _, path := range returnedPaths {
			paths = append(paths, path)
		}
	}
	return
}
