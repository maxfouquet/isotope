package script

import (
	"encoding/json"
	"reflect"
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
			[]byte(`{"sleep": "100ms"}`),
			SleepCommand{100 * time.Millisecond},
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

func TestRequestCommand_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input   []byte
		command RequestCommand
		err     error
	}{
		{
			[]byte(`{"call": "A"}`),
			RequestCommand{ServiceName: "A"},
			nil,
		},
		{
			[]byte(`{"call": {"service": "A"}}`),
			RequestCommand{ServiceName: "A"},
			nil,
		},
		{
			[]byte(`{"call": {"service": "a", "size": 128}}`),
			RequestCommand{ServiceName: "a", Size: 128},
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var command RequestCommand
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
				SleepCommand{1 * time.Second},
			},
			nil,
		},
		{
			[]byte(`[{"call": "A"}, {"sleep": "10ms"}]`),
			ConcurrentCommand{
				RequestCommand{ServiceName: "A"},
				SleepCommand{10 * time.Millisecond},
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
