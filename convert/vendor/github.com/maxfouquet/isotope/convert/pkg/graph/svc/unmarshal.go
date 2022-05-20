package svc

import (
	"encoding/json"
	"errors"

	"github.com/maxfouquet/isotope/convert/pkg/graph/svctype"
)

var (
	// DefaultService is used by UnmarshalJSON and describes the default settings.
	DefaultService = Service{Type: svctype.ServiceHTTP, NumReplicas: 1}
)

// UnmarshalJSON converts b to a Service, applying the default values from
// DefaultService.
func (svc *Service) UnmarshalJSON(b []byte) (err error) {
	unmarshallable := unmarshallableService(DefaultService)
	err = json.Unmarshal(b, &unmarshallable)
	if err != nil {
		return
	}
	*svc = Service(unmarshallable)
	if svc.Name == "" {
		err = ErrEmptyName
		return
	}
	return
}

type unmarshallableService Service

// ErrEmptyName is returned when attempting to parse JSON without an empty name
// field.
var ErrEmptyName = errors.New("services must have a name")
