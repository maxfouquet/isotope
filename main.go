package main

import (
	"encoding/json"
	"fmt"

	"github.com/Tahler/service-grapher/pkg/graph"

	"gopkg.in/yaml.v2"
)

var data = `
apiVersion: v1alpha1
default:
  computeUsage: 1%
  memoryUsage: 10%
  errorRate: 0.05
  payloadSize: 1KB
services:
  A:
    script:
    - - get: B
      - get: C
    - sleep: 100ms
    - post:
        service: B
        payloadSize: 512KB
`

func main() {
	var g graph.ServiceGraph
	err := yaml.Unmarshal([]byte(data), &g)
	if err == nil {
		pp, _ := json.MarshalIndent(g, "", "  ")
		fmt.Printf("%s\n", pp)
	} else {
		fmt.Printf("%v\n", err)
	}
}
