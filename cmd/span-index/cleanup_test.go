package main

import (
	"strings"
	"testing"
)

func TestShellSingleQuote(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"plain", "'plain'"},
		{`source_id:"53"`, `'source_id:"53"'`},
		{"it's", `'it'\''s'`},
	}
	for _, tt := range cases {
		if got := shellSingleQuote(tt.in); got != tt.want {
			t.Errorf("shellSingleQuote(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRenderCleanupCommand(t *testing.T) {
	q := `source_id:"53" AND last_indexed:[* TO 2026-01-01T00:00:00Z]`
	got, err := renderCleanupCommand("http://localhost:8983/solr/biblio", q, 12345)
	if err != nil {
		t.Fatal(err)
	}
	// Count comment on the first line.
	if !strings.HasPrefix(got, "# 12345 docs match "+q+"\n") {
		t.Errorf("missing/incorrect count comment, got:\n%s", got)
	}
	// curl hits the collection's update handler with a commit.
	if !strings.Contains(got, "curl 'http://localhost:8983/solr/biblio/update?commit=true'") {
		t.Errorf("missing update URL, got:\n%s", got)
	}
	// The query is carried as a JSON delete body with its quotes escaped.
	if !strings.Contains(got, `{"delete":{"query":"source_id:\"53\" AND last_indexed:[* TO 2026-01-01T00:00:00Z]"}}`) {
		t.Errorf("missing/incorrect JSON delete body, got:\n%s", got)
	}
}
