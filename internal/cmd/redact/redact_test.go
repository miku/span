package redact

import (
	"bytes"
	"strings"
	"testing"

	"github.com/miku/span/formats/finc"

	json "github.com/segmentio/encoding/json"
)

func TestRun(t *testing.T) {
	var cases = []struct {
		name  string
		input string
	}{
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "single record with fulltext",
			input: `{"finc.id":"ai-1","finc.source_id":"1","x.fulltext":"lorem ipsum"}` + "\n",
		},
		{
			name: "multiple records",
			input: `{"finc.id":"ai-1","x.fulltext":"lorem ipsum"}` + "\n" +
				`{"finc.id":"ai-2","x.fulltext":"dolor sit amet"}` + "\n" +
				`{"finc.id":"ai-3"}` + "\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			// A single worker keeps output order deterministic for the
			// assertions below.
			cfg := Config{BatchSize: 1, NumWorkers: 1}
			if err := Run(cfg, strings.NewReader(c.input), &buf); err != nil {
				t.Fatalf("Run: %v", err)
			}
			in := splitLines(c.input)
			out := splitLines(buf.String())
			if len(in) != len(out) {
				t.Fatalf("got %d output lines, want %d", len(out), len(in))
			}
			for i := range in {
				var before, after finc.IntermediateSchema
				if err := json.Unmarshal([]byte(in[i]), &before); err != nil {
					t.Fatalf("unmarshal input: %v", err)
				}
				if err := json.Unmarshal([]byte(out[i]), &after); err != nil {
					t.Fatalf("unmarshal output: %v", err)
				}
				if after.Fulltext != "" {
					t.Errorf("line %d: fulltext not redacted: %q", i, after.Fulltext)
				}
				// Everything except fulltext must be preserved.
				before.Fulltext = ""
				if before.ID != after.ID || before.SourceID != after.SourceID {
					t.Errorf("line %d: non-fulltext fields changed: %+v vs %+v", i, before, after)
				}
			}
		})
	}
}

func splitLines(s string) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
