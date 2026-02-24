package span

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/segmentio/encoding/json"
)

func filenames(des []os.DirEntry) (result []string) {
	for _, de := range des {
		result = append(result, de.Name())
	}
	return
}

func TestUnfreezeFilterConfig(t *testing.T) {
	dir, blob, err := UnfreezeFilterConfig("fixtures/frozen.zip")
	if err != nil {
		t.Errorf("expected err nil, got %v", err)
	}
	f, err := os.Open(blob)
	if err != nil {
		t.Errorf("could not open blob file %v", err)
	}
	defer f.Close()
	defer os.RemoveAll(dir)

	dec := json.NewDecoder(f)
	var payload = make(map[string]any)
	if err := dec.Decode(&payload); err != nil {
		t.Errorf("could not decode JSON: %v", err)
	}
	fis, err := os.ReadDir(dir)
	if err != nil {
		t.Errorf("could not read dir: %v", err)
	}
	want := []string{"blob", "files", "mapping.json"}
	if !reflect.DeepEqual(filenames(fis), want) {
		t.Errorf("want %v, got %v", want, filenames(fis))
	}
}

// createTestZip builds a frozen zip file with the given blob and mapping for
// testing. Files referenced in the mapping are created with dummy content.
func createTestZip(t *testing.T, blobJSON string, mapping map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	// Write blob.
	bf, err := w.Create("blob")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bf.Write([]byte(blobJSON)); err != nil {
		t.Fatal(err)
	}
	// Write each file referenced in the mapping.
	for _, name := range mapping {
		ff, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ff.Write([]byte("test-content")); err != nil {
			t.Fatal(err)
		}
	}
	// Write mapping.json.
	mf, err := w.Create("mapping.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(mf).Encode(mapping); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestUnfreezeISSNLink verifies that an ISSN "url" field is correctly replaced
// with a file:// path after unfreezing. This is a regression test for the case
// where the link was not resolved because of a JSON key mismatch or brittle
// byte-level replacement.
func TestUnfreezeISSNLink(t *testing.T) {
	blob := `{
  "DE-15": {
    "and": [
      {"source": ["49"]},
      {"issn": {"url": "https://example.com/finc-config/files/abc-123"}}
    ]
  }
}`
	mapping := map[string]string{
		"https://example.com/finc-config/files/abc-123": "files/aaa111",
	}
	zipPath := createTestZip(t, blob, mapping)
	dir, blobPath, err := UnfreezeFilterConfig(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	b, err := os.ReadFile(blobPath)
	if err != nil {
		t.Fatal(err)
	}
	// The URL must have been replaced with a file:// path.
	if strings.Contains(string(b), "https://example.com") {
		t.Error("URL was not replaced in blob")
	}
	expectedPath := "file://" + filepath.Join(dir, "files/aaa111")
	if !strings.Contains(string(b), expectedPath) {
		t.Errorf("expected file path %q not found in blob:\n%s", expectedPath, b)
	}
}

// TestUnfreezePrefixURLs verifies that URLs sharing a common prefix are each
// replaced independently. For example, ".../as" must not corrupt ".../asxv".
func TestUnfreezePrefixURLs(t *testing.T) {
	blob := `{
  "DE-1": {
    "or": [
      {"holdings": {"urls": ["http://example.com/collections/as"]}},
      {"holdings": {"urls": ["http://example.com/collections/asx"]}},
      {"holdings": {"urls": ["http://example.com/collections/asxv"]}}
    ]
  }
}`
	mapping := map[string]string{
		"http://example.com/collections/as":   "files/file_as",
		"http://example.com/collections/asx":  "files/file_asx",
		"http://example.com/collections/asxv": "files/file_asxv",
	}
	zipPath := createTestZip(t, blob, mapping)
	dir, blobPath, err := UnfreezeFilterConfig(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	b, err := os.ReadFile(blobPath)
	if err != nil {
		t.Fatal(err)
	}
	// No original URLs should remain.
	if strings.Contains(string(b), "http://example.com") {
		t.Error("original URLs remain in blob")
	}
	// Parse the result and verify each URL was replaced with its own file.
	var result map[string]any
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	orList := result["DE-1"].(map[string]any)["or"].([]any)
	wantFiles := []string{"file_as", "file_asx", "file_asxv"}
	for i, entry := range orList {
		urls := entry.(map[string]any)["holdings"].(map[string]any)["urls"].([]any)
		got := urls[0].(string)
		wantSuffix := "files/" + wantFiles[i]
		if !strings.HasSuffix(got, wantSuffix) {
			t.Errorf("entry %d: got %q, want suffix %q", i, got, wantSuffix)
		}
		if !strings.HasPrefix(got, "file://") {
			t.Errorf("entry %d: got %q, want file:// prefix", i, got)
		}
	}
}

// TestUnfreezeURLWithAmpersand verifies that URLs containing & are correctly
// replaced. JSON encoders may encode & as \u0026, which broke the old
// byte-level replacement using %q.
func TestUnfreezeURLWithAmpersand(t *testing.T) {
	// The blob uses \u0026 for & (as some JSON encoders do).
	blob := `{
  "DE-1": {
    "holdings": {
      "urls": ["http://example.com/?a=1\u0026b=2"]
    }
  }
}`
	mapping := map[string]string{
		"http://example.com/?a=1&b=2": "files/ampersand_file",
	}
	zipPath := createTestZip(t, blob, mapping)
	dir, blobPath, err := UnfreezeFilterConfig(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	b, err := os.ReadFile(blobPath)
	if err != nil {
		t.Fatal(err)
	}
	expected := "file://" + filepath.Join(dir, "files/ampersand_file")
	if !strings.Contains(string(b), expected) {
		t.Errorf("expected %q in blob, got:\n%s", expected, b)
	}
}
