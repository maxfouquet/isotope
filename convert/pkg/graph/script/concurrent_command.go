package script

// ConcurrentCommand describes a set of commands that should be executed
// simultaneously.
type ConcurrentCommand []Command

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
