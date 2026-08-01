package crossrefcmd

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFields(t *testing.T) {
	var buf bytes.Buffer
	n, err := WriteFields(&buf, "\t", 42, "2026-03-02", "10.1/x")
	if err != nil {
		t.Fatal(err)
	}
	want := "42\t2026-03-02\t10.1/x\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
	if n != len(want) {
		t.Errorf("n = %d, want %d", n, len(want))
	}
}

func TestStage1Extract(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(nil)

	in := `{"DOI":"10.1/keep","indexed":{"date-parts":[[2026,3,2]]}}` + "\n" +
		`{"DOI":"10.2/drop","indexed":{"date-parts":[[2025,1,1]]}}` + "\n"
	excludes := map[string]struct{}{"10.2/drop": {}}
	var buf bytes.Buffer
	if err := Stage1Extract(DefaultStage1Config(), excludes, strings.NewReader(in), &buf); err != nil {
		t.Fatal(err)
	}
	out := strings.TrimRight(buf.String(), "\n")
	// Only the non-excluded record survives.
	if out != "1\t2026-03-02\t10.1/keep" {
		t.Errorf("got %q", out)
	}
}

func TestStage1ExtractErrorThreshold(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(nil)

	// Two malformed lines; threshold 1 tolerates the first, fails on the second.
	in := "not json\nalso not json\n"
	cfg := DefaultStage1Config() // ErrCountThreshold 1
	var buf bytes.Buffer
	if err := Stage1Extract(cfg, nil, strings.NewReader(in), &buf); err == nil {
		t.Error("expected error once threshold exceeded")
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := CopyFile(dst, src, 0644); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "hello" {
		t.Errorf("got %q", string(b))
	}
}
