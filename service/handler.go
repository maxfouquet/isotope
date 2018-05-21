package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/Tahler/service-grapher/pkg/graph/svc"
)

// pathTracesHeaderKey must be in Train-Case.
const pathTracesHeaderKey = "Path-Traces"

var hostname = os.Getenv("HOSTNAME")

type serviceHandler struct {
	svc.Service
}

func (h serviceHandler) ServeHTTP(
	writer http.ResponseWriter, request *http.Request) {

	respond := func(status int, paths []string, isLocalErr bool) {
		stampHeader(writer.Header(), paths, isLocalErr)

		log.Printf(
			"Echoing (%v) to client %s, %s=%v",
			status, request.RemoteAddr,
			pathTracesHeaderKey, writer.Header()[pathTracesHeaderKey])
		writer.WriteHeader(status)
		request.Write(writer)
	}

	if err := h.errorChance(); err != nil {
		respond(http.StatusInternalServerError, nil, true)
		return
	}

	allPaths := make([]string, 0, len(h.Script))
	for _, step := range h.Script {
		forwardableHeader := extractForwardableHeader(request.Header)
		paths, err := execute(step, forwardableHeader)
		for _, path := range paths {
			allPaths = append(allPaths, path)
		}
		if err != nil {
			log.Println(err)
			respond(http.StatusInternalServerError, allPaths, false)
			return
		}
	}

	respond(http.StatusOK, allPaths, false)
}

func stampHeader(header http.Header, paths []string, isLocalErr bool) {
	stamp := fmt.Sprintf("%s (%s)", service.Name, hostname)
	if isLocalErr {
		stamp += " (ERROR)"
	}

	var stampedPaths []string
	if len(paths) == 0 {
		stampedPaths = []string{stamp}
	} else {
		stampedPaths = stampPaths(paths, stamp)
	}

	header[pathTracesHeaderKey] = stampedPaths
}

func stampPaths(paths []string, stamp string) []string {
	stampedPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		stampedPath := fmt.Sprintf("%s %s", stamp, path)
		stampedPaths = append(stampedPaths, stampedPath)
	}
	return stampedPaths
}

// errorChance randomly returns an error h.ErrorRate percent of the time.
func (h serviceHandler) errorChance() (err error) {
	random := rand.Float64()
	if random < float64(h.ErrorRate) {
		err = fmt.Errorf("server randomly failed with a chance of %v", h.ErrorRate)
	}
	return
}
