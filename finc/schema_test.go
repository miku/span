package finc

import (
	"reflect"
	"testing"
)

func TestAddInstitution(t *testing.T) {
	var tests = []struct {
		schema      SolrSchema
		institution string
		expected    SolrSchema
	}{
		{schema: SolrSchema{},
			institution: "X",
			expected:    SolrSchema{Institutions: []string{"X"}}},
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
