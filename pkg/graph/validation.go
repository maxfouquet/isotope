package graph

import (
	"errors"
	"fmt"

	"github.com/Tahler/service-grapher/pkg/graph/script"
)

// validate returns nil if g is valid.
// g is valid if a ServiceGraph:
// - Each of its services only makes requests to other defined services.
// - ConcurrentCommands do not contain other ConcurrentCommands.
func validate(g ServiceGraph) (err error) {
	svcNames := map[string]bool{}
	for _, svc := range g.Services {
		svcNames[svc.Name] = true
	}
	for _, svc := range g.Services {
		err = validateCommands(svc.Script, svcNames)
		if err != nil {
			return
		}
	}
	return
}

func validateCommands(cmds []script.Command, svcNames map[string]bool) error {
	for _, cmd := range cmds {
		switch cmd := cmd.(type) {
		case script.RequestCommand:
			if !svcNames[cmd.ServiceName] {
				return ErrRequestToUndefinedService{cmd.ServiceName}
			}
		case script.ConcurrentCommand:
			err := validateCommands(cmd, svcNames)
			if err != nil {
				return err
			}
			if containsConcurrentCommand([]script.Command(cmd)) {
				return ErrNestedConcurrentCommand
			}
		}
	}
	return nil
}

func containsConcurrentCommand(cmds []script.Command) bool {
	for _, cmd := range cmds {
		if _, ok := cmd.(script.ConcurrentCommand); ok {
			return true
		}
	}
	return false
}

// ErrRequestToUndefinedService is returned when a RequestCommand has a
// ServiceName that is not the name of a defined service.
type ErrRequestToUndefinedService struct {
	ServiceName string
}

func (e ErrRequestToUndefinedService) Error() string {
	return fmt.Sprintf(`cannot call undefined service "%s"`, e.ServiceName)
}

// ErrNestedConcurrentCommand is returned when a ConcurrentCommand contains
// a ConcurrentCommand.
var ErrNestedConcurrentCommand = errors.New(
	"concurrent commands may not be nested")
