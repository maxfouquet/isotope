package graph

import (
	"encoding/json"
	"sync"

	"github.com/Tahler/isotope/convert/pkg/graph/pct"
	"github.com/Tahler/isotope/convert/pkg/graph/script"
	"github.com/Tahler/isotope/convert/pkg/graph/size"
	"github.com/Tahler/isotope/convert/pkg/graph/svc"
	"github.com/Tahler/isotope/convert/pkg/graph/svctype"
)

// UnmarshalJSON converts b into a valid ServiceGraph. See validate() for the
// details on what it means to be "valid".
func (g *ServiceGraph) UnmarshalJSON(b []byte) (err error) {
	metadata := serviceGraphJSONMetadata{Defaults: defaultDefaults}
	err = json.Unmarshal(b, &metadata)
	if err != nil {
		return
	}

	*g, err = parseJSONServiceGraphWithDefaults(b, metadata.Defaults)
	if err != nil {
		return
	}

	err = validate(*g)
	if err != nil {
		return
	}

	return
}

func parseJSONServiceGraphWithDefaults(
	b []byte, defaults defaults) (sg ServiceGraph, err error) {
	withGlobalDefaults(defaults, func() {
		var unmarshallable unmarshallableServiceGraph
		innerErr := json.Unmarshal(b, &unmarshallable)
		if innerErr == nil {
			sg = ServiceGraph(unmarshallable)
		} else {
			err = innerErr
		}
	})
	return
}

// defaultDefaults is a stuttery but validly semantic name for the default
// values when parsing JSON defaults.
var (
	defaultDefaults = defaults{Type: svctype.ServiceHTTP}
	defaultMutex    sync.Mutex
)

type serviceGraphJSONMetadata struct {
	Defaults defaults `json:"defaults"`
}

type defaults struct {
	Type         svctype.ServiceType `json:"type"`
	ErrorRate    pct.Percentage      `json:"errorRate"`
	ResponseSize size.ByteSize       `json:"responseSize"`
	Script       script.Script       `json:"script"`
	RequestSize  size.ByteSize       `json:"requestSize"`
}

func withGlobalDefaults(defaults defaults, f func()) {
	defaultMutex.Lock()

	origDefaultService := svc.DefaultService
	svc.DefaultService = svc.Service{
		Type:         defaults.Type,
		ErrorRate:    defaults.ErrorRate,
		ResponseSize: defaults.ResponseSize,
		Script:       defaults.Script,
	}

	origDefaultRequestCommand := script.DefaultRequestCommand
	script.DefaultRequestCommand = script.RequestCommand{
		Size: defaults.RequestSize,
	}

	f()

	svc.DefaultService = origDefaultService
	script.DefaultRequestCommand = origDefaultRequestCommand

	defaultMutex.Unlock()
}

type unmarshallableServiceGraph ServiceGraph
