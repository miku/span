package container

import (
	"reflect"
	"testing"
)

func TestStringMap(t *testing.T) {
	var (
		m = make(MapDefault)
		a = "a"
		b = "b"
	)
	m["a"] = a
	t.Run("existing key", func(t *testing.T) {
		if v := m.Lookup("a", b); v != a {
			t.Errorf("got %v, want %v", v, a)
		}
	})
	t.Run("missing key returns default", func(t *testing.T) {
		if v := m.Lookup("b", b); v != b {
			t.Errorf("got %v, want %s", v, b)
		}
	})
}

func TestStringSliceMap(t *testing.T) {
	var (
		m = make(MapSliceDefault)
		a = []string{"a"}
		b = []string{"b"}
	)
	m["a"] = a
	t.Run("existing key", func(t *testing.T) {
		if v := m.Lookup("a", b); !reflect.DeepEqual(v, a) {
			t.Errorf("got %v, want %v", v, a)
		}
	})
	t.Run("missing key returns default", func(t *testing.T) {
		if v := m.Lookup("b", b); !reflect.DeepEqual(v, b) {
			t.Errorf("got %v, want %v", v, b)
		}
	})
}
