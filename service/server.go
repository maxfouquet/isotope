package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Tahler/service-grapher/pkg/graph"
)

const port = 8080

func main() {
	service := readService()
	handler := serviceHandler{Service: service}
	log.Printf("Listening on port %v\n", port)
	addr := fmt.Sprintf(":%v", port)
	http.ListenAndServe(addr, handler)
}

type serviceHandler struct {
	graph.Service
}

func (h serviceHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	for _, cmd := range h.Script {
		cmd.Execute()
	}
	log.Printf("Echoing %s to client %s", request.URL.Path, request.RemoteAddr)
	request.Write(writer)
}

func readService() graph.Service {
	return graph.Service{
		ServiceSettings: graph.ServiceSettings{
			ComputeUsage: 0.1,
			MemoryUsage:  0.2,
			ErrorRate:    0.5,
		},
		Name: "A",
		Script: []graph.Executable{
			graph.SleepCommand{Duration: 5 * time.Second},
		},
	}
}
