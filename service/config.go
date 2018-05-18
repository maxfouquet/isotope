package main

import (
	"io/ioutil"
	"log"

	"github.com/Tahler/service-grapher/pkg/graph/svc"
	"github.com/ghodss/yaml"
)

const configFilePath = "/etc/config/service.yaml"

func getService() (service svc.Service, err error) {
	serviceYAML, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return
	}
	log.Printf("Unmarshalling\n%s", serviceYAML)
	err = yaml.Unmarshal(serviceYAML, &service)
	return
}
