package script

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSleepCommand_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input   []byte
		command SleepCommand
		err     error
	}{
		{
			[]byte(`"100ms"`),
			SleepCommand(100 * time.Millisecond),
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var command SleepCommand
			err := json.Unmarshal(test.input, &command)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if test.command != command {
				t.Errorf("expected %v; actual %v", test.command, command)
			}
		})
	}
}
