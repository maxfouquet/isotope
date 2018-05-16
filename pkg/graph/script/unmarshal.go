package script

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/Tahler/service-grapher/pkg/graph/size"
)

// UnmarshalJSON converts b to a Script. b must be a JSON array of Commands.
func (s *Script) UnmarshalJSON(b []byte) (err error) {
	cmds, err := parseJSONCommands(b)
	if err != nil {
		return
	}
	*s = Script(cmds)
	return
}

func parseJSONCommands(b []byte) ([]Command, error) {
	var wrappedCmds []unmarshallableCommand
	err := json.Unmarshal(b, &wrappedCmds)
	if err != nil {
		return nil, err
	}

	cmds := make([]Command, 0, len(wrappedCmds))
	for _, wrappedCmd := range wrappedCmds {
		cmd := wrappedCmd.Command
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

// unmarshallableCommand wraps a Command so that it may act as a receiver.
type unmarshallableCommand struct {
	Command
}

func (c *unmarshallableCommand) UnmarshalJSON(b []byte) (err error) {
	isJSONArray := b[0] == '['
	if isJSONArray {
		var concurrentCommand ConcurrentCommand
		err = json.Unmarshal(b, &concurrentCommand)
		if err != nil {
			return
		}
		c.Command = concurrentCommand
	} else {
		var m map[string]interface{}
		err = json.Unmarshal(b, &m)
		if err != nil {
			return
		}
		numKeys := len(m)
		keys := make([]string, 0, numKeys)
		for key := range m {
			keys = append(keys, key)
		}
		if numKeys > 1 {
			err = InvalidCommandKeysError{keys}
			return
		}
		key := keys[0]
		switch key {
		case "sleep":
			var sleepCommand SleepCommand
			err = json.Unmarshal(b, &sleepCommand)
			if err != nil {
				return
			}
			c.Command = sleepCommand
		case "call":
			var requestCommand RequestCommand
			err = json.Unmarshal(b, &requestCommand)
			if err != nil {
				return
			}
			c.Command = requestCommand
		default:
			err = InvalidCommandKeysError{keys}
		}
	}
	return
}

// UnmarshalJSON converts a JSON object to a SleepCommand.
func (c *SleepCommand) UnmarshalJSON(b []byte) (err error) {
	var m map[string]string
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	durationStr, ok := m["sleep"]
	if !ok {
		err = MissingKeyError{"sleep", reflect.TypeOf(c)}
	}
	c.Duration, err = time.ParseDuration(durationStr)
	if err != nil {
		return
	}
	return
}

// UnmarshalJSON converts b to a SleepCommand. If b is a JSON string, it is set
// as c's ServiceName. If b is a JSON object, it's properties are mapped to c.
func (c *RequestCommand) UnmarshalJSON(b []byte) (err error) {
	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	impl, ok := m["call"]
	if !ok {
		err = MissingKeyError{"call", reflect.TypeOf(c)}
		return
	}

	if s, ok := impl.(string); ok {
		c.ServiceName = s
	} else {
		var m map[string]struct {
			ServiceName string        `json:"service"`
			Size        size.ByteSize `json:"size"`
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			return
		}
		settings, _ := m["call"]
		c.ServiceName = settings.ServiceName
		c.Size = settings.Size
	}
	return
}

// UnmarshalJSON converts b to a ConcurrentCommand. b must be a JSON array of
// commands.
func (c *ConcurrentCommand) UnmarshalJSON(b []byte) (err error) {
	cmds, err := parseJSONCommands(b)
	if err != nil {
		return
	}
	*c = ConcurrentCommand(cmds)
	return
}

// MissingKeyError is returned when unmarshalling fails because a key is not
// present in an intermediate map.
type MissingKeyError struct {
	MissingKey             string
	AttemptedUnmarshalType reflect.Type
}

func (e MissingKeyError) Error() string {
	return fmt.Sprintf(
		"failed to unmarshal to %T: missing key: \"%s\"",
		e.AttemptedUnmarshalType, e.MissingKey)
}

// InvalidCommandKeysError is returned when the keys of a map representing a
// command are invalid.
type InvalidCommandKeysError struct {
	Keys []string
}

func (e InvalidCommandKeysError) Error() string {
	return fmt.Sprintf("invalid keys for command: %v", e.Keys)
}
