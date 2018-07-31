package script

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestConcurrentCommand_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input   []byte
		command ConcurrentCommand
		err     error
	}{
		{
			[]byte(`[]`),
			ConcurrentCommand{},
			nil,
		},
		{
			[]byte(`[{"sleep": "1s"}]`),
			ConcurrentCommand{
				SleepCommand(1 * time.Second),
			},
			nil,
		},
		{
			[]byte(`[{"call": "A"}, {"sleep": "10ms"}]`),
			ConcurrentCommand{
				RequestCommand{ServiceName: "A"},
				SleepCommand(10 * time.Millisecond),
			},
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var command ConcurrentCommand
			err := json.Unmarshal(test.input, &command)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if !reflect.DeepEqual(test.command, command) {
				t.Errorf("expected %v; actual %v", test.command, command)
			}
		})
	}
}
