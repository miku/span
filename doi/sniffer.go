// Package doi helps to find DOI in JSON documents.
package doi

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/miku/parallel"
	"github.com/miku/span/container"
	"github.com/segmentio/encoding/json"
)

const PatDOI = "10[.][0-9]{2,6}/[^ \"\u001f\u001e]{3,}"

var bNewline = []byte("\n")

// Sniffer can read, transform and write a stream of newline delimited JSON
// documents.
type Sniffer struct {
	Reader         io.Reader
	Writer         io.Writer
	MapSniffer     *MapSniffer
	IdentifierKey  string
	UpdateKey      string // if set, update the document in place
	SkipUnmatched  bool
	ForceOverwrite bool // if set, overwrite existing values in "UpdateKey"
	PostProcess    func(s string) string
	BatchSize      int
	NumWorkers     int
}

// NewSniffer sets up a new sniffer with defaults keys matching the current
// SOLR schema. Can process around 20K docs/s.
func NewSniffer(r io.Reader, w io.Writer) *Sniffer {
	return &Sniffer{
		Reader:        r,
		Writer:        w,
		IdentifierKey: "id",
		UpdateKey:     "doi_str_mv",
		MapSniffer: &MapSniffer{
			Pattern: regexp.MustCompile(PatDOI),
			IgnoreKeys: []*regexp.Regexp{
				regexp.MustCompile(`barcode`),
				regexp.MustCompile(`dewey`),
			},
		},
		PostProcess: func(s string) string {
			switch {
			case strings.HasSuffix(s, "/epdf"):
				return s[:len(s)-5]
			case strings.HasSuffix(s, ".") || strings.HasSuffix(s, "*"):
				return s[:len(s)-1]
			default:
				return s
			}
		},
	}
}

// Run sniffs out DOI and eventually updates a document in place.
func (s *Sniffer) Run() error {
	pp := parallel.NewProcessor(s.Reader, s.Writer, func(p []byte) ([]byte, error) {
		var (
			data   map[string]any
			result []string
		)
		if err := json.Unmarshal(p, &data); err != nil {
			return nil, err
		}
		for _, v := range s.MapSniffer.SearchMap(data) {
			result = append(result, s.PostProcess(v))
		}
		if s.SkipUnmatched && len(result) == 0 {
			return nil, nil
		}
		switch {
		case s.UpdateKey != "":
			if len(result) > 0 {
				v, ok := data[s.UpdateKey]
				if !ok || s.ForceOverwrite {
					var shouldOverwrite bool
					switch w := v.(type) {
					// Overwrite, if the existing value v does not contain any
					// value or if it is not a list at all (which we do not
					// expect).
					case []string:
						shouldOverwrite = len(w) == 0
					case []any:
						shouldOverwrite = len(w) == 0
					default:
						shouldOverwrite = true
					}
					if shouldOverwrite {
						data[s.UpdateKey] = result
					}
				}
			}
			b, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			b = append(b, bNewline...)
			return b, nil
		default:
			id, ok := data[s.IdentifierKey]
			if !ok {
				return nil, fmt.Errorf("missing identifier key: %s", s.IdentifierKey)
			}
			s := fmt.Sprintf("%s\t%s\n", id, strings.Join(result, "\t"))
			return []byte(s), nil
		}
	})
	pp.NumWorkers = s.NumWorkers
	pp.BatchSize = s.BatchSize
	return pp.Run()
}

// MapSniffer tries to find values in a map.
type MapSniffer struct {
	Pattern    *regexp.Regexp
	IgnoreKeys []*regexp.Regexp
}

func (s *MapSniffer) SearchMap(doc map[string]any) []string {
	ss := container.NewStringSet()
	for k, v := range doc {
		if anyMatchString(s.IgnoreKeys, k) {
			continue
		}
		switch w := v.(type) {
		case string:
			if m := s.Pattern.FindString(w); m != "" {
				ss.Add(m)
			}
		case []string:
			for _, e := range w {
				if m := s.Pattern.FindString(e); m != "" {
					ss.Add(m)
				}
			}
		default:
			continue
		}
	}
	return ss.SortedValues()
}

// anyMatchString is like MatchString but for a list of pattern.
func anyMatchString(r []*regexp.Regexp, v string) bool {
	for _, p := range r {
		if p.MatchString(v) {
			return true
		}
	}
	return false
}
