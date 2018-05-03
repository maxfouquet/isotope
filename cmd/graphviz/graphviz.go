// Converts a YAML document into a PNG via graphviz.
package main

import (
	"bytes"
	"fmt"
	"text/template"
)

func main() {
	// TODO: args
	// TODO: convert yaml to svcgraph
	// TODO: convert svcgraph to values
	values := graphvizValues{
		Services: []Service{
			Service{
				Name:         "A",
				ComputeUsage: "50%",
				MemoryUsage:  "20%",
				ErrorRate:    "0.01%",
				Steps: [][]string{
					[]string{
						"SLEEP 100ms",
					},
				},
			},
			Service{
				Name:         "B",
				ComputeUsage: "10%",
				MemoryUsage:  "10%",
				ErrorRate:    "0%",
			},
			Service{
				Name:         "C",
				ComputeUsage: "10%",
				MemoryUsage:  "10%",
				ErrorRate:    "0%",
				Steps: [][]string{
					[]string{
						"GET \"A\" 10K",
					},
					[]string{
						"POST \"B\" 1K",
					},
				},
			},
			Service{
				Name:         "D",
				ComputeUsage: "10%",
				MemoryUsage:  "10%",
				ErrorRate:    "0%",
				Steps: [][]string{
					[]string{
						"GET \"A\" 1K",
						"GET \"C\" 1K",
					},
					[]string{
						"SLEEP 10ms",
					},
					[]string{
						"DELETE \"B\" 1K",
					},
				},
			},
		},
		Connections: []Connection{
			Connection{
				From:      "D",
				To:        "A",
				StepIndex: 0,
			},
			Connection{
				From:      "D",
				To:        "C",
				StepIndex: 0,
			},
			Connection{
				From:      "D",
				To:        "B",
				StepIndex: 2,
			},
			Connection{
				From:      "C",
				To:        "A",
				StepIndex: 0,
			},
			Connection{
				From:      "C",
				To:        "B",
				StepIndex: 1,
			},
		},
	}
	s, err := fromTemplate(values)
	panicIfErr(err)
	fmt.Println(s)
	// f, _ := os.Create("output.gv")
	// defer f.Close()
	// f.WriteString(s)
}

type graphvizValues struct {
	Services    []Service
	Connections []Connection
}

type Service struct {
	Name         string
	ComputeUsage string
	MemoryUsage  string
	ErrorRate    string
	Steps        [][]string
}

type Connection struct {
	From      string
	To        string
	StepIndex int
}

const graphvizTemplate = `digraph {
    node [
        fontsize = "16"
        shape = plaintext
    ];

    {{ range .Services -}}
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

    {{- range .Connections }}
    {{ .From -}}:{{- .StepIndex }} -> {{ .To }}
    {{- end }}
}
`

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func fromTemplate(values graphvizValues) (s string, err error) {
	tmpl, err := template.New("digraph").Parse(graphvizTemplate)
	if err != nil {
		return
	}
	var b bytes.Buffer
	if err = tmpl.Execute(&b, values); err == nil {
		s = b.String()
	}
	return
}
