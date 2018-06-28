package script

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestScript_UnmarshalJSON(t *testing.T) {
	DefaultRequestCommand = RequestCommand{}

	tests := []struct {
		input  []byte
		script Script
		err    error
	}{
		{
			[]byte(`[]`),
			Script{},
			nil,
		},
		{
			[]byte(`[{"sleep": "1s"}]`),
			Script{
				SleepCommand(1 * time.Second),
			},
			nil,
		},
		{
			[]byte(`[{"call": "A"}, {"sleep": "10ms"}]`),
			Script{
				RequestCommand{ServiceName: "A"},
				SleepCommand(10 * time.Millisecond),
			},
			nil,
		},
		{
			[]byte(`[[{"call": "A"}, {"call": "B"}], {"sleep": "10ms"}]`),
			Script{
				ConcurrentCommand{
					RequestCommand{ServiceName: "A"},
					RequestCommand{ServiceName: "B"},
				},
				SleepCommand(10 * time.Millisecond),
			},
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var script Script
			err := json.Unmarshal(test.input, &script)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if !reflect.DeepEqual(test.script, script) {
				t.Errorf("expected %v; actual %v", test.script, script)
			}
		})
	}
}
