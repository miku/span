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

type testA struct{}

func (testA) Collections() []string {
	return []string{"a"}
}

func (testA) SerialNumbers() []string {
	return []string{"1234"}
}

func TestAny(t *testing.T) {
	var applier = AnyFilter{}
	if applier.Apply(1) != true {
		t.Errorf("Apply, got false, want true")
	}
	if applier.Apply(false) != true {
		t.Errorf("Apply, got false, want true")
	}
}

func TestCollection(t *testing.T) {
	var cases = []struct {
		config string
		value  interface{}
		result bool
	}{
		{
			config: `{"collection": {"values": ["a", "b"]}}`,
			value:  3,
			result: false,
		},
		{
			config: `{"collection": {"values": ["a", "b"]}}`,
			value:  testC{},
			result: true,
		},
		{
			config: `{"collection": {"values": ["c"]}}`,
			value:  testC{},
			result: false,
		},
		{
			config: `{"collection": {"values": ["c"], "fallback": true}}`,
			value:  testC{},
			result: false,
		},
	}

	for i, c := range cases {
		ap := new(CollectionsFilter)
		if err := json.Unmarshal([]byte(c.config), ap); err != nil {
			t.Error(err)
		}
		if got, want := ap.Apply(c.value), c.result; got != want {
			t.Errorf("#%d, %T.Apply(%v, %T): got %v, want %v", i, ap, c.value, c.value, got, want)
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
		ap := new(SerialNumbersFilter)
		if err := json.Unmarshal([]byte(c.config), ap); err != nil {
			t.Error(err)
		}
		if got, want := ap.Apply(c.value), c.result; got != want {
			t.Errorf("Apply: got %v, want %v", got, want)
		}
	}
}

func TestAndFilter(t *testing.T) {
	var cases = []struct {
		config string
		value  interface{}
		result bool
	}{
		{
			config: `{"and":[{"collection":{"values":["a"]}},{"issn":{"values":["1234"]}}]}`,
			value:  3,
			result: false,
		},
		{
			config: `{"and":[{"collection":{"values":["a"]}},{"issn":{"values":["1234"]}}]}`,
			value:  testA{},
			result: true,
		},
		{
			config: `{"and":[{"collection":{"values":["b"]}},{"issn":{"values":["1234"]}}]}`,
			value:  testA{},
			result: false,
		},
		{
			config: `{"and":[{"collection":{"values":["b", "a"]}},{"issn":{"values":["1234", "1234"]}}]}`,
			value:  testA{},
			result: true,
		},
	}

	for i, c := range cases {
		ap := new(AndFilter)
		if err := json.Unmarshal([]byte(c.config), ap); err != nil {
			t.Error(err)
		}
		if got, want := ap.Apply(c.value), c.result; got != want {
			t.Errorf("#%d, %T.Apply(%v, %T): got %v, want %v (filter was: %s)", i, ap, c.value, c.value, got, want, ap)
		}
	}
}

func TestOrFilter(t *testing.T) {
	var cases = []struct {
		config string
		value  interface{}
		result bool
	}{
		{
			config: `{"or":[{"collection":{"values":["a"]}},{"issn":{"values":["1234"]}}]}`,
			value:  3,
			result: false,
		},
		{
			config: `{"or":[{"collection":{"values":["b"]}},{"issn":{"values":["3333"]}}]}`,
			value:  testA{},
			result: false,
		},
		{
			config: `{"or":[{"collection":{"values":["a"]}},{"issn":{"values":["1234"]}}]}`,
			value:  testA{},
			result: true,
		},
		{
			config: `{"or":[{"collection":{"values":["b"]}},{"issn":{"values":["1234"]}}]}`,
			value:  testA{},
			result: true,
		},
		{
			config: `{"or":[{"collection":{"values":["b", "a"]}},{"issn":{"values":["1234", "1234"]}}]}`,
			value:  testA{},
			result: true,
		},
	}

	for i, c := range cases {
		ap := new(OrFilter)
		if err := json.Unmarshal([]byte(c.config), ap); err != nil {
			t.Error(err)
		}
		if got, want := ap.Apply(c.value), c.result; got != want {
			t.Errorf("#%d, %T.Apply(%v, %T): got %v, want %v (filter was: %s)", i, ap, c.value, c.value, got, want, ap)
		}
	}
}

func TestNotFilter(t *testing.T) {
	var cases = []struct {
		config string
		value  interface{}
		result bool
	}{
		{
			config: `{"not": {"and":[{"collection":{"values":["a"]}},{"issn":{"values":["1234"]}}]}}`,
			value:  3,
			result: true,
		},
		{
			config: `{"not": {"and":[{"collection":{"values":["a"]}},{"issn":{"values":["1234"]}}]}}`,
			value:  testA{},
			result: false,
		},
		{
			config: `{"not": {"and":[{"collection":{"values":["b"]}},{"issn":{"values":["1234"]}}]}}`,
			value:  testA{},
			result: true,
		},
		{
			config: `{"not": {"not": {"and":[{"collection":{"values":["b"]}},{"issn":{"values":["1234"]}}]}}}`,
			value:  testA{},
			result: false,
		},
		{
			config: `{"not": {"and":[{"collection":{"values":["b", "a"]}},{"issn":{"values":["1234", "1234"]}}]}}`,
			value:  testA{},
			result: false,
		},
	}

	for i, c := range cases {
		ap := new(NotFilter)
		if err := json.Unmarshal([]byte(c.config), ap); err != nil {
			t.Error(err)
		}
		if got, want := ap.Apply(c.value), c.result; got != want {
			t.Errorf("#%d, %T.Apply(%v, %T): got %v, want %v (filter was: %s)", i, ap, c.value, c.value, got, want, ap)
		}
	}
}
