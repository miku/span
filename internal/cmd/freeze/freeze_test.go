package freezecmd

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/segmentio/encoding/json"
)

// stubGetter returns canned responses keyed by URL.
type stubGetter struct {
	bodies map[string]string
	calls  []string
}

func (s *stubGetter) Get(url string) (*http.Response, error) {
	s.calls = append(s.calls, url)
	body, ok := s.bodies[url]
	if !ok {
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(""))}, nil
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func readZip(t *testing.T, b []byte) map[string]string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	out := make(map[string]string)
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", f.Name, err)
		}
		data, _ := io.ReadAll(rc)
		rc.Close()
		out[f.Name] = string(data)
	}
	return out
}

func TestRunLegacy(t *testing.T) {
	blob := "see http://example.com/a and http://example.com/b"
	client := &stubGetter{bodies: map[string]string{
		"http://example.com/a": "content-a",
		"http://example.com/b": "content-b",
	}}
	var buf bytes.Buffer
	if err := RunLegacy(LegacyConfig{}, client, strings.NewReader(blob), &buf); err != nil {
		t.Fatalf("RunLegacy: %v", err)
	}
	files := readZip(t, buf.Bytes())

	if files[NameBlob] != blob {
		t.Errorf("blob = %q, want %q", files[NameBlob], blob)
	}
	mappingRaw, ok := files[NameMapping]
	if !ok {
		t.Fatal("missing mapping.json")
	}
	var mapping map[string]string
	if err := json.Unmarshal([]byte(mappingRaw), &mapping); err != nil {
		t.Fatalf("mapping not JSON: %v", err)
	}
	if len(mapping) != 2 {
		t.Errorf("mapping has %d entries, want 2: %v", len(mapping), mapping)
	}
	// Each mapped file must be present with the fetched content.
	for url, name := range mapping {
		content, ok := files[name]
		if !ok {
			t.Errorf("missing content file %s for %s", name, url)
		}
		if !strings.HasPrefix(name, NameDir+"/") {
			t.Errorf("content file %s not under %s/", name, NameDir)
		}
		want := client.bodies[url]
		if content != want {
			t.Errorf("content for %s = %q, want %q", url, content, want)
		}
	}
}

func TestRunLegacyBestEffort(t *testing.T) {
	blob := "http://example.com/missing"
	client := &stubGetter{bodies: map[string]string{}} // 404 for everything
	var buf bytes.Buffer
	// Best effort: 404 should not fail the run.
	if err := RunLegacy(LegacyConfig{BestEffort: true}, client, strings.NewReader(blob), &buf); err != nil {
		t.Fatalf("RunLegacy best-effort: %v", err)
	}
	files := readZip(t, buf.Bytes())
	var mapping map[string]string
	if err := json.Unmarshal([]byte(files[NameMapping]), &mapping); err != nil {
		t.Fatal(err)
	}
	if len(mapping) != 0 {
		t.Errorf("expected empty mapping, got %v", mapping)
	}
}

func TestRunLegacyFailFast(t *testing.T) {
	blob := "http://example.com/missing"
	client := &stubGetter{bodies: map[string]string{}}
	var buf bytes.Buffer
	// Without best effort, a 404 should abort.
	if err := RunLegacy(LegacyConfig{}, client, strings.NewReader(blob), &buf); err == nil {
		t.Error("expected error on 404 without best-effort")
	}
}

func TestRunFolioValidation(t *testing.T) {
	if err := RunFolio(FolioConfig{OkapiURL: "http://x"}, "/tmp/out.zip"); err == nil {
		t.Error("expected error for missing token")
	}
	if err := RunFolio(FolioConfig{Token: "t"}, "/tmp/out.zip"); err == nil {
		t.Error("expected error for missing okapi url")
	}
}
