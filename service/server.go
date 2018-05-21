package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Tahler/service-grapher/pkg/consts"
)

func main() {
	handler := serviceHandler{Service: service}
	log.Printf("Listening on port %v\n", consts.ServicePort)
	addr := fmt.Sprintf(":%v", consts.ServicePort)
	http.ListenAndServe(addr, handler)
}
