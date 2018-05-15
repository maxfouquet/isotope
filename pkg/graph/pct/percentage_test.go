package pct

import "testing"

func TestFromFloat64(t *testing.T) {
	tests := []struct {
		input      float64
		percentage Percentage
		err        error
	}{
		{0.0, 0.0, nil},
		{0.1, 0.1, nil},
		{1.0, 1.0, nil},
		{1.1, 0, OutOfRangeError{1.1}},
		{100, 0, OutOfRangeError{100}},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			percentage, err := FromFloat64(test.input)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if test.percentage != percentage {
				t.Errorf("expected %v; actual %v", test.percentage, percentage)
			}
		})
	}
}

func TestFromString(t *testing.T) {
	tests := []struct {
		input      string
		percentage Percentage
		err        error
	}{
		{"0%", 0.0, nil},
		{"10%", 0.1, nil},
		{"100%", 1.0, nil},
		{"110%", 0, OutOfRangeError{1.1}},
		{"100", 0, InvalidPercentageStringError{"100"}},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			percentage, err := FromString(test.input)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if test.percentage != percentage {
				t.Errorf("expected %v; actual %v", test.percentage, percentage)
			}
		})
	}
}
