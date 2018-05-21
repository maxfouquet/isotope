package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/Tahler/service-grapher/pkg/consts"
	"github.com/Tahler/service-grapher/pkg/graph"
	"github.com/Tahler/service-grapher/pkg/graph/svc"
	"github.com/ghodss/yaml"
)

var (
	serviceGraphYAMLFilePath = path.Join(
		consts.ConfigPath, consts.ServiceGraphYAMLFileName)

	// Set by init().
	serviceGraph graph.ServiceGraph
	service      svc.Service
)

func init() {
	var err error
	serviceGraph, err = readServiceGraphFromYAMLFile(serviceGraphYAMLFilePath)
	if err != nil {
		log.Fatal(err)
	}
	name, ok := os.LookupEnv(consts.ServiceNameEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, consts.ServiceNameEnvKey)
	}
	service, ok = lookupService(serviceGraph, name)
	if !ok {
		log.Fatalf(`service with name "%s" does not exist`, name)
	}
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
