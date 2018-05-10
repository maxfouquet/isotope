package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/Tahler/service-grapher/pkg/graph"
)

const port = 8080

func main() {
	service, err := getService()
	if err != nil {
		log.Fatal(err)
	}
	handler := serviceHandler{Service: service}
	log.Printf("Listening on port %v\n", port)
	addr := fmt.Sprintf(":%v", port)
	http.ListenAndServe(addr, handler)
}

type serviceHandler struct {
	graph.Service
}

func (h serviceHandler) ServeHTTP(
	writer http.ResponseWriter, request *http.Request) {

	respond := func(status int) {
		log.Printf("Echoing (%v) to client %s", status, request.RemoteAddr)
		writer.WriteHeader(status)
		request.Write(writer)
	}

	if err := h.errorChance(); err != nil {
		respond(http.StatusInternalServerError)
		return
	}

	for _, step := range h.Script {
		forwardableHeader := extractForwardableHeader(request.Header)
		exe, err := toExecutable(step, forwardableHeader)
		if err != nil {
			log.Fatalf("error in script: %s", err)
			return
		}

		err = exe.Execute()
		if err != nil {
			log.Println(err)
			respond(http.StatusInternalServerError)
			return
		}
	}

	respond(http.StatusOK)
}

// errorChance randomly returns an error h.ErrorRate percent of the time.
func (h serviceHandler) errorChance() (err error) {
	random := rand.Float64()
	if random < h.ErrorRate {
		err = fmt.Errorf("server randomly failed with a chance of %v", h.ErrorRate)
	}
	return
}
