package graph

type ServiceGraph struct {
	Services map[string]Service
}

type Service struct {
	ServiceSettings
	Name string
	// Script is sequentially called each time the service is called.
	Script []Executable
}

// ServiceSettings describes the configurable settings for a service.
type ServiceSettings struct {
	ComputeUsage float64
	// MemoryUsage is the percentage of memory that should be used during script
	// execution.
	MemoryUsage float64
	// ErrorRate is the percentage chance between 0 and 1 that this service
	// should respond with a 500 server error rather than200 OK.
	ErrorRate float64
}

// RequestSettings describes the configurable settings for service requests.
type RequestSettings struct {
	// PayloadSize is the number of bytes in request payloads.
	PayloadSize int64
}
