package graph

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/Tahler/isotope/pkg/graph/script"
	"github.com/Tahler/isotope/pkg/graph/svc"

	"github.com/Tahler/isotope/pkg/graph/svctype"
)

func TestServiceGraph_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input []byte
		graph ServiceGraph
		err   error
	}{
		{jsonWithOneService, graphWithOneService, nil},
		{jsonWithDefaultsAndManyServices, graphWithDefaultsAndManyServices, nil},
		{
			jsonWithRequestToUndefinedService,
			ServiceGraph{},
			ErrRequestToUndefinedService{"b"},
		},
		{
			jsonWithNestedConcurrentCommand,
			ServiceGraph{},
			ErrNestedConcurrentCommand,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var graph ServiceGraph
			err := json.Unmarshal(test.input, &graph)
			if test.err == nil {
				if !reflect.DeepEqual(test.graph, graph) {
					t.Errorf("expected %v; actual %v", test.graph, graph)
				}
			} else {
				if test.err != err {
					t.Errorf("expected %v; actual %v", test.err, err)
				}
			}
		})
	}
}

var (
	jsonWithOneService = []byte(`
		{
			"apiVersion": "v1alpha1",
			"services": [{"name": "a"}]
		}
	`)
	graphWithOneService = ServiceGraph{[]svc.Service{
		{
			Name: "a",
			Type: svctype.ServiceHTTP,
		},
	}}
	jsonWithDefaultsAndManyServices = []byte(`
		{
			"apiVersion": "v1alpha1",
			"defaults": {
				"errorRate": 0.1,
				"requestSize": 516,
				"responseSize": 128,
				"script": [
					{ "sleep": "100ms" }
				]
			},
			"services": [
				{
					"name": "a"
				},
				{
					"name": "b",
					"script": [
						{
							"call": {
								"service": "a",
								"size": "1KiB"
							}
						},
						{ "sleep": "10ms" }
					]
				},
				{
					"name": "c",
					"type": "grpc",
					"errorRate": "20%",
					"responseSize": "1K",
					"script": [
						[
							{ "call": "a" },
							{ "call": "b" }
						],
						{ "sleep": "10ms" }
					]
				}
			]
		}
	`)
	graphWithDefaultsAndManyServices = ServiceGraph{[]svc.Service{
		{
			Name:         "a",
			Type:         svctype.ServiceHTTP,
			ErrorRate:    0.1,
			ResponseSize: 128,
			Script: script.Script([]script.Command{
				script.SleepCommand(100 * time.Millisecond),
			}),
		},
		{
			Name:         "b",
			Type:         svctype.ServiceHTTP,
			ErrorRate:    0.1,
			ResponseSize: 128,
			Script: script.Script([]script.Command{
				script.RequestCommand{ServiceName: "a", Size: 1024},
				script.SleepCommand(10 * time.Millisecond),
			}),
		},
		{
			Name:         "c",
			Type:         svctype.ServiceGRPC,
			ErrorRate:    0.2,
			ResponseSize: 1024,
			Script: script.Script([]script.Command{
				script.ConcurrentCommand{
					script.RequestCommand{ServiceName: "a", Size: 516},
					script.RequestCommand{ServiceName: "b", Size: 516},
				},
				script.SleepCommand(10 * time.Millisecond),
			}),
		},
	}}
	jsonWithRequestToUndefinedService = []byte(`
		{
			"services": [
				{
					"name": "a",
					"script": [{ "call": "b"}]
				}
			]
		}
	`)
	jsonWithNestedConcurrentCommand = []byte(`
		{
			"services": [
				{
					"name": "a"
				},
				{
					"name": "b",
					"script": [
						[
							[{ "call": "a" }, { "call": "a" }],
							{ "sleep": "10ms" }
						]
					]
				}
			]
		}
	`)
)
