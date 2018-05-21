package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/Tahler/service-grapher/pkg/consts"
	"github.com/Tahler/service-grapher/pkg/graph"
	"github.com/Tahler/service-grapher/pkg/graph/svc"
	"github.com/Tahler/service-grapher/pkg/graph/svctype"
	"github.com/ghodss/yaml"
)

var (
	serviceGraphYAMLFilePath = path.Join(
		consts.ConfigPath, consts.ServiceGraphYAMLFileName)

	// Set by init().
	serviceTypes map[string]svctype.ServiceType
	service      svc.Service
)

func init() {
	serviceGraph, err := readServiceGraphFromYAMLFile(serviceGraphYAMLFilePath)
	if err != nil {
		log.Fatal(err)
	}

	serviceTypes = extractServiceTypes(serviceGraph)

	name, ok := os.LookupEnv(consts.ServiceNameEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, consts.ServiceNameEnvKey)
	}
	service, ok = lookupService(serviceGraph, name)
	if !ok {
		log.Fatalf(`service with name "%s" does not exist`, name)
	}
	logService(service)
}

func readServiceGraphFromYAMLFile(
	path string) (serviceGraph graph.ServiceGraph, err error) {
	graphYAML, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	log.Printf("Unmarshalling\n%s", graphYAML)
	err = yaml.Unmarshal(graphYAML, &serviceGraph)
	if err != nil {
		return
	}
	return
}

func extractServiceTypes(
	serviceGraph graph.ServiceGraph) map[string]svctype.ServiceType {
	serviceTypes := make(
		map[string]svctype.ServiceType, len(serviceGraph.Services))
	for _, service := range serviceGraph.Services {
		serviceTypes[service.Name] = service.Type
	}
	return serviceTypes
}

func lookupService(
	serviceGraph graph.ServiceGraph, name string) (service svc.Service, ok bool) {
	for _, svc := range serviceGraph.Services {
		if svc.Name == name {
			service = svc
			ok = true
			break
		}
	}
	return
}

func logService(service svc.Service) {
	serviceYAML, err := yaml.Marshal(service)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Acting as service %s:\n%s", service.Name, serviceYAML)
}
