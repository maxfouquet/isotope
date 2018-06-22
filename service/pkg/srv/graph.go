package srv

import (
	"fmt"
	"io/ioutil"

	"github.com/Tahler/isotope/convert/pkg/graph"
	"github.com/Tahler/isotope/convert/pkg/graph/svc"
	"github.com/Tahler/isotope/convert/pkg/graph/svctype"
	"github.com/ghodss/yaml"
	"istio.io/fortio/log"
)

// ServiceGraphFromYAMLFile unmarshals the ServiceGraph from the YAML at path.
func ServiceGraphFromYAMLFile(
	path string) (serviceGraph graph.ServiceGraph, err error) {
	graphYAML, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	log.Infof("unmarshalling\n%s", graphYAML)
	err = yaml.Unmarshal(graphYAML, &serviceGraph)
	if err != nil {
		return
	}
	return
}

// ExtractService finds the service in serviceGraph with the specified name.
func ExtractService(
	serviceGraph graph.ServiceGraph, name string) (
	service svc.Service, err error) {
	for _, svc := range serviceGraph.Services {
		if svc.Name == name {
			service = svc
			return
		}
	}
	err = fmt.Errorf(
		"service with name %s does not exist in %v", name, serviceGraph)
	return
}

// ExtractServiceTypes builds a map from service name to its type
// (i.e. HTTP or gRPC).
func ExtractServiceTypes(
	serviceGraph graph.ServiceGraph) map[string]svctype.ServiceType {
	types := make(map[string]svctype.ServiceType, len(serviceGraph.Services))
	for _, service := range serviceGraph.Services {
		types[service.Name] = service.Type
	}
	return types
}
