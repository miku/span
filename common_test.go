package span

import (
	"fmt"
	"testing"
)

func TestGenFincID(t *testing.T) {
	var cases = []struct {
		sid      string
		rid      string
		expected string
	}{
		{"1", "123", "ai-1-MTIz"},
		{"1", "10.1234/5678", "ai-1-MTAuMTIzNC81Njc4"},
		{"", "", "ai--"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s/%s", c.sid, c.rid), func(t *testing.T) {
			result := GenFincID(c.sid, c.rid)
			if result != c.expected {
				t.Errorf("want %v, got %v", c.expected, result)
			}
		})
	}
}
