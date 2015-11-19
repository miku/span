package span

import "testing"

func TestISSNCompact(t *testing.T) {
	var cases = []struct {
		s      string
		result string
		err    error
	}{
		{"12341222", "1234-1222", nil},
	}

	for _, c := range cases {
		issn := ISSN(c.s)
		err := issn.Validate()
		if err != c.err {
			t.Errorf("want %v, got %v", c.err, err)
		}
		s := issn.String()
		if s != c.result {
			t.Errorf("want %v, got %v", c.result, s)
		}
	}
}
