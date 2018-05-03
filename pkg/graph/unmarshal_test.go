package graph

import (
	"reflect"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"
)

func TestUnmarshalYAML(t *testing.T) {
	expected := ServiceGraph{
		Services: map[string]Service{
			"A": Service{
				Name: "A",
				ServiceSettings: ServiceSettings{
					ComputeUsage: 0.5,
					MemoryUsage:  0.2,
					ErrorRate:    0.0001,
				},
				Script: []Executable{
					SleepCommand{
						Duration: 100 * time.Millisecond,
					},
				},
			},
			"B": Service{
				Name: "B",
				ServiceSettings: ServiceSettings{
					ComputeUsage: 0.1,
					MemoryUsage:  0.1,
					ErrorRate:    0,
				},
			},
			"C": Service{
				Name: "C",
				ServiceSettings: ServiceSettings{
					ComputeUsage: 0.1,
					MemoryUsage:  0.1,
					ErrorRate:    0,
				},
				Script: []Executable{
					RequestCommand{
						HTTPMethod:  "GET",
						ServiceName: "A",
						RequestSettings: RequestSettings{
							PayloadSize: 10240,
						},
					},
					RequestCommand{
						HTTPMethod:  "POST",
						ServiceName: "B",
						RequestSettings: RequestSettings{
							PayloadSize: 1024,
						},
					},
				},
			},
			"D": Service{
				Name: "D",
				ServiceSettings: ServiceSettings{
					ComputeUsage: 0.1,
					MemoryUsage:  0.1,
					ErrorRate:    0,
				},
				Script: []Executable{
					ConcurrentCommand{
						Commands: []Executable{
							RequestCommand{
								HTTPMethod:  "GET",
								ServiceName: "A",
								RequestSettings: RequestSettings{
									PayloadSize: 1024,
								},
							},
							RequestCommand{
								HTTPMethod:  "GET",
								ServiceName: "C",
								RequestSettings: RequestSettings{
									PayloadSize: 1024,
								},
							},
						},
					},
					SleepCommand{
						Duration: 10 * time.Millisecond,
					},
					RequestCommand{
						HTTPMethod:  "DELETE",
						ServiceName: "B",
						RequestSettings: RequestSettings{
							PayloadSize: 1024,
						},
					},
				},
			},
		},
	}

	inputYAML := `apiVersion: v1alpha1
default:
  payloadSize: 1 KB
  computeUsage: 10%
  memoryUsage: 10%
services:
  A:
    computeUsage: 50%
    memoryUsage: 20%
    errorRate: 0.01%
    script:
    - sleep: 100ms
  B:
  C:
    script:
    - get:
        service: A
        payloadSize: 10K
    - post: B
  D:
    # Call A and C concurrently, process, then call B.
    script:
    - - get: A
      - get: C
    - sleep: 10ms
    - delete: B
`
	var actual ServiceGraph
	err := yaml.Unmarshal([]byte(inputYAML), &actual)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected: %v, actual: %v", expected, actual)
	}
}
