package script

import (
	"encoding/json"
	"testing"
)

func TestRequestCommand_UnmarshalJSON(t *testing.T) {
	DefaultRequestCommand = RequestCommand{}

	tests := []struct {
		input   []byte
		command RequestCommand
		err     error
	}{
		{
			[]byte(`"A"`),
			RequestCommand{ServiceName: "A"},
			nil,
		},
		{
			[]byte(`{"service": "A"}`),
			RequestCommand{ServiceName: "A"},
			nil,
		},
		{
			[]byte(`{"service": "a", "size": 128}`),
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

func TestRequestCommand_UnmarshalJSON_Default(t *testing.T) {
	DefaultRequestCommand = RequestCommand{Size: 512}

	tests := []struct {
		input   []byte
		command RequestCommand
		err     error
	}{
		{
			[]byte(`"A"`),
			RequestCommand{ServiceName: "A", Size: 512},
			nil,
		},
		{
			[]byte(`{"service": "A"}`),
			RequestCommand{ServiceName: "A", Size: 512},
			nil,
		},
		{
			[]byte(`{"service": "a", "size": 128}`),
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
