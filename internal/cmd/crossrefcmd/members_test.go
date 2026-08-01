package crossrefcmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
)

func mkResp(body string) *http.Response {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestRunMembersPagination(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(nil)

	// Three canned pages; the third has start-index >= total-results and stops
	// the loop before being printed.
	pages := []string{
		`{"status":"ok","message":{"query":{"start-index":0},"total-results":40}}`,
		`{"status":"ok","message":{"query":{"start-index":20},"total-results":40}}`,
		`{"status":"ok","message":{"query":{"start-index":40},"total-results":40}}`,
	}
	call := 0
	get := func(url string) (*http.Response, error) {
		if call >= len(pages) {
			return nil, fmt.Errorf("unexpected extra call %d", call)
		}
		r := mkResp(pages[call])
		call++
		return r, nil
	}

	cfg := DefaultMembersConfig()
	cfg.Sleep = 0
	var buf bytes.Buffer
	if err := RunMembers(cfg, get, &buf); err != nil {
		t.Fatal(err)
	}
	// Two pages printed (the terminating third is not).
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 printed pages, got %d: %q", len(lines), buf.String())
	}
	if call != 3 {
		t.Errorf("expected 3 API calls, got %d", call)
	}
}

func TestRunMembersStatusNotOK(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(nil)
	get := func(url string) (*http.Response, error) {
		return mkResp(`{"status":"error","message":{}}`), nil
	}
	var buf bytes.Buffer
	if err := RunMembers(DefaultMembersConfig(), get, &buf); err == nil {
		t.Error("expected error on non-ok status")
	}
}

func TestRunMembersRetryExceeded(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(nil)
	get := func(url string) (*http.Response, error) {
		return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader(""))}, nil
	}
	cfg := DefaultMembersConfig()
	cfg.RetryCount = 0
	var buf bytes.Buffer
	// With RetryCount 0, the first 500 increments retries to 1 > 0 on the next
	// iteration and returns an error. time.Sleep(2s) fires once here.
	if err := RunMembers(cfg, get, &buf); err == nil {
		t.Error("expected retry-count-exceeded error")
	}
}
