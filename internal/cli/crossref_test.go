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
	const body = `{"status":"ok","message":{"total-results":1,"next-cursor":"x","items":[{"DOI":"10.1/a","indexed":{"timestamp":1}}]}}`
	day := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	iv := crossrefcmd.Interval{Start: day, End: day}
	for _, program := range []string{"gzip", "zstd"} {
		t.Run(program, func(t *testing.T) {
			dir := t.TempDir()
			cachePath := filepath.Join(dir, "window.json."+program)
			var calls int
			sync := &crossrefcmd.Sync{ApiFilter: "index", Rows: 1000, Client: doerFunc(func(*http.Request) (*http.Response, error) {
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
	t.Run("failed sync leaves no window file", func(t *testing.T) {
		dir := t.TempDir()
		sync := &crossrefcmd.Sync{ApiFilter: "index", Rows: 1000, Client: doerFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("network down")
		})}
		cachePath := filepath.Join(dir, "window.json.gz")
		if err := syncWindow(sync, iv, cachePath, "gzip"); err == nil {
			t.Fatal("expected error")
		}
		if _, err := os.Stat(cachePath); err == nil {
			t.Errorf("expected no window file")
		}
	})
}

func TestLockCache(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "feed-1-index-2026-08-20-2026-08-20.json.zst-tmp-123")
	other := filepath.Join(dir, "feed-2-index-2026-08-20-2026-08-20.json.zst-tmp-123")
	for _, name := range []string{stale, other} {
		if err := os.WriteFile(name, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	unlock, err := lockCache(dir, "feed-1-")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lockCache(dir, "feed-1-"); err == nil {
		t.Error("expected second lock to fail")
	}
	if _, err := os.Stat(stale); err == nil {
		t.Error("stale temporary file not removed")
	}
	if _, err := os.Stat(other); err != nil {
		t.Error("temporary file of other prefix removed")
	}
	if err := unlock(); err != nil {
		t.Fatal(err)
	}
	unlock, err = lockCache(dir, "feed-1-")
	if err != nil {
		t.Fatal(err)
	}
	unlock()
}
