package graphviz

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/Tahler/isotope/pkg/graph"
	"github.com/Tahler/isotope/pkg/graph/script"
	"github.com/Tahler/isotope/pkg/graph/svc"
	"github.com/Tahler/isotope/pkg/graph/svctype"
)

func TestServiceGraphToGraph(t *testing.T) {
	expected := Graph{
		Nodes: []Node{
			Node{
				Name:         "a",
				Type:         "HTTP",
				ErrorRate:    "0.01%",
				ResponseSize: "10KiB",
				Steps: [][]string{
					[]string{
						"SLEEP 100ms",
					},
				},
			},
			Node{
				Name:         "b",
				Type:         "gRPC",
				ErrorRate:    "0.00%",
				ResponseSize: "10KiB",
				Steps:        [][]string{},
			},
			Node{
				Name:         "c",
				Type:         "HTTP",
				ErrorRate:    "0.00%",
				ResponseSize: "10KiB",
				Steps: [][]string{
					[]string{
						"CALL \"a\" 10KiB",
					},
					[]string{
						"CALL \"b\" 1KiB",
					},
				},
			},
			Node{
				Name:         "d",
				Type:         "HTTP",
				ErrorRate:    "0.00%",
				ResponseSize: "10KiB",
				Steps: [][]string{
					[]string{
						"CALL \"a\" 1KiB",
						"CALL \"c\" 1KiB",
					},
					[]string{
						"SLEEP 10ms",
					},
					[]string{
						"CALL \"b\" 1KiB",
					},
				},
			},
		},
		Edges: []Edge{
			Edge{
				From:      "c",
				To:        "a",
				StepIndex: 0,
			},
			Edge{
				From:      "c",
				To:        "b",
				StepIndex: 1,
			},
			Edge{
				From:      "d",
				To:        "a",
				StepIndex: 0,
			},
			Edge{
				From:      "d",
				To:        "c",
				StepIndex: 0,
			},
			Edge{
				From:      "d",
				To:        "b",
				StepIndex: 2,
			},
		},
	}

	serviceGraph := graph.ServiceGraph{
		Services: []svc.Service{
			{
				Name:         "a",
				Type:         svctype.ServiceHTTP,
				ErrorRate:    0.0001,
				ResponseSize: 10240,
				Script: []script.Command{
					script.SleepCommand(100 * time.Millisecond),
				},
			},
			{
				Name:         "b",
				Type:         svctype.ServiceGRPC,
				ErrorRate:    0,
				ResponseSize: 10240,
			},
			{
				Name:         "c",
				Type:         svctype.ServiceHTTP,
				ErrorRate:    0,
				ResponseSize: 10240,
				Script: []script.Command{
					script.RequestCommand{
						ServiceName: "a",
						Size:        10240,
					},
					script.RequestCommand{
						ServiceName: "b",
						Size:        1024,
					},
				},
			},
			{
				Name:         "d",
				Type:         svctype.ServiceHTTP,
				ErrorRate:    0,
				ResponseSize: 10240,
				Script: []script.Command{
					script.ConcurrentCommand([]script.Command{
						script.RequestCommand{
							ServiceName: "a",
							Size:        1024,
						},
						script.RequestCommand{
							ServiceName: "c",
							Size:        1024,
						},
					}),
					script.SleepCommand(10 * time.Millisecond),
					script.RequestCommand{
						ServiceName: "b",
						Size:        1024,
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
	// sortNodes(left.Nodes)
	// sortEdges(left.Edges)
	// sortNodes(right.Nodes)
	// sortEdges(right.Edges)
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
