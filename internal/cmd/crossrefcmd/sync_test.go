package crossrefcmd

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// script serves responses in sequence and records the request queries.
type script struct {
	responses []string // body, or "HTTP 429" and the like
	queries   []string
}

func (s *script) Do(r *http.Request) (*http.Response, error) {
	s.queries = append(s.queries, r.URL.RawQuery)
	if len(s.responses) == 0 {
		return nil, errors.New("no more responses")
	}
	body := s.responses[0]
	s.responses = s.responses[1:]
	var status int
	if _, err := fmt.Sscanf(body, "HTTP %d", &status); err == nil {
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(""))}, nil
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
}

func item(doi string, ts int64) string {
	return fmt.Sprintf(`{"DOI":%q,"indexed":{"timestamp":%d}}`, doi, ts)
}

func page(total int, cursor string, items ...string) string {
	return fmt.Sprintf(`{"status":"ok","message":{"total-results":%d,"next-cursor":%q,"items":[%s]}}`,
		total, cursor, strings.Join(items, ","))
}

func gzipCompressor(w io.Writer) (io.WriteCloser, error) { return gzip.NewWriter(w), nil }

func newTestSync(d Doer) *Sync {
	return &Sync{
		ApiEndpoint: "http://example.com/works",
		ApiFilter:   "index",
		Rows:        2,
		Client:      d,
		Backoff:     func(int) time.Duration { return 0 },
	}
}

// readDOIs returns the DOIs of all records in a gzip file.
func readDOIs(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	var dois []string
	for line := range strings.Lines(string(b)) {
		doi, _, err := recordKey([]byte(line), "indexed")
		if err != nil {
			t.Fatal(err)
		}
		dois = append(dois, doi)
	}
	return dois
}

var day = time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)

func TestHarvestWindow(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	d := &script{responses: []string{
		page(3, "c1", item(`10.1\/a`, 1000), item("10.1/b", 2000)),
		page(3, "c2", item("10.1/c", 3000)),
	}}
	path := filepath.Join(t.TempDir(), "window.json.gz")
	if err := newTestSync(d).HarvestWindow(path, gzipCompressor, day, day); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(readDOIs(t, path), " "); got != "10.1/a 10.1/b 10.1/c" {
		t.Errorf("got %s", got)
	}
	// Every request must carry the full query, otherwise the API falls
	// back to all works, twenty at a time.
	for _, q := range d.queries {
		for _, want := range []string{"filter=from-index-date", "rows=2", "sort=indexed", "order=asc"} {
			if !strings.Contains(q, want) {
				t.Errorf("query %q misses %q", q, want)
			}
		}
	}
	if entries, _ := os.ReadDir(filepath.Dir(path)); len(entries) != 1 {
		t.Errorf("expected only the window file, got %v", entries)
	}
}

func TestHarvestWindowResume(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	ts := day.Add(90 * time.Minute).UnixMilli()
	d := &script{responses: []string{
		page(4, "c1", item("10.1/a", ts), item("10.1/b", ts+500)),
		"HTTP 429",
		// Resumed at the second of b, so a and b come again.
		page(3, "c2", item("10.1/a", ts), item("10.1/b", ts+500)),
		page(3, "c3", item("10.1/c", ts+500), item("10.1/d", ts+2000)),
	}}
	path := filepath.Join(t.TempDir(), "window.json.gz")
	s := newTestSync(d)
	s.MaxResumes = 1
	if err := s.HarvestWindow(path, gzipCompressor, day, day); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(readDOIs(t, path), " "); got != "10.1/a 10.1/b 10.1/c 10.1/d" {
		t.Errorf("got %s", got)
	}
	if !strings.Contains(d.queries[2], "from-index-date%3A2026-08-20T01%3A30%3A00") {
		t.Errorf("resume should start at last record, got %s", d.queries[2])
	}
}

func TestHarvestWindowResumeNextRun(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	path := filepath.Join(t.TempDir(), "window.json.gz")
	d := &script{responses: []string{
		page(3, "c1", item("10.1/a", 1000), item("10.1/b", 2000)),
		"HTTP 500",
	}}
	if err := newTestSync(d).HarvestWindow(path, gzipCompressor, day, day); err == nil {
		t.Fatal("expected error")
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("incomplete window must not be committed")
	}
	// Simulate a write interrupted after the last checkpoint.
	f, err := os.OpenFile(path+".partial", os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("garbage")
	f.Close()
	d = &script{responses: []string{
		page(2, "c2", item("10.1/b", 2000), item("10.1/c", 3000)),
		page(2, ""),
	}}
	if err := newTestSync(d).HarvestWindow(path, gzipCompressor, day, day); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(readDOIs(t, path), " "); got != "10.1/a 10.1/b 10.1/c" {
		t.Errorf("got %s", got)
	}
}

func TestHarvestWindowIncomplete(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	// The API reports more results than it delivers. A fresh query from the
	// checkpoint yielding nothing new ends the harvest.
	d := &script{responses: []string{
		page(5, "c1", item("10.1/a", 1000), item("10.1/b", 2000)),
		page(5, "c2"),
		page(4, "c3", item("10.1/b", 2000)),
		page(4, "c4"),
	}}
	path := filepath.Join(t.TempDir(), "window.json.gz")
	s := newTestSync(d)
	s.MaxResumes = 1
	if err := s.HarvestWindow(path, gzipCompressor, day, day); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(readDOIs(t, path), " "); got != "10.1/a 10.1/b" {
		t.Errorf("got %s", got)
	}
}

func TestHarvestWindowEmpty(t *testing.T) {
	d := &script{responses: []string{page(0, "")}}
	path := filepath.Join(t.TempDir(), "window.json.gz")
	if err := newTestSync(d).HarvestWindow(path, gzipCompressor, day, day); err != nil {
		t.Fatal(err)
	}
	if got := readDOIs(t, path); len(got) != 0 {
		t.Errorf("got %v", got)
	}
}

func TestHarvestWindowFails(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	d := &script{responses: []string{"HTTP 429", `{"status":"error","message":{}}`, "HTTP 503"}}
	s := newTestSync(d)
	s.MaxResumes = 2
	path := filepath.Join(t.TempDir(), "window.json.gz")
	if err := s.HarvestWindow(path, gzipCompressor, day, day); err == nil {
		t.Fatal("expected error")
	}
	if len(d.queries) != 3 {
		t.Errorf("got %d requests, want 3", len(d.queries))
	}
}

func TestCount(t *testing.T) {
	d := &script{responses: []string{page(123, "")}}
	total, err := newTestSync(d).Count(day, day)
	if err != nil {
		t.Fatal(err)
	}
	if total != 123 {
		t.Errorf("got %d", total)
	}
	if !strings.Contains(d.queries[0], "rows=0") {
		t.Errorf("got %s", d.queries[0])
	}
}
