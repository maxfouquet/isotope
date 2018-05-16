package script

import (
	"time"

	"github.com/Tahler/service-grapher/pkg/graph/size"
)

// Command is the top level interface for commands.
type Command interface{}

// SleepCommand describes a command to pause for a duration.
type SleepCommand struct {
	Duration time.Duration
}

// RequestCommand describes a command to send an HTTP request to another
// service.
type RequestCommand struct {
	ServiceName string
	// Size is the number of bytes in the request body.
	Size size.ByteSize
}

// ConcurrentCommand describes a set of commands that should be executed
// simultaneously.
type ConcurrentCommand struct {
	Commands []Command
}
