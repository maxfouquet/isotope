package main

import (
	"io/ioutil"
	"log"

	"github.com/Tahler/service-grapher/pkg/graph"
	yaml "gopkg.in/yaml.v2"
)

const configFilePath = "/etc/config/service.yaml"

func getService() (service graph.Service, err error) {
	serviceYAML, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return
	}
	log.Printf("Unmarshalling\n%s", serviceYAML)
	err = yaml.Unmarshal([]byte(serviceYAML), &service)
	return
}
