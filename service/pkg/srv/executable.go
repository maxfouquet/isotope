package srv

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/Tahler/isotope/convert/pkg/graph/script"
	"github.com/Tahler/isotope/convert/pkg/graph/svctype"
	"github.com/Tahler/isotope/service/pkg/srv/prometheus"
	multierror "github.com/hashicorp/go-multierror"
	"istio.io/fortio/log"
)

func execute(
	step interface{},
	forwardableHeader http.Header,
	serviceTypes map[string]svctype.ServiceType,
	metrics prometheus.Metrics) (paths []string, err error) {
	switch cmd := step.(type) {
	case script.SleepCommand:
		executeSleepCommand(cmd)
	case script.RequestCommand:
		paths, err = executeRequestCommand(
			cmd, forwardableHeader, serviceTypes, metrics)
	case script.ConcurrentCommand:
		paths, err = executeConcurrentCommand(
			cmd, forwardableHeader, serviceTypes, metrics)
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
	cmd script.RequestCommand,
	forwardableHeader http.Header,
	serviceTypes map[string]svctype.ServiceType,
	metrics prometheus.Metrics) (paths []string, err error) {
	destName := cmd.ServiceName
	destType, ok := serviceTypes[destName]
	if !ok {
		err = fmt.Errorf("service %s does not exist", destName)
		return
	}
	response, err := sendRequest(destName, destType, cmd.Size, forwardableHeader)
	if err != nil {
		return
	}
	metrics.RecordRequestSent(destName, uint64(cmd.Size))
	paths = response.Header[pathTracesHeaderKey]
	log.Debugf("%s responded with %s", destName, response.Status)
	if response.StatusCode == http.StatusInternalServerError {
		err = fmt.Errorf("service %s responded with %s", destName, response.Status)
	}

	// Necessary for reusing HTTP/1.x "keep-alive" TCP connections.
	// https://golang.org/pkg/net/http/#Response
	readAllAndClose(response.Body)

	return
}

func readAllAndClose(r io.ReadCloser) {
	io.Copy(ioutil.Discard, r)
	r.Close()
}

// executeConcurrentCommand calls each command in exe.Commands asynchronously
// and waits for each to complete.
func executeConcurrentCommand(
	cmd script.ConcurrentCommand,
	forwardableHeader http.Header,
	serviceTypes map[string]svctype.ServiceType,
	metrics prometheus.Metrics) (paths []string, errs error) {
	numSubCmds := len(cmd)
	wg := sync.WaitGroup{}
	wg.Add(numSubCmds)
	pathsChan := make(chan []string, numSubCmds)
	for _, subCmd := range cmd {
		go func(step interface{}) {
			defer wg.Done()

			// TODO: Differentiate between actual error and errorRate-caused error.
			stepPaths, err := execute(step, forwardableHeader, serviceTypes, metrics)
			pathsChan <- stepPaths
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}(subCmd)
	}
	wg.Wait()
	close(pathsChan)
	for returnedPaths := range pathsChan {
		paths = append(paths, returnedPaths...)
	}
	return
}
