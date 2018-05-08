package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = 8080

func echoHandler(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Echoing %s to client %s", request.URL.Path, request.RemoteAddr)
	request.Write(writer)
}

func main() {
	http.HandleFunc("/", echoHandler)
	log.Printf("Listening on port %v\n", port)
	addr := fmt.Sprintf(":%v", port)
	http.ListenAndServe(addr, nil)
}
