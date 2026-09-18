package cli

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/miku/span/internal/cmd/crossrefcmd"
)

// doerFunc adapts a function to crossrefcmd.Doer.
type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }

func TestSyncWindow(t *testing.T) {
	const body = `{"status":"ok","message":{"total-results":1,"next-cursor":"x","items":[{"DOI":"10.1/a"}]}}`
	day := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	iv := crossrefcmd.Interval{Start: day, End: day}
	for _, program := range []string{"gzip", "zstd"} {
		t.Run(program, func(t *testing.T) {
			dir := t.TempDir()
			cachePath := filepath.Join(dir, "window.json."+program)
			var calls int
			sync := &crossrefcmd.Sync{Rows: 1000, Mode: "s", Client: doerFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
			})}
			for range 2 {
				if err := syncWindow(sync, iv, cachePath, program); err != nil {
					t.Fatal(err)
				}
			}
			if calls != 1 {
				t.Errorf("got %d calls, want 1, an existing window must not be synced again", calls)
			}
			var buf bytes.Buffer
			if err := copyDecompressed(&buf, cachePath, program); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(buf.String(), `"10.1/a"`) {
				t.Errorf("unexpected content: %q", buf.String())
			}
		})
	}
	t.Run("failed sync leaves no file", func(t *testing.T) {
		dir := t.TempDir()
		sync := &crossrefcmd.Sync{Rows: 1000, Mode: "s", Client: doerFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("network down")
		})}
		if err := syncWindow(sync, iv, filepath.Join(dir, "window.json.gz"), "gzip"); err == nil {
			t.Fatal("expected error")
		}
		if entries, _ := os.ReadDir(dir); len(entries) != 0 {
			t.Errorf("expected empty cache dir, got %v", entries)
		}
	})
}
