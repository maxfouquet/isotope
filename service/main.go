package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/Tahler/isotope/convert/pkg/consts"
	"github.com/Tahler/isotope/convert/pkg/graph/svc"
	"github.com/Tahler/isotope/service/pkg/srv"
	"github.com/Tahler/isotope/service/pkg/srv/prometheus"
	"github.com/ghodss/yaml"
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
	log.SetLogLevel(log.Debug)

	serviceGraph, err := srv.ServiceGraphFromYAMLFile(serviceGraphYAMLFilePath)
	if err != nil {
		log.Fatalf("%s", err)
	}

	name, ok := os.LookupEnv(consts.ServiceNameEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, consts.ServiceNameEnvKey)
	}

	service, err := srv.ExtractService(serviceGraph, name)
	logService(service)

	defaultHandler := srv.Handler{Service: service}

	err = serveWithPrometheus(defaultHandler)
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func serveWithPrometheus(defaultHandler http.Handler) (err error) {
	log.Infof(`exposing Prometheus endpoint "%s"`, promEndpoint)
	http.Handle(promEndpoint, prometheus.Handler())

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

func logService(service svc.Service) error {
	if log.Log(log.Info) {
		serviceYAML, err := yaml.Marshal(service)
		if err != nil {
			return err
		}
		log.Infof("acting as service %s:\n%s", service.Name, serviceYAML)
	}
	return nil
}
