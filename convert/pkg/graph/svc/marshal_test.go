package svc

import (
	"encoding/json"
	"testing"

	"github.com/maxfouquet/isotope/convert/pkg/graph/svctype"
)

func TestService_MarshalJSON(t *testing.T) {
	tests := []struct {
		input  Service
		output []byte
		err    error
	}{
		{
			Service{
				Name:        "a",
				Type:        svctype.ServiceHTTP,
				NumReplicas: 1,
			},
			[]byte(`{"name":"a","type":"http","numReplicas":1}`),
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			output, err := json.Marshal(test.input)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if string(test.output) != string(output) {
				t.Errorf("expected %s; actual %s", test.output, output)
			}
		})
	}
}
