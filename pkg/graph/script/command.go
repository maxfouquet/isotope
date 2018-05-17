package script

import (
	"encoding/json"
	"fmt"
)

// Command is the top level interface for commands.
type Command interface{}

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
type unmarshallableCommand struct{ Command }

func (c *unmarshallableCommand) UnmarshalJSON(b []byte) error {
	isJSONArray := b[0] == '['
	if isJSONArray {
		var concurrentCommand ConcurrentCommand
		err := json.Unmarshal(b, &concurrentCommand)
		if err != nil {
			return err
		}
		c.Command = concurrentCommand
	} else {
		key, err := parseJSONCommandKey(b)
		if err != nil {
			return err
		}
		switch key {
		case "sleep":
			c.Command, err = parseSleepCommandFromJSONMap(b)
			if err != nil {
				return err
			}
		case "call":
			c.Command, err = parseRequestCommandFromJSONMap(b)
			if err != nil {
				return err
			}
		default:
			return UnknownCommandKeyError{key}
		}
	}
	return nil
}

func parseJSONCommandKey(b []byte) (s string, err error) {
	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	if len(m) > 1 {
		err = MultipleKeysInCommandMapError{m}
		return
	}
	// Should only loop once, setting s to the single command key in the map.
	for s = range m {
	}
	return
}

// b must contain a single key whose value is an unmarshallable SleepCommand.
func parseSleepCommandFromJSONMap(b []byte) (cmd SleepCommand, err error) {
	var m map[string]SleepCommand
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	for _, cmd = range m {
	}
	return
}

// b must contain a single key whose value is an unmarshallable RequestCommand.
func parseRequestCommandFromJSONMap(b []byte) (cmd RequestCommand, err error) {
	var m map[string]RequestCommand
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	for _, cmd = range m {
	}
	return
}

// MultipleKeysInCommandMapError is returned when there is more than one key in
// a command map.
type MultipleKeysInCommandMapError struct {
	CommandMap map[string]interface{}
}

func (e MultipleKeysInCommandMapError) Error() string {
	return fmt.Sprintf("multiple keys for command: %v", e.CommandMap)
}

// UnknownCommandKeyError is returned when a command's key (i.e. "sleep") does
// not match a known command.
type UnknownCommandKeyError struct {
	CommandKey string
}

func (e UnknownCommandKeyError) Error() string {
	return fmt.Sprintf("unknown command: %s", e.CommandKey)
}
