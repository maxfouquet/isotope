package main

import (
	"fmt"

	"github.com/Tahler/service-grapher/pkg/graph"

	"gopkg.in/yaml.v2"
)

var data = `
apiVersion: v1alpha1
default:
  computeUsage: 1%
  memoryUsage: 10%
  errorRate: 5%
  payloadSize: 1KB
services:
  A:
    computeUsage: 10%
  B:
    computeUsage: 5%
    memoryUsage: 1%
    errorRate: 5%
  C:
`

func main() {
	var g graph.ServiceGraph
	err := yaml.Unmarshal([]byte(data), &g)
	if err == nil {
		fmt.Printf("%+v\n", g)
	} else {
		fmt.Printf("%v\n", err)
	}
}
