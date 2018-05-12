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
	// Type describes what protocol the service supports (e.g. HTTP, gRPC).
	Type ServiceType
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

// ServiceType describes what protocol the service supports.
type ServiceType int

const (
	// UnknownService is the default, useless value for ServiceType.
	UnknownService ServiceType = iota
	// HTTPService indicates the service should run an HTTP server.
	HTTPService
	// GRPCService indicates the service should run a GRPC server.
	GRPCService
)
