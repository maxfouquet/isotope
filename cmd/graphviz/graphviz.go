// Converts a YAML document into a PNG via graphviz.
package main

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/docker/go-units"

	"github.com/Tahler/service-grapher/pkg/graph"
)

func main() {
	// TODO: args
	// TODO: convert yaml to svcgraph
	serviceGraph := graph.ServiceGraph{
		Services: map[string]graph.Service{
			"A": graph.Service{
				Name: "A",
				ServiceSettings: graph.ServiceSettings{
					ComputeUsage: 0.5,
					MemoryUsage:  0.2,
					ErrorRate:    0.0001,
				},
				Script: []graph.Executable{
					graph.SleepCommand{
						Duration: 100 * time.Millisecond,
					},
				},
			},
			"B": graph.Service{
				Name: "B",
				ServiceSettings: graph.ServiceSettings{
					ComputeUsage: 0.1,
					MemoryUsage:  0.1,
					ErrorRate:    0,
				},
			},
			"C": graph.Service{
				Name: "C",
				ServiceSettings: graph.ServiceSettings{
					ComputeUsage: 0.1,
					MemoryUsage:  0.1,
					ErrorRate:    0,
				},
				Script: []graph.Executable{
					graph.RequestCommand{
						HTTPMethod:  "GET",
						ServiceName: "A",
						RequestSettings: graph.RequestSettings{
							PayloadSize: 10240,
						},
					},
					graph.RequestCommand{
						HTTPMethod:  "POST",
						ServiceName: "B",
						RequestSettings: graph.RequestSettings{
							PayloadSize: 1024,
						},
					},
				},
			},
			"D": graph.Service{
				Name: "D",
				ServiceSettings: graph.ServiceSettings{
					ComputeUsage: 0.1,
					MemoryUsage:  0.1,
					ErrorRate:    0,
				},
				Script: []graph.Executable{
					graph.ConcurrentCommand{
						Commands: []graph.Executable{
							graph.RequestCommand{
								HTTPMethod:  "GET",
								ServiceName: "A",
								RequestSettings: graph.RequestSettings{
									PayloadSize: 1024,
								},
							},
							graph.RequestCommand{
								HTTPMethod:  "GET",
								ServiceName: "C",
								RequestSettings: graph.RequestSettings{
									PayloadSize: 1024,
								},
							},
						},
					},
					graph.SleepCommand{
						Duration: 10 * time.Millisecond,
					},
					graph.RequestCommand{
						HTTPMethod:  "DELETE",
						ServiceName: "B",
						RequestSettings: graph.RequestSettings{
							PayloadSize: 1024,
						},
					},
				},
			},
		},
	}
	g := toGraphvizGraph(serviceGraph)
	s, err := toString(g)
	panicIfErr(err)
	fmt.Println(s)

	// f, _ := os.Create("output.gv")
	// defer f.Close()
	// f.WriteString(s)
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
