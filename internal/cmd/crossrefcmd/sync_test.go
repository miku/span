package crossrefcmd

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

// fakeDoer returns canned responses in sequence.
type fakeDoer struct {
	bodies []string
	calls  int
}

func (d *fakeDoer) Do(*http.Request) (*http.Response, error) {
	body := d.bodies[d.calls]
	d.calls++
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func TestWriteWindowSync(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(nil)

	// Single page, two items, total-results 2 -> loop ends after first page.
	// One item has an escaped slash to exercise cleanEscapedSlashes.
	body := `{"status":"ok","message":{"total-results":2,"next-cursor":"x","items":[` +
		`{"DOI":"10.1\/a"},{"DOI":"10.2/b"}]}}`
	s := &Sync{
		ApiEndpoint: "http://example.com/works",
		ApiFilter:   "index",
		Rows:        1000,
		Mode:        "s",
		Client:      &fakeDoer{bodies: []string{body}},
	}
	var buf bytes.Buffer
	f := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := s.WriteWindow(&buf, f, f); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 items, got %d: %q", len(lines), out)
	}
	// escaped slash should be unescaped
	if !strings.Contains(lines[0], `10.1/a`) {
		t.Errorf("expected unescaped slash, got %q", lines[0])
	}
}

func TestWriteWindowTabs(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(nil)

	body := `{"status":"ok","message":{"total-results":123,"items":[{},{},{}]}}`
	s := &Sync{
		ApiEndpoint: "http://example.com/works",
		ApiFilter:   "index",
		Rows:        1000,
		Mode:        "t",
		Client:      &fakeDoer{bodies: []string{body}},
	}
	var buf bytes.Buffer
	f := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	if err := s.WriteWindow(&buf, f, f); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "2026-01-02\t3\t123\n" {
		t.Errorf("got %q", got)
	}
}

func TestWriteWindowStatusNotOK(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(nil)

	s := &Sync{
		ApiEndpoint: "http://example.com/works",
		ApiFilter:   "index",
		Mode:        "s",
		Client:      &fakeDoer{bodies: []string{`{"status":"error","message":{}}`}},
	}
	var buf bytes.Buffer
	f := time.Now()
	if err := s.WriteWindow(&buf, f, f); err == nil {
		t.Error("expected error on non-ok status")
	}
}
