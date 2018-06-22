package graph

import "github.com/Tahler/isotope/convert/pkg/graph/svc"

// ServiceGraph describes a set of services which mock a service-oriented
// architecture.
type ServiceGraph struct {
	Services []svc.Service `json:"services"`
}
