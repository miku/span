package crossrefcmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/miku/span/crossref"
)

func TestRunTable(t *testing.T) {
	in := `{"DOI":"10.1234/foo","member":"http://id.crossref.org/member/78"}` + "\n"
	var buf bytes.Buffer
	if err := RunTable(DefaultTableConfig(), strings.NewReader(in), &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "10.1234/foo\t") {
		t.Errorf("expected line to start with DOI, got %q", out)
	}
	fields := strings.Split(strings.TrimRight(out, "\n"), "\t")
	if len(fields) != 6 {
		t.Errorf("expected 6 tab-separated fields, got %d: %q", len(fields), out)
	}
	if len(fields[5]) != 32 {
		t.Errorf("expected 32-char md5, got %q", fields[5])
	}
}

func TestRunTableSkipsBadLine(t *testing.T) {
	var buf bytes.Buffer
	if err := RunTable(DefaultTableConfig(), strings.NewReader("not json\n"), &buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected bad line to be skipped, got %q", buf.String())
	}
}

func TestParseExcludes(t *testing.T) {
	got, err := ParseExcludes(strings.NewReader("10.1/a\n10.2/b\n"))
	if err != nil {
		t.Fatal(err)
	}
	// strings.Split on trailing newline yields a trailing empty element
	// (preserved behavior).
	want := []string{"10.1/a", "10.2/b", ""}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFastSnapshotOptionsAndRun(t *testing.T) {
	cfg := DefaultFastSnapshotConfig()
	cfg.InputFiles = []string{"a.zst", "b.zst"}
	cfg.Excludes = []string{"10.1/x"}
	cfg.CacheDir = "/tmp/cache"

	opts := cfg.Options()
	if len(opts.InputFiles) != 2 || opts.InputFiles[0] != "a.zst" {
		t.Errorf("input files not mapped: %v", opts.InputFiles)
	}
	if opts.BatchSize != 100000 || opts.SortBufferSize != "25%" || !opts.CacheEnabled {
		t.Errorf("defaults not mapped: %+v", opts)
	}
	if opts.CacheDir != "/tmp/cache" || len(opts.Excludes) != 1 {
		t.Errorf("fields not mapped: %+v", opts)
	}

	// RunFastSnapshot should hand the mapped options to the injected func.
	var got crossref.SnapshotOptions
	err := RunFastSnapshot(cfg, func(o crossref.SnapshotOptions) error {
		got = o
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.InputFiles) != 2 {
		t.Errorf("injected snapshot got %+v", got)
	}
}
