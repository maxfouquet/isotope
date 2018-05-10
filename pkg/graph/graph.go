package graph

// ServiceGraph describes a set of services which mock a service-oriented
// architecture.
type ServiceGraph struct {
	Services map[string]Service
}

// Service describes a service in the service graph.
type Service struct {
	ServiceSettings
	Name string
	// Script is sequentially called each time the service is called.
	Script []Command
}

// ServiceSettings describes the configurable settings for a service.
type ServiceSettings struct {
	// ComputeUsage is the percentage of CPU power that should be used during
	// script execution.
	ComputeUsage float64
	// MemoryUsage is the percentage of memory that should be used during script
	// execution.
	MemoryUsage float64
	// ErrorRate is the percentage chance between 0 and 1 that this service
	// should respond with a 500 server error rather than200 OK.
	ErrorRate float64
	// ResponseSize is the number of bytes in the response body.
	ResponseSize int64
}

// RequestSettings describes the configurable settings for service requests.
type RequestSettings struct {
	// Size is the number of bytes in the request body.
	Size int64
}
