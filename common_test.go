package span

import "testing"

func TestUnescapeTrim(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{in: "Hello", out: "Hello"},
		{in: "Hello &#x000F6;", out: "Hello ö"},
	}

	for _, tt := range tests {
		r := UnescapeTrim(tt.in)
		if r != tt.out {
			t.Errorf("UnescapeTrim(%s): got %s, want %s", tt.in, r, tt.out)
		}
	}
}
