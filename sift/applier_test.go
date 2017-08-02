package sift

import (
	"encoding/json"
	"testing"
)

type testC struct{}

func (testC) Collections() []string {
	return []string{"a"}
}

func TestCollection(t *testing.T) {
	var cases = []struct {
		config string
		value  interface{}
		result bool
	}{
		{
			config: `{"collections": ["a", "b"]}`,
			value:  3,
			result: false,
		},
		{
			config: `{"collections": ["a", "b"]}`,
			value:  testC{},
			result: true,
		},
		{
			config: `{"collections": ["c"]}`,
			value:  testC{},
			result: false,
		},
		{
			config: `{"collections": ["c"], "fallback": true}`,
			value:  testC{},
			result: true,
		},
	}

	for _, c := range cases {
		var applier Collection
		if err := json.Unmarshal([]byte(c.config), &applier); err != nil {
			t.Error(err)
		}
		if got, want := applier.Apply(c.value), c.result; got != want {
			t.Errorf("Apply: got %v, want %v", got, want)
		}
	}
}
