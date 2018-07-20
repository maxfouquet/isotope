package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/Tahler/isotope/convert/pkg/consts"
	"github.com/Tahler/isotope/service/pkg/srv"
	"istio.io/fortio/log"
)

const (
	promEndpoint    = "/metrics"
	defaultEndpoint = "/"
)

var (
	serviceGraphYAMLFilePath = path.Join(
		consts.ConfigPath, consts.ServiceGraphYAMLFileName)
)

func main() {
	log.SetLogLevel(log.Info)

	setMaxIdleConnectionsPerHost(100)

	serviceName, ok := os.LookupEnv(consts.ServiceNameEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, consts.ServiceNameEnvKey)
	}

	defaultHandler, err := srv.HandlerFromServiceGraphYAML(
		serviceGraphYAMLFilePath, serviceName)
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = serveWithPrometheus(defaultHandler)
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func serveWithPrometheus(defaultHandler srv.Handler) (err error) {
	log.Infof(`exposing Prometheus endpoint "%s"`, promEndpoint)
	promHandler, err := defaultHandler.Metrics.Handler()
	if err != nil {
		return
	}
	http.Handle(promEndpoint, promHandler)

	log.Infof(`exposing default endpoint "%s"`, defaultEndpoint)
	http.Handle(defaultEndpoint, defaultHandler)

	addr := fmt.Sprintf(":%d", consts.ServicePort)
	log.Infof("listening on port %v\n", consts.ServicePort)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		return
	}
	return
}

func setMaxIdleConnectionsPerHost(n int) {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = n
}
