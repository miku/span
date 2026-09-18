package atomicfile

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCommit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := Create(path, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString("new"); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != "old" {
		t.Fatalf("target changed before commit: %q", b)
	}
	if err := f.Commit(); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != "new" {
		t.Errorf("got %q, want new", b)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0644 {
		t.Errorf("got mode %v, want 0644", fi.Mode().Perm())
	}
	if err := f.Commit(); err == nil {
		t.Error("expected error on second commit")
	}
	assertOnly(t, dir, "out.txt")
}

func TestCloseWithoutCommit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	f, err := Create(path, 0644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("partial"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected no file at %s, got %v", path, err)
	}
	assertOnly(t, dir)
}

func TestCreateMissingDir(t *testing.T) {
	if _, err := Create(filepath.Join(t.TempDir(), "missing", "out.txt"), 0644); err == nil {
		t.Error("expected error for missing directory")
	}
}

// assertOnly checks that dir contains exactly the given names.
func assertOnly(t *testing.T, dir string, names ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range entries {
		got = append(got, e.Name())
	}
	if !slices.Equal(got, names) {
		t.Errorf("got %v, want %v", got, names)
	}
}
