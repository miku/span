package formeta

import (
	"fmt"
	"testing"
	"time"

	"github.com/segmentio/encoding/json"
)

func TestEncoding(t *testing.T) {
	var cases = []struct {
		about string
		in    any
		out   string
		err   error
	}{
		{"empty string", "", "", nil},
		{"single char rejected", "x", "", ErrValueNotAllowed},
		{"simple struct", struct{ A string }{A: "B"}, `{ A: 'B',  }`, nil},
		{"single quote in value", struct{ A string }{A: "B 'A"}, `{ A: 'B \'A',  }`, nil},
		{"string slice", struct{ A []string }{A: []string{"B", "C"}}, `{ A: 'B', A: 'C',  }`, nil},
		{"int field", struct{ A int }{A: 1}, `{ A: 1,  }`, nil},
		{"int64 field", struct{ A int64 }{A: 1}, `{ A: 1,  }`, nil},
		{"newline in value", struct{ A string }{A: "B\nA"}, `{ A: 'B\nA',  }`, nil},
		{"backslash in value", struct{ A string }{A: `B\ A`}, `{ A: 'B\\ A',  }`, nil},
		{"mixed escapes", struct{ A string }{A: "B\\\n'A \\"}, `{ A: 'B\\\n\'A \\',  }`, nil},
	}

	for _, c := range cases {
		t.Run(c.about, func(t *testing.T) {
			b, err := Marshal(c.in)
			if err != c.err {
				t.Errorf("Marshal got error %v, want %v", err, c.err)
			}
			if string(b) != c.out {
				t.Errorf("Marshal got %v, want %v", string(b), c.out)
			}
		})
	}
}

type TestPosition struct {
	Longitude float64
	Latitude  float64
}

type TestPeak struct {
	Name     string `json:"name"`
	Location TestPosition
	Ascent   time.Time
	Variants []string
	Camps    []TestPosition
}

func TestNested(t *testing.T) {
	p := TestPeak{
		Name: "пик Сталина",
		Location: TestPosition{
			38.916667, 72.016667,
		},
		Variants: []string{
			"Ismoil Somoni Peak",
			"Қуллаи Исмоили Сомонӣ",
		},
		Camps: []TestPosition{
			{38.916667, 72.016667},
			{38.916667, 72.016667},
			{38.916667, 72.016667},
		},
	}

	want := `{ name: 'пик Сталина', Location { Longitude: 38.916667, Latitude: 72.016667,  } Ascent: '0001-01-01T00:00:00Z', Variants: 'Ismoil Somoni Peak', Variants: 'Қуллаи Исмоили Сомонӣ', Camps { Longitude: 38.916667, Latitude: 72.016667,  } Camps { Longitude: 38.916667, Latitude: 72.016667,  } Camps { Longitude: 38.916667, Latitude: 72.016667,  }  }`

	b, err := Marshal(p)
	if err != nil {
		t.Error(err.Error())
	}
	if string(b) != want {
		t.Errorf("Marshal got %v, want %v", string(b), want)
	}
}

func TestDanglingCR(t *testing.T) {
	var cases = []struct {
		about string
		in    string
		out   string
		err   error
	}{
		{
			"japanese title with CR",
			`{"rft.atitle":"多様な生息地から採取したギョウジャニンニク系統の萌芽期の早晩性およびRAPD分析による分類\r Variations on Sprouting Time and Classification by RAPD Analysis of Allium victorialis L. Clones Collected from Diverse Habitats"}`,
			`{ rft.atitle: '多様な生息地から採取したギョウジャニンニク系統の萌芽期の早晩性およびRAPD分析による分類  Variations on Sprouting Time and Classification by RAPD Analysis of Allium victorialis L. Clones Collected from Diverse Habitats',  }`,
			nil,
		},
	}

	for _, c := range cases {
		t.Run(c.about, func(t *testing.T) {
			var v struct {
				ArticleTitle string `json:"rft.atitle"`
			}
			if err := json.Unmarshal([]byte(c.in), &v); err != nil {
				t.Fatal(err)
			}
			b, err := Marshal(v)
			if err != c.err {
				t.Errorf("got error %v, want %v", err, c.err)
			}
			if string(b) != c.out {
				t.Errorf("got %v, want %v", string(b), c.out)
			}
		})
	}
}

func BenchmarkMarshal(b *testing.B) {
	benchmarks := []struct {
		name string
		in   any
	}{
		{"simple", struct{ A string }{A: "B"}},
		{"slice", struct{ A []string }{A: []string{"B", "C", "D"}}},
		{"nested", TestPeak{
			Name:     "test",
			Location: TestPosition{38.916667, 72.016667},
			Variants: []string{"a", "b"},
			Camps:    []TestPosition{{1, 2}, {3, 4}},
		}},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Marshal(bm.in)
			}
		})
	}
}

func BenchmarkMarshalString(b *testing.B) {
	inputs := []string{
		"simple",
		"with 'quotes' and \\backslashes",
		"with\nnewlines\nand\ttabs",
	}
	for _, in := range inputs {
		b.Run(fmt.Sprintf("len=%d", len(in)), func(b *testing.B) {
			v := struct{ A string }{A: in}
			for i := 0; i < b.N; i++ {
				_, _ = Marshal(v)
			}
		})
	}
}
