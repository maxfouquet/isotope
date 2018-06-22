package main

import (
	"fmt"
	"net/http"

	"github.com/Tahler/isotope/convert/pkg/consts"
	"github.com/Tahler/isotope/service/pkg/srv"
	"github.com/Tahler/isotope/service/pkg/srv/prometheus"
	"istio.io/fortio/log"
)

const (
	promEndpoint    = "/metrics"
	defaultEndpoint = "/"
)

func main() {
	log.SetLogLevel(log.Debug)

	log.Infof(`exposing Prometheus endpoint "%s"`, promEndpoint)
	http.Handle(promEndpoint, prometheus.Handler())

	log.Infof(`exposing default endpoint "%s"`, defaultEndpoint)
	http.Handle(defaultEndpoint, srv.Handler{Service: srv.Service})

	addr := fmt.Sprintf(":%d", consts.ServicePort)
	log.Infof("listening on port %v\n", consts.ServicePort)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("%s", err)
	}
}
