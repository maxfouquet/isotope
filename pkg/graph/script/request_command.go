package script

import (
	"encoding/json"

	"github.com/Tahler/service-grapher/pkg/graph/size"
)

// RequestCommand describes a command to send an HTTP request to another
// service.
type RequestCommand struct {
	ServiceName string `json:"service"`
	// Size is the number of bytes in the request body.
	Size size.ByteSize `json:"size"`
}

var (
	// DefaultRequestCommand is used by UnmarshalJSON to set defaults.
	DefaultRequestCommand RequestCommand
)

// UnmarshalJSON converts b to a RequestCommand. If b is a JSON string, it is
// set as c's ServiceName. If b is a JSON object, it's properties are mapped to
// c.
func (c *RequestCommand) UnmarshalJSON(b []byte) (err error) {
	*c = DefaultRequestCommand
	isJSONString := b[0] == '"'
	if isJSONString {
		var s string
		err = json.Unmarshal(b, &s)
		if err != nil {
			return
		}
		c.ServiceName = s
	} else {
		// Wrap the RequestCommand to dodge the custom UnmarshalJSON.
		unmarshallableRequestCommand := unmarshallableRequestCommand(*c)
		err = json.Unmarshal(b, &unmarshallableRequestCommand)
		if err != nil {
			return
		}
		*c = RequestCommand(unmarshallableRequestCommand)
	}
	return
}

type unmarshallableRequestCommand RequestCommand
