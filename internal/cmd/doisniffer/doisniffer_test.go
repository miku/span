package doisniffer

import (
	"bytes"
	"strings"
	"testing"

	json "github.com/segmentio/encoding/json"
)

func TestPostProcess(t *testing.T) {
	var cases = []struct {
		in, want string
	}{
		{"10.24072/pci.ecology.100076])", "10.24072/pci.ecology.100076"},
		{"10.1234/foo/epdf", "10.1234/foo"},
		{"10.1016/j.jenvp.2019.01.011)", "10.1016/j.jenvp.2019.01.011"},
		{"10.5329/RECADM.20090802005]", "10.5329/RECADM.20090802005"},
		{"10.1/x.", "10.1/x"},
		{"10.1/x,", "10.1/x"},
		{"  10.1/x  ", "10.1/x"},
		{"10.1/x", "10.1/x"},
		{"10.1/(foo)", "10.1/(foo)"}, // has matching '(' so not trimmed
	}
	for _, c := range cases {
		if got := postProcess(c.in); got != c.want {
			t.Errorf("postProcess(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestStringToRegexpSlice(t *testing.T) {
	res, err := stringToRegexpSlice("barcode,dewey", ",")
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("got %d patterns, want 2", len(res))
	}
	if _, err := stringToRegexpSlice("[", ","); err == nil {
		t.Error("expected error for invalid regexp")
	}
	res, err = stringToRegexpSlice("", ",")
	if err != nil || res != nil {
		t.Errorf("empty input: got %v, %v", res, err)
	}
}

func TestRunUpdatesDOI(t *testing.T) {
	input := `{"id":"rec-1","url":"http://dx.doi.org/10.1234/foo"}` + "\n"
	var buf bytes.Buffer
	cfg := DefaultConfig()
	cfg.NumWorkers = 1
	cfg.BatchSize = 1
	if err := Run(cfg, strings.NewReader(input), &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &doc); err != nil {
		t.Fatalf("output not JSON: %v (%q)", err, buf.String())
	}
	got, ok := doc["doi_str_mv"]
	if !ok {
		t.Fatalf("expected doi_str_mv in output, got: %v", doc)
	}
	vals, ok := got.([]any)
	if !ok || len(vals) == 0 || vals[0] != "10.1234/foo" {
		t.Errorf("doi_str_mv = %v, want [10.1234/foo]", got)
	}
}

func TestRunSkipsUnmatched(t *testing.T) {
	input := `{"id":"rec-1","title":"no doi here"}` + "\n"
	var buf bytes.Buffer
	cfg := DefaultConfig()
	cfg.NumWorkers = 1
	cfg.BatchSize = 1
	if err := Run(cfg, strings.NewReader(input), &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "" {
		t.Errorf("expected unmatched doc to be skipped, got: %q", buf.String())
	}
}
