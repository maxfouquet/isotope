// Converts a YAML document into a PNG via graphviz.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/docker/go-units"
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

	s, err := toString(g)
	panicIfErr(err)
	err = writeStringToFile(s, "output.gv")
	panicIfErr(err)
}

func writeStringToFile(s string, fileName string) (err error) {
	f, err := os.Create("output.gv")
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.WriteString(s)
	return
}

func toGraphvizGraph(sg graph.ServiceGraph) graphvizGraph {
	nodes := make([]node, 0, len(sg.Services))
	edges := make([]edge, 0, len(sg.Services))
	for _, service := range sg.Services {
		node, connections := toGraphvizNode(service)
		nodes = append(nodes, node)
		for _, connection := range connections {
			edges = append(edges, connection)
		}
	}
	return graphvizGraph{
		Nodes: nodes,
		Edges: edges,
	}
}

func getEdgesFromExe(
	exe graph.Executable, idx int, fromServiceName string) (edges []edge) {
	switch cmd := exe.(type) {
	case graph.ConcurrentCommand:
		for _, subCmd := range cmd.Commands {
			subEdges := getEdgesFromExe(subCmd, idx, fromServiceName)
			for _, e := range subEdges {
				edges = append(edges, e)
			}
		}
	case graph.RequestCommand:
		e := edge{
			From:      fromServiceName,
			To:        cmd.ServiceName,
			StepIndex: idx,
		}
		edges = append(edges, e)
	case graph.SleepCommand:
	default:
	}
	return
}

func toGraphvizNode(service graph.Service) (node, []edge) {
	steps := make([][]string, 0, len(service.Script))
	edges := make([]edge, 0, len(service.Script))
	for idx, exe := range service.Script {
		step, err := executableToStringSlice(exe)
		panicIfErr(err)
		steps = append(steps, step)

		stepEdges := getEdgesFromExe(exe, idx, service.Name)
		for _, e := range stepEdges {
			edges = append(edges, e)
		}
	}
	n := node{
		Name:         service.Name,
		ComputeUsage: toPercentage(service.ComputeUsage),
		MemoryUsage:  toPercentage(service.MemoryUsage),
		ErrorRate:    toPercentage(service.ErrorRate),
		Steps:        steps,
	}
	return n, edges
}

func nonConcurrentCommandToString(exe graph.Executable) (s string, err error) {
	switch cmd := exe.(type) {
	case graph.SleepCommand:
		s = fmt.Sprintf("SLEEP %s", cmd.Duration)
	case graph.RequestCommand:
		readablePayloadSize := units.BytesSize(float64(cmd.PayloadSize))
		s = fmt.Sprintf(
			"%s \"%s\" %s",
			cmd.HTTPMethod, cmd.ServiceName, readablePayloadSize)
	default:
		err = fmt.Errorf("unexpected type of executable %T", exe)
	}
	return
}

func executableToStringSlice(exe graph.Executable) (ss []string, err error) {
	appendNonConcurrentExe := func(exe graph.Executable) {
		s, err := nonConcurrentCommandToString(exe)
		panicIfErr(err)
		ss = append(ss, s)
	}
	switch cmd := exe.(type) {
	case graph.SleepCommand:
		appendNonConcurrentExe(exe)
	case graph.RequestCommand:
		appendNonConcurrentExe(exe)
	case graph.ConcurrentCommand:
		for _, exe := range cmd.Commands {
			appendNonConcurrentExe(exe)
		}
	default:
		err = fmt.Errorf("unexpected type of executable %T", exe)
	}
	return
}

func toPercentage(f float64) string {
	p := f * 100
	return fmt.Sprintf("%.1f%%", p)
}

type graphvizGraph struct {
	Nodes []node
	Edges []edge
}

type node struct {
	Name         string
	ComputeUsage string
	MemoryUsage  string
	ErrorRate    string
	Steps        [][]string
}

type edge struct {
	From      string
	To        string
	StepIndex int
}

const graphvizTemplate = `digraph {
    node [
        fontsize = "16"
        shape = plaintext
    ];

    {{ range .Nodes -}}
    {{ .Name }} [label=<
<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0">
  <TR><TD><B>{{ .Name }}</B><BR />CPU: {{ .ComputeUsage }}<BR />RAM: {{ .MemoryUsage }}<BR />Err: {{ .ErrorRate }}</TD></TR>
  {{- range $i, $cmds := .Steps }}
  <TR><TD PORT="{{ $i }}">
    {{- range $j, $cmd := $cmds -}}
        {{- if $j -}}<BR />{{- end -}}
        {{- $cmd -}}
    {{- end -}}
  </TD></TR>
  {{- end }}
</TABLE>>];

	{{ end }}

    {{- range .Edges }}
    {{ .From -}}:{{- .StepIndex }} -> {{ .To }}
    {{- end }}
}
`

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func toString(g graphvizGraph) (s string, err error) {
	tmpl, err := template.New("digraph").Parse(graphvizTemplate)
	if err != nil {
		return
	}
	var b bytes.Buffer
	if err = tmpl.Execute(&b, g); err == nil {
		s = b.String()
	}
	return
}
