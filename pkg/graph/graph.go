package graph

import "fmt"

type ServiceGraph struct {
	Services map[string]Service
}

func (g *ServiceGraph) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var document struct {
		APIVersion string                     `yaml:"apiVersion"`
		Default    DefaultSettings            `yaml:"default"`
		Services   map[string]ServiceSettings `yaml:"services"`
	}
	if err := unmarshal(&document); err != nil {
		fmt.Println(document)
		return err
	}
	g.Services = make(map[string]Service)
	for name, details := range document.Services {
		g.Services[name] = Service{
			ServiceSettings: details,
			Name:            name,
		}
	}
	return nil
}

type Service struct {
	ServiceSettings
	Name string
}

// DefaultSettings describes the global defaults for the service graph.
type DefaultSettings struct {
	ServiceSettings
	RequestSettings
}

// ServiceSettings describes the configurable settings for a service.
type ServiceSettings struct {
	ComputeUsage string `yaml:"computeUsage"`
	MemoryUsage  string `yaml:"memoryUsage"`
	ErrorRate    string `yaml:"errorRate"`
}

// RequestSettings describes the configurable settings for service requests.
type RequestSettings struct {
	// PayloadSize is the number of bytes in request payloads.
	PayloadSize uint64 `yaml:"payloadSize"`
}
