package svc

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Tahler/isotope/convert/pkg/graph/svctype"
)

func TestService_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input []byte
		svc   Service
		err   error
	}{
		{
			[]byte(`{"name": "A"}`),
			Service{
				Name:        "A",
				Type:        svctype.ServiceHTTP,
				NumReplicas: 1,
			},
			nil,
		},
		{
			[]byte(`{}`),
			Service{Type: svctype.ServiceHTTP, NumReplicas: 1},
			ErrEmptyName,
		},
		{
			[]byte(`{"name": ""}`),
			Service{Type: svctype.ServiceHTTP, NumReplicas: 1},
			ErrEmptyName,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var svc Service
			err := json.Unmarshal(test.input, &svc)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if !reflect.DeepEqual(test.svc, svc) {
				t.Errorf("expected %v; actual %v", test.svc, svc)
			}
		})
	}
}
