package graph

import (
	"encoding/json"
	"sync"

	"github.com/Tahler/service-grapher/pkg/graph/pct"
	"github.com/Tahler/service-grapher/pkg/graph/script"
	"github.com/Tahler/service-grapher/pkg/graph/size"
	"github.com/Tahler/service-grapher/pkg/graph/svc"
	"github.com/Tahler/service-grapher/pkg/graph/svctype"
)

// UnmarshalJSON converts b into a valid ServiceGraph. See validate() for the
// details on what it means to be "valid".
func (g *ServiceGraph) UnmarshalJSON(b []byte) (err error) {
	metadata := serviceGraphJSONMetadata{Defaults: defaultDefaults}
	err = json.Unmarshal(b, &metadata)
	if err != nil {
		return
	}

	withGlobalDefaults(metadata.Defaults, func() {
		var unmarshallable unmarshallableServiceGraph
		err = json.Unmarshal(b, &unmarshallable)
		if err != nil {
			return
		}
		*g = ServiceGraph(unmarshallable)
	})

	err = validate(*g)
	if err != nil {
		return
	}

	return
}

// defaultDefaults is a stuttery but validly semantic name for the default
// values when parsing JSON defaults.
var (
	defaultDefaults = defaults{Type: svctype.ServiceHTTP}
	defaultMutex    sync.Mutex
)

type serviceGraphJSONMetadata struct {
	APIVersion string   `json:"apiVersion"`
	Defaults   defaults `json:"defaults"`
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
