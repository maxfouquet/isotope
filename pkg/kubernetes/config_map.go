package kubernetes

import (
	"fmt"
	"strings"

	"github.com/Tahler/service-grapher/pkg/graph"
)

type configMapService struct {
	ComputeUsage float64       `json:"computeUsage"`
	MemoryUsage  float64       `json:"memoryUsage"`
	ErrorRate    float64       `json:"errorRate"`
	Script       []interface{} `json:"script,omitempty"`
}

// MarshalYAML implements the Marshaler interface and converts a Service to a
// marshallable interface{}.
func serviceToMarshallable(
	service graph.Service) (marshallable interface{}, err error) {
	script, err := scriptToMarshallable(service.Script)
	if err != nil {
		return
	}
	marshallable = configMapService{
		ComputeUsage: service.ComputeUsage,
		MemoryUsage:  service.MemoryUsage,
		ErrorRate:    service.ErrorRate,
		Script:       script,
	}
	return
}

func scriptToMarshallable(script []graph.Command) ([]interface{}, error) {
	marshallableSlice := make([]interface{}, 0, len(script))
	for _, exe := range script {
		marshallable, err := executableToMarshallable(exe)
		if err != nil {
			return nil, err
		}
		marshallableSlice = append(marshallableSlice, marshallable)
	}
	return marshallableSlice, nil
}

func executableToMarshallable(
	exe graph.Command) (marshallable interface{}, err error) {
	switch cmd := exe.(type) {
	case graph.SleepCommand:
		marshallable = sleepCommandToMarshallable(cmd)
	case graph.RequestCommand:
		marshallable = requestCommandToMarshallable(cmd)
	case graph.ConcurrentCommand:
		marshallable, err = concurrentCommandToMarshallable(cmd)
	default:
		err = fmt.Errorf("unexpected type %T", cmd)
	}
	return
}

func sleepCommandToMarshallable(cmd graph.SleepCommand) interface{} {
	return map[string]string{"sleep": cmd.Duration.String()}
}

func requestCommandToMarshallable(cmd graph.RequestCommand) interface{} {
	marshallable := make(map[string]configMapRequestCommand, 1)
	key := strings.ToLower(string(cmd.HTTPMethod))
	marshallable[key] = configMapRequestCommand{
		Service:     cmd.ServiceName,
		PayloadSize: cmd.PayloadSize,
	}
	return marshallable
}

type configMapRequestCommand struct {
	Service     string `json:"service"`
	PayloadSize int64  `json:"payloadSize"`
}

func concurrentCommandToMarshallable(
	cmd graph.ConcurrentCommand) (interface{}, error) {
	cmdList := make([]interface{}, 0, len(cmd.Commands))
	for _, subCmd := range cmd.Commands {
		marshallable, err := executableToMarshallable(subCmd)
		if err != nil {
			return nil, err
		}
		cmdList = append(cmdList, marshallable)
	}
	return cmdList, nil
}
