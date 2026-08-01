package crossrefcmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/miku/span/filter"
)

func TestOutputFilename(t *testing.T) {
	var cases = []struct {
		in, want string
	}{
		{
			"feed-2-index-2026-03-02-2026-03-02.json.zst",
			"feed-2-index-2026-03-02-2026-03-02-solr-export-with-fullrecord.json.zst",
		},
		{
			"/some/dir/slice.json.zst",
			"slice-solr-export-with-fullrecord.json.zst",
		},
		{
			"slice.json",
			"slice-solr-export-with-fullrecord.json.zst",
		},
	}
	for _, c := range cases {
		if got := OutputFilename(c.in); got != c.want {
			t.Errorf("OutputFilename(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestProcessStreamBadJSON(t *testing.T) {
	var tagger filter.Tagger
	var buf bytes.Buffer
	err := ProcessStream(DefaultFastprocConfig(), &tagger, strings.NewReader("not json\n"), &buf)
	if err == nil {
		t.Error("expected error on malformed input")
	}
}

func TestProcessStreamEmpty(t *testing.T) {
	var tagger filter.Tagger
	var buf bytes.Buffer
	if err := ProcessStream(DefaultFastprocConfig(), &tagger, strings.NewReader(""), &buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}
