package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = 8080

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	service, err := getService()
	if err != nil {
		log.Fatal(err)
	}
	handler := serviceHandler{Service: service}
	log.Printf("Listening on port %v\n", port)
	addr := fmt.Sprintf(":%v", port)
	http.ListenAndServe(addr, handler)
}
