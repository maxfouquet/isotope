// Package graphviz converts service graphs into Graphviz DOT language.
package graphviz

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/docker/go-units"

	"github.com/Tahler/service-grapher/pkg/graph"
)

// ServiceGraphToDotLanguage converts a ServiceGraph to a Graphviz DOT language
// string.
func ServiceGraphToDotLanguage(
	serviceGraph graph.ServiceGraph) (dotLang string, err error) {
	graph, err := ServiceGraphToGraph(serviceGraph)
	if err != nil {
		return
	}
	dotLang, err = GraphToDotLanguage(graph)
	return
}

// GraphToDotLanguage converts a graphviz graph to a Graphviz DOT language
// string via a template.
func GraphToDotLanguage(g Graph) (dotLang string, err error) {
	tmpl, err := template.New("digraph").Parse(graphvizTemplate)
	if err != nil {
		return
	}
	var b bytes.Buffer
	if err = tmpl.Execute(&b, g); err == nil {
		dotLang = b.String()
	}
	return
}

// ServiceGraphToGraph converts a service graph to a graphviz graph.
func ServiceGraphToGraph(sg graph.ServiceGraph) (Graph, error) {
	nodes := make([]Node, 0, len(sg.Services))
	edges := make([]Edge, 0, len(sg.Services))
	for _, service := range sg.Services {
		node, connections, err := toGraphvizNode(service)
		if err != nil {
			return Graph{}, err
		}
		nodes = append(nodes, node)
		for _, connection := range connections {
			edges = append(edges, connection)
		}
	}
	return Graph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

// Graph represents a Graphviz graph.
type Graph struct {
	Nodes []Node
	Edges []Edge
}

// Node represents a node in the Graphviz graph.
type Node struct {
	Name         string
	ComputeUsage string
	MemoryUsage  string
	ErrorRate    string
	ResponseSize string
	Steps        [][]string
}

// Edge represents a directed edge in the Graphviz graph.
type Edge struct {
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

func getEdgesFromExe(
	exe graph.Command, idx int, fromServiceName string) (edges []Edge) {
	switch cmd := exe.(type) {
	case graph.ConcurrentCommand:
		for _, subCmd := range cmd.Commands {
			subEdges := getEdgesFromExe(subCmd, idx, fromServiceName)
			for _, e := range subEdges {
				edges = append(edges, e)
			}
		}
	case graph.RequestCommand:
		e := Edge{
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

func toGraphvizNode(service graph.Service) (Node, []Edge, error) {
	steps := make([][]string, 0, len(service.Script))
	edges := make([]Edge, 0, len(service.Script))
	for idx, exe := range service.Script {
		step, err := executableToStringSlice(exe)
		if err != nil {
			return Node{}, nil, err
		}
		steps = append(steps, step)

		stepEdges := getEdgesFromExe(exe, idx, service.Name)
		for _, e := range stepEdges {
			edges = append(edges, e)
		}
	}
	n := Node{
		Name:         service.Name,
		ComputeUsage: toPercentage(service.ComputeUsage),
		MemoryUsage:  toPercentage(service.MemoryUsage),
		ErrorRate:    toPercentage(service.ErrorRate),
		ResponseSize: units.BytesSize(float64(service.ResponseSize)),
		Steps:        steps,
	}
	return n, edges, nil
}

func nonConcurrentCommandToString(exe graph.Command) (s string, err error) {
	switch cmd := exe.(type) {
	case graph.SleepCommand:
		s = fmt.Sprintf("SLEEP %s", cmd.Duration)
	case graph.RequestCommand:
		readableRequestSize := units.BytesSize(float64(cmd.Size))
		s = fmt.Sprintf(
			"%s \"%s\" %s",
			cmd.HTTPMethod, cmd.ServiceName, readableRequestSize)
	default:
		err = fmt.Errorf("unexpected type of executable %T", exe)
	}
	return
}

func executableToStringSlice(exe graph.Command) (ss []string, err error) {
	appendNonConcurrentExe := func(exe graph.Command) error {
		s, err := nonConcurrentCommandToString(exe)
		if err != nil {
			return err
		}
		ss = append(ss, s)
		return nil
	}
	switch cmd := exe.(type) {
	case graph.SleepCommand:
		err = appendNonConcurrentExe(exe)
	case graph.RequestCommand:
		err = appendNonConcurrentExe(exe)
	case graph.ConcurrentCommand:
		for _, exe := range cmd.Commands {
			err = appendNonConcurrentExe(exe)
			if err != nil {
				return
			}
		}
	default:
		err = fmt.Errorf("unexpected type of executable %T", exe)
	}
	return
}

func toPercentage(f float64) string {
	p := f * 100
	return fmt.Sprintf("%.2f%%", p)
}
