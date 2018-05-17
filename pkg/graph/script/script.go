package script

// Script is a list of commands to be sequentially executed.
type Script []Command

// UnmarshalJSON converts b to a Script. b must be a JSON array of Commands.
func (s *Script) UnmarshalJSON(b []byte) (err error) {
	cmds, err := parseJSONCommands(b)
	if err != nil {
		return
	}
	*s = Script(cmds)
	return
}
