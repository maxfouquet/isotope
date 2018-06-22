package srv

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/Tahler/isotope/convert/pkg/consts"
	"github.com/Tahler/isotope/convert/pkg/graph"
	"github.com/Tahler/isotope/convert/pkg/graph/svc"
	"github.com/Tahler/isotope/convert/pkg/graph/svctype"
	"github.com/ghodss/yaml"
	"istio.io/fortio/log"
)

var (
	serviceGraphYAMLFilePath = path.Join(
		consts.ConfigPath, consts.ServiceGraphYAMLFileName)

	// Set by init().
	serviceTypes map[string]svctype.ServiceType
	// Service is initialized by init() as read from the config.
	Service svc.Service
)

func init() {
	serviceGraph, err := readServiceGraphFromYAMLFile(serviceGraphYAMLFilePath)
	if err != nil {
		log.Fatalf("%s", err)
	}

	serviceTypes = extractServiceTypes(serviceGraph)

	name, ok := os.LookupEnv(consts.ServiceNameEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, consts.ServiceNameEnvKey)
	}
	Service, ok = lookupService(serviceGraph, name)
	if !ok {
		log.Fatalf(`service with name "%s" does not exist`, name)
	}
	logService(Service)
}

func readServiceGraphFromYAMLFile(
	path string) (serviceGraph graph.ServiceGraph, err error) {
	graphYAML, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	log.Debugf("unmarshalling\n%s", graphYAML)
	err = yaml.Unmarshal(graphYAML, &serviceGraph)
	if err != nil {
		return
	}
	return
}

func extractServiceTypes(
	serviceGraph graph.ServiceGraph) map[string]svctype.ServiceType {
	types := make(
		map[string]svctype.ServiceType, len(serviceGraph.Services))
	for _, service := range serviceGraph.Services {
		types[service.Name] = service.Type
	}
	return types
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
		log.Fatalf("%s", err)
	}
	log.Infof("acting as service %s:\n%s", service.Name, serviceYAML)
}
