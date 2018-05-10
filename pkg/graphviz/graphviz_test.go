package graphviz

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/Tahler/service-grapher/pkg/graph"
)

func TestServiceGraphToGraph(t *testing.T) {
	expected := Graph{
		Nodes: []Node{
			Node{
				Name:         "A",
				ComputeUsage: "50.00%",
				MemoryUsage:  "20.00%",
				ErrorRate:    "0.01%",
				Steps: [][]string{
					[]string{
						"SLEEP 100ms",
					},
				},
			},
			Node{
				Name:         "B",
				ComputeUsage: "10.00%",
				MemoryUsage:  "10.00%",
				ErrorRate:    "0.00%",
				Steps:        [][]string{},
			},
			Node{
				Name:         "C",
				ComputeUsage: "10.00%",
				MemoryUsage:  "10.00%",
				ErrorRate:    "0.00%",
				Steps: [][]string{
					[]string{
						"GET \"A\" 10KiB",
					},
					[]string{
						"POST \"B\" 1KiB",
					},
				},
			},
			Node{
				Name:         "D",
				ComputeUsage: "10.00%",
				MemoryUsage:  "10.00%",
				ErrorRate:    "0.00%",
				Steps: [][]string{
					[]string{
						"GET \"A\" 1KiB",
						"GET \"C\" 1KiB",
					},
					[]string{
						"SLEEP 10ms",
					},
					[]string{
						"DELETE \"B\" 1KiB",
					},
				},
			},
		},
		Edges: []Edge{
			Edge{
				From:      "C",
				To:        "A",
				StepIndex: 0,
			},
			Edge{
				From:      "C",
				To:        "B",
				StepIndex: 1,
			},
			Edge{
				From:      "D",
				To:        "A",
				StepIndex: 0,
			},
			Edge{
				From:      "D",
				To:        "C",
				StepIndex: 0,
			},
			Edge{
				From:      "D",
				To:        "B",
				StepIndex: 2,
			},
		},
	}

	serviceGraph := graph.ServiceGraph{
		Services: map[string]graph.Service{
			"A": graph.Service{
				Name: "A",
				ServiceSettings: graph.ServiceSettings{
					ComputeUsage: 0.5,
					MemoryUsage:  0.2,
					ErrorRate:    0.0001,
				},
				Script: []graph.Command{
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
				Script: []graph.Command{
					graph.RequestCommand{
						HTTPMethod:  "GET",
						ServiceName: "A",
						RequestSettings: graph.RequestSettings{
							Size: 10240,
						},
					},
					graph.RequestCommand{
						HTTPMethod:  "POST",
						ServiceName: "B",
						RequestSettings: graph.RequestSettings{
							Size: 1024,
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
				Script: []graph.Command{
					graph.ConcurrentCommand{
						Commands: []graph.Command{
							graph.RequestCommand{
								HTTPMethod:  "GET",
								ServiceName: "A",
								RequestSettings: graph.RequestSettings{
									Size: 1024,
								},
							},
							graph.RequestCommand{
								HTTPMethod:  "GET",
								ServiceName: "C",
								RequestSettings: graph.RequestSettings{
									Size: 1024,
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
							Size: 1024,
						},
					},
				},
			},
		},
	}
	actual, err := ServiceGraphToGraph(serviceGraph)
	if err != nil {
		t.Fatal(err)
	}
	if !graphsAreEqual(expected, actual) {
		t.Errorf("\nexpect: %+v, \nactual: %+v", expected, actual)
	}
}

func graphsAreEqual(left Graph, right Graph) bool {
	sortNodes(left.Nodes)
	sortEdges(left.Edges)
	sortNodes(right.Nodes)
	sortEdges(right.Edges)
	return reflect.DeepEqual(left, right)
}

func sortNodes(nodes []Node) {
	sort.SliceStable(nodes, func(i int, j int) bool {
		leftNode := nodes[i]
		rightNode := nodes[j]
		return leftNode.Name < rightNode.Name
	})
}

// sortEdges sorts edges by From, StepIndex, To.
func sortEdges(edges []Edge) {
	sort.SliceStable(edges, func(i int, j int) bool {
		leftEdge := edges[i]
		rightEdge := edges[j]
		return leftEdgeIsLessThanRightEdge(leftEdge, rightEdge)
	})
}

func leftEdgeIsLessThanRightEdge(left Edge, right Edge) (isLess bool) {
	if left.From == right.From {
		if left.StepIndex == right.StepIndex {
			isLess = left.To < right.To
		} else {
			isLess = left.StepIndex < right.StepIndex
		}
	} else {
		isLess = left.From < right.From
	}
	return
}
