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
	if v := m.Lookup("a", b); v != a {
		t.Errorf("got %v, want %v", v, a)
	}
	if v := m.Lookup("b", b); v != b {
		t.Errorf("got %v, want %s", v, b)
	}
}

func TestStringSliceMap(t *testing.T) {
	var (
		m = make(MapSliceDefault)
		a = []string{"a"}
		b = []string{"b"}
	)
	m["a"] = a
	if v := m.Lookup("a", b); !reflect.DeepEqual(v, a) {
		t.Errorf("got %v, want %v", v, a)
	}
	if v := m.Lookup("b", b); !reflect.DeepEqual(v, b) {
		t.Errorf("got %v, want %v", v, b)
	}
}

func TestStringSet(t *testing.T) {
	ss := NewStringSet()
	if ss.Size() != 0 {
		t.Errorf("empty set should be 0")
	}
	ss.Add("a")
	if ss.Size() != 1 {
		t.Errorf("expected size 1")
	}
	if !ss.Contains("a") {
		t.Errorf("expected a in set")
	}
	ss.Add("a")
	ss.Add("c", "b")
	if exp := []string{"a", "b", "c"}; !reflect.DeepEqual(ss.SortedValues(), exp) {
		t.Errorf("expected: %s", exp)
	}
	st := NewStringSet("a", "x")
	if st.Intersection(ss).Size() != 1 {
		t.Errorf("expected overlap of 1")
	}
	if st.Difference(ss).Size() != 1 {
		t.Errorf("expected difference of 1")
	}
}
