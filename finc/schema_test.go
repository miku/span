package finc

import (
	"reflect"
	"testing"
)

func TestAddInstitution(t *testing.T) {
	var tests = []struct {
		schema      Schema
		institution string
		expected    Schema
	}{
		{schema: Schema{},
			institution: "X",
			expected:    Schema{Institutions: []string{"X"}}},
	}

	for _, tt := range tests {
		tt.schema.AddInstitution(tt.institution)
		if !reflect.DeepEqual(tt.schema, tt.expected) {
			t.Errorf("got: %s, want: %s", tt.schema, tt.expected)
		}
		tt.schema.AddInstitution(tt.institution)
		if !reflect.DeepEqual(tt.schema, tt.expected) {
			t.Errorf("got: %s, want: %s", tt.schema, tt.expected)
		}
	}
}
