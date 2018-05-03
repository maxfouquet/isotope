package main

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
	"github.com/Tahler/service-grapher/pkg/graph"
)

func main() {
	// TODO: args

	yamlContents, err := ioutil.ReadFile("input.yaml")
	panicIfErr(err)

	var serviceGraph graph.ServiceGraph
	err = yaml.Unmarshal(yamlContents, &serviceGraph)
	panicIfErr(err)

	g := toGraphvizGraph(serviceGraph)

	s, err := toGraphvizDotLanguage(g)
	panicIfErr(err)

	err = ioutil.WriteFile("output.gv", []byte(s), 0644)
	panicIfErr(err)
}
