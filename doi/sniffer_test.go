package doi

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type M map[string]interface{}

func TestSniffer(t *testing.T) {
	var cases = []struct {
		about      string
		doc        M
		ignoreKeys []*regexp.Regexp
		doi        []string
	}{
		{
			"nothing to find",
			M{
				"a": "b",
			},
			nil,
			nil,
		},
		{
			"a single field",
			M{
				"a": "10.5617/osla.8504",
			},
			nil,
			[]string{"10.5617/osla.8504"},
		},
		{
			"extra chars",
			M{
				"a": "some value: 10.5617/osla.8504",
			},
			nil,
			[]string{"10.5617/osla.8504"},
		},
		{
			"only the first is discovered",
			M{
				"a": "some value: 10.5617/osla.8504 and 10.5617/osla.8503",
			},
			nil,
			[]string{"10.5617/osla.8504"},
		},
		{
			"only the first is discovered",
			M{
				"a": "some value: 10.5617/osla.8504 and 10.5617/osla.8503",
			},
			nil,
			[]string{"10.5617/osla.8504"},
		},
		{
			"ignore key",
			M{
				"a": "some value: 10.5617/osla.8504 and 10.5617/osla.8503",
			},
			[]*regexp.Regexp{
				regexp.MustCompile(`a`),
			},
			nil,
		},
		{
			"find in list",
			M{
				"a": []string{
					"some value: 10.5617/osla.8504",
				},
			},
			nil,
			[]string{"10.5617/osla.8504"},
		},
	}
	sniffer := &MapSniffer{
		Pattern: regexp.MustCompile(PatDOI),
	}
	for _, c := range cases {
		sniffer.IgnoreKeys = c.ignoreKeys
		result := sniffer.SearchMap(c.doc)
		if !cmp.Equal(result, c.doi) {
			t.Fatalf("got %v, want %v [%s]", result, c.doi, c.about)
		}
	}
}
