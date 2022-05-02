// Package graphviz converts service graphs into Graphviz DOT language.
package graphviz

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/maxfouquet/isotope/convert/pkg/graph"
	"github.com/maxfouquet/isotope/convert/pkg/graph/script"
	"github.com/maxfouquet/isotope/convert/pkg/graph/svc"
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
	Type         string
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
    fontname = "courier"
    shape = plaintext
  ];

  {{ range .Nodes -}}
  {{ .Name }} [label=<
<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0">
  <TR><TD><B>{{ .Name }}</B><BR />Type: {{ .Type }}<BR />Err: {{ .ErrorRate }}</TD></TR>
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
	exe script.Command, idx int, fromServiceName string) (edges []Edge) {
	switch cmd := exe.(type) {
	case script.ConcurrentCommand:
		for _, subCmd := range cmd {
			subEdges := getEdgesFromExe(subCmd, idx, fromServiceName)
			for _, e := range subEdges {
				edges = append(edges, e)
			}
		}
	case script.RequestCommand:
		e := Edge{
			From:      fromServiceName,
			To:        cmd.ServiceName,
			StepIndex: idx,
		}
		edges = append(edges, e)
	}
	return
}

func toGraphvizNode(service svc.Service) (Node, []Edge, error) {
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
		Type:         service.Type.String(),
		ErrorRate:    service.ErrorRate.String(),
		ResponseSize: service.ResponseSize.String(),
		Steps:        steps,
	}
	return n, edges, nil
}

func nonConcurrentCommandToString(exe script.Command) (s string, err error) {
	switch cmd := exe.(type) {
	case script.SleepCommand:
		s = fmt.Sprintf("SLEEP %s", cmd)
	case script.RequestCommand:
		s = fmt.Sprintf(
			"CALL \"%s\" %s",
			cmd.ServiceName, cmd.Size.String())
	default:
		err = fmt.Errorf("unexpected type of executable %T", exe)
	}
	return
}

func executableToStringSlice(exe script.Command) (ss []string, err error) {
	appendNonConcurrentExe := func(exe script.Command) error {
		s, err := nonConcurrentCommandToString(exe)
		if err != nil {
			return err
		}
		ss = append(ss, s)
		return nil
	}
	switch cmd := exe.(type) {
	case script.SleepCommand:
		err = appendNonConcurrentExe(exe)
	case script.RequestCommand:
		err = appendNonConcurrentExe(exe)
	case script.ConcurrentCommand:
		for _, exe := range cmd {
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
