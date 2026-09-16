package index

import (
	"reflect"
	"testing"
)

func TestNormalizeISSN(t *testing.T) {
	var cases = []struct{ in, want string }{
		{"2421454x", "2421-454X"},
		{"1234-5678", "1234-5678"},
		{"weird", "WEIRD"},
	}
	for _, c := range cases {
		if got := normalizeISSN(c.in); got != c.want {
			t.Errorf("normalizeISSN(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPartition(t *testing.T) {
	var cases = []struct {
		in   []string
		size int
		want [][]string
	}{
		{[]string{"a", "b", "c", "d"}, 2, [][]string{{"a", "b"}, {"c", "d"}}},
		{[]string{"a", "b", "c"}, 2, [][]string{{"a", "b"}, {"c"}}},
		{[]string{"a", "b"}, 0, [][]string{{"a"}, {"b"}}},
		{[]string{}, 2, nil},
	}
	for _, c := range cases {
		if got := partition(c.in, c.size); !reflect.DeepEqual(got, c.want) {
			t.Errorf("partition(%v, %d) = %v, want %v", c.in, c.size, got, c.want)
		}
	}
}
