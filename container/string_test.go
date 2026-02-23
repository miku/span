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

func TestStringSet(t *testing.T) {
	ss := NewStringSet()
	t.Run("empty set", func(t *testing.T) {
		if ss.Size() != 0 {
			t.Errorf("empty set should be 0, got %d", ss.Size())
		}
	})
	ss.Add("a")
	t.Run("add one element", func(t *testing.T) {
		if ss.Size() != 1 {
			t.Errorf("expected size 1, got %d", ss.Size())
		}
		if !ss.Contains("a") {
			t.Errorf("expected a in set")
		}
	})
	ss.Add("a")
	ss.Add("c", "b")
	t.Run("sorted values", func(t *testing.T) {
		if exp := []string{"a", "b", "c"}; !reflect.DeepEqual(ss.SortedValues(), exp) {
			t.Errorf("got %v, want %v", ss.SortedValues(), exp)
		}
	})
	st := NewStringSet("a", "x")
	t.Run("intersection", func(t *testing.T) {
		if st.Intersection(ss).Size() != 1 {
			t.Errorf("expected overlap of 1, got %d", st.Intersection(ss).Size())
		}
	})
	t.Run("difference", func(t *testing.T) {
		if st.Difference(ss).Size() != 1 {
			t.Errorf("expected difference of 1, got %d", st.Difference(ss).Size())
		}
	})
}

func BenchmarkStringSetAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ss := NewStringSet()
		ss.Add("a", "b", "c", "d", "e")
	}
}

func BenchmarkStringSetContains(b *testing.B) {
	ss := NewStringSet("a", "b", "c", "d", "e")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ss.Contains("c")
		_ = ss.Contains("z")
	}
}

func BenchmarkStringSetIntersection(b *testing.B) {
	a := NewStringSet("a", "b", "c", "d", "e")
	bset := NewStringSet("c", "d", "e", "f", "g")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Intersection(bset)
	}
}
