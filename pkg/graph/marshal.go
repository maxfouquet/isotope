package graph

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var _ yaml.Marshaler = Service{}

// MarshalYAML implements the Marshaler interface and converts a Service to a
// marshallable interface{}.
func (svc Service) MarshalYAML() (marshallable interface{}, err error) {
	marshallable = serviceConfiguration{
		ServiceSettings: svc.ServiceSettings,
		Script:          svc.Script,
	}
	return
}

type serviceConfiguration struct {
	ServiceSettings `yaml:",inline"`
	Script          []Executable `yaml:"script,omitempty"`
}

// MarshalYAML implements the Marshaler interface and converts a
// ConcurrentCommand to a marshallable interface{}.
func (cmd ConcurrentCommand) MarshalYAML() (interface{}, error) {
	cmdList := make([]interface{}, 0, len(cmd.Commands))
	for _, subCmd := range cmd.Commands {
		if marshaler, ok := subCmd.(yaml.Marshaler); ok {
			marshallableSubCmd, err := marshaler.MarshalYAML()
			if err != nil {
				return nil, err
			}
			cmdList = append(cmdList, marshallableSubCmd)
		} else {
			return nil, fmt.Errorf(
				"sub command of type %T does not implement yaml.Marshaler",
				subCmd)
		}
	}
	return cmdList, nil
}

// MarshalYAML implements the Marshaler interface and converts a SleepCommand to
// a marshallable interface{}.
func (cmd SleepCommand) MarshalYAML() (marshallable interface{}, err error) {
	marshallable = map[string]interface{}{
		"sleep": cmd.Duration.String(),
	}
	return
}

// MarshalYAML implements the Marshaler interface and converts a RequestCommand
// to a marshallable interface{}.
func (cmd RequestCommand) MarshalYAML() (interface{}, error) {
	marshallable := make(map[string]requestConfiguration, 1)
	key := strings.ToLower(string(cmd.HTTPMethod))
	marshallable[key] = requestConfiguration{
		ServiceName:     cmd.ServiceName,
		RequestSettings: cmd.RequestSettings,
	}
	return marshallable, nil
}

type requestConfiguration struct {
	RequestSettings `yaml:",inline"`
	ServiceName     string `yaml:"service"`
}
