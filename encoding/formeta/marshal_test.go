package formeta

import "testing"

func TestEncoding(t *testing.T) {
	var cases = []struct {
		in  interface{}
		out string
		err error
	}{
		{in: "", out: "", err: nil},
		{in: "x", out: "", err: ErrValueNotAllowed},
		{in: struct{ A string }{A: "B"}, out: ` { A: 'B',  } `, err: nil},
		{in: struct{ A string }{A: "B 'A"}, out: ` { A: 'B \'A',  } `, err: nil},
		{
			in: struct{ A string }{A: `B
A`}, out: ` { A: 'B\nA',  } `, err: nil,
		},
		{
			in: struct{ A string }{A: `B\ A`}, out: ` { A: 'B\\ A',  } `, err: nil,
		},
		{
			in: struct{ A string }{A: `B\
'A \`}, out: ` { A: 'B\\\n\'A \\',  } `, err: nil,
		},
		{in: struct{ A []string }{A: []string{"B", "C"}}, out: ` { A: 'B', A: 'C',  } `, err: nil},
		{in: struct{ A int }{A: 1}, out: ` { A: 1,  } `, err: nil},
		{in: struct{ A int64 }{A: 1}, out: ` { A: 1,  } `, err: nil},
	}

	for _, c := range cases {
		b, err := Marshal(c.in)
		if err != c.err {
			t.Errorf("Marshal got %v, want %v", err, c.err)
		}
		if string(b) != c.out {
			t.Errorf("Marshal got %v, want %v", string(b), c.out)
		}
	}
}
