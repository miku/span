package freeze

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeZip creates a frozen zip with the given blob JSON and mapping using only
// stdlib. Returns the path to the zip file.
func makeZip(t *testing.T, blobJSON string, mapping map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	bf, err := w.Create(nameBlob)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bf.Write([]byte(blobJSON)); err != nil {
		t.Fatal(err)
	}
	for _, name := range mapping {
		ff, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ff.Write([]byte("dummy")); err != nil {
			t.Fatal(err)
		}
	}
	mf, err := w.Create(nameMapping)
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

func TestUnfreezeCreatesExpectedFiles(t *testing.T) {
	blob := `{"DE-1": {"source": ["49"]}}`
	zp := makeZip(t, blob, map[string]string{})
	dir, blobPath, err := UnfreezeFilterConfig(zp)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if _, err := os.Stat(blobPath); err != nil {
		t.Fatalf("blob file not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, nameMapping)); err != nil {
		t.Fatalf("%s not created: %v", nameMapping, err)
	}
	if _, err := os.Stat(filepath.Join(dir, nameDir)); err != nil {
		t.Fatalf("%s dir not created: %v", nameDir, err)
	}
}

func TestUnfreezeReplacesURLs(t *testing.T) {
	blob := `{"DE-1": {"holdings": {"urls": ["http://example.com/data.tsv"]}}}`
	mapping := map[string]string{
		"http://example.com/data.tsv": "files/abc123",
	}
	zp := makeZip(t, blob, mapping)
	dir, blobPath, err := UnfreezeFilterConfig(zp)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	b, err := os.ReadFile(blobPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "http://example.com") {
		t.Error("URL was not replaced")
	}
	want := "file://" + filepath.Join(dir, "files/abc123")
	if !strings.Contains(string(b), want) {
		t.Errorf("expected %q in blob, got:\n%s", want, b)
	}
}

func TestUnfreezeNoMapping(t *testing.T) {
	blob := `{"DE-1": {"source": ["49"]}}`
	zp := makeZip(t, blob, map[string]string{})
	dir, blobPath, err := UnfreezeFilterConfig(zp)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	b, err := os.ReadFile(blobPath)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["DE-1"]; !ok {
		t.Error("expected DE-1 key in output")
	}
}

func TestUnfreezeNonExistentFile(t *testing.T) {
	_, _, err := UnfreezeFilterConfig("/no/such/file.zip")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestReplaceURLsString(t *testing.T) {
	m := map[string]string{"http://x.com/a": "files/fa"}
	got := replaceURLs("http://x.com/a", m, "/tmp/d")
	want := "file:///tmp/d/files/fa"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReplaceURLsNoMatch(t *testing.T) {
	m := map[string]string{"http://x.com/a": "files/fa"}
	got := replaceURLs("http://other.com/b", m, "/tmp/d")
	if got != "http://other.com/b" {
		t.Errorf("non-matching string was modified: %v", got)
	}
}

func TestReplaceURLsNested(t *testing.T) {
	m := map[string]string{
		"http://a.com/1": "files/f1",
		"http://a.com/2": "files/f2",
	}
	input := map[string]any{
		"x": []any{"http://a.com/1", "http://a.com/2", "unchanged"},
		"y": map[string]any{"z": "http://a.com/1"},
	}
	got := replaceURLs(input, m, "/d").(map[string]any)
	xs := got["x"].([]any)
	if xs[0] != "file:///d/files/f1" {
		t.Errorf("xs[0] = %v", xs[0])
	}
	if xs[1] != "file:///d/files/f2" {
		t.Errorf("xs[1] = %v", xs[1])
	}
	if xs[2] != "unchanged" {
		t.Errorf("xs[2] = %v", xs[2])
	}
	z := got["y"].(map[string]any)["z"]
	if z != "file:///d/files/f1" {
		t.Errorf("z = %v", z)
	}
}

func TestReplaceURLsNonStringTypes(t *testing.T) {
	got := replaceURLs(42.0, nil, "/d")
	if got != 42.0 {
		t.Errorf("numeric value was modified: %v", got)
	}
	got = replaceURLs(true, nil, "/d")
	if got != true {
		t.Errorf("bool value was modified: %v", got)
	}
	got = replaceURLs(nil, nil, "/d")
	if got != nil {
		t.Errorf("nil was modified: %v", got)
	}
}
