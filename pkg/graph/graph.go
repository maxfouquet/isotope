package graph

type ServiceGraph struct {
	Services map[string]Service
}

type Service struct {
	ServiceSettings
	Name string
}

// ServiceSettings describes the configurable settings for a service.
type ServiceSettings struct {
	ComputeUsage float64 `yaml:"computeUsage"`
	MemoryUsage  float64 `yaml:"memoryUsage"`
	ErrorRate    float64 `yaml:"errorRate"`
}

// RequestSettings describes the configurable settings for service requests.
type RequestSettings struct {
	// PayloadSize is the number of bytes in request payloads.
	PayloadSize int64 `yaml:"payloadSize"`
}
