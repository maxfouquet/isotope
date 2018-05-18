package script

import (
	"encoding/json"
)

// Script is a list of commands to be sequentially executed.
type Script []Command

// MarshalJSON encodes the Script as a JSON array of JSON objects.
func (s Script) MarshalJSON() ([]byte, error) {
	marshallableCmds, err := commandsToMarshallable(s)
	if err != nil {
		return nil, err
	}
	return json.Marshal(marshallableCmds)
}

// UnmarshalJSON converts b to a Script. b must be a JSON array of Commands.
func (s *Script) UnmarshalJSON(b []byte) (err error) {
	cmds, err := parseJSONCommands(b)
	if err != nil {
		return
	}
	*s = Script(cmds)
	return
}
