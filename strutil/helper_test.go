package strutil

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
		t.Run(tt.in, func(t *testing.T) {
			r := UnescapeTrim(tt.in)
			if r != tt.out {
				t.Errorf("UnescapeTrim(%s): got %s, want %s", tt.in, r, tt.out)
			}
		})
	}
}

func BenchmarkUnescapeTrim(b *testing.B) {
	inputs := []string{
		"Hello",
		"Hello &#x000F6;",
		"complex &#x000E4; string &#x000FC; with &#x000F6; entities",
	}
	for _, in := range inputs {
		b.Run(in, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = UnescapeTrim(in)
			}
		})
	}
}
