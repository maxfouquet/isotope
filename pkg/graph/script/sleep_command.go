package script

import (
	"encoding/json"
	"time"
)

// SleepCommand describes a command to pause for a duration.
type SleepCommand time.Duration

// UnmarshalJSON converts a JSON object to a SleepCommand.
func (c *SleepCommand) UnmarshalJSON(b []byte) (err error) {
	var durationStr string
	err = json.Unmarshal(b, &durationStr)
	if err != nil {
		return
	}
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return
	}
	*c = SleepCommand(duration)
	return
}
