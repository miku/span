package sift

import (
	"encoding/json"
	"testing"
)

type testC struct{}

func (testC) Collections() []string {
	return []string{"a"}
}

type testS struct{}

func (testS) SerialNumbers() []string {
	return []string{"1234-5678"}
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
		ap := new(Collection)
		if err := json.Unmarshal([]byte(c.config), ap); err != nil {
			t.Error(err)
		}
		if got, want := ap.Apply(c.value), c.result; got != want {
			t.Errorf("Apply: got %v, want %v", got, want)
		}
	}
}

func TestSerialNumber(t *testing.T) {
	var cases = []struct {
		config string
		value  interface{}
		result bool
	}{
		{
			config: `{"issn": {"values": ["1"]}}`,
			value:  3,
			result: false,
		},
		{
			config: `{"issn": {"values": ["1"], "fallback": true}}`,
			value:  3,
			result: true,
		},
		{
			config: `{"issn": {"values": ["1234-5678"]}}`,
			value:  testS{},
			result: true,
		},
	}

	for _, c := range cases {
		ap := new(SerialNumber)
		if err := json.Unmarshal([]byte(c.config), ap); err != nil {
			t.Error(err)
		}
		if got, want := ap.Apply(c.value), c.result; got != want {
			t.Errorf("Apply: got %v, want %v", got, want)
		}
	}
}
