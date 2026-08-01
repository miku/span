package compare

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrependHTTP(t *testing.T) {
	if got := prependHTTP("localhost:8983"); got != "http://localhost:8983" {
		t.Errorf("got %q", got)
	}
	if got := prependHTTP("https://x"); got != "https://x" {
		t.Errorf("got %q, want unchanged", got)
	}
}

func TestTabWriter(t *testing.T) {
	var buf bytes.Buffer
	w := &TabWriter{w: &buf}
	w.WriteHeader("A", "B") // no-op for tab writer
	w.WriteFields("DE-15", 49, "CrossRef")
	if w.Err() != nil {
		t.Fatal(w.Err())
	}
	if got := buf.String(); got != "DE-15\t49\tCrossRef\n" {
		t.Errorf("got %q", got)
	}
}

func TestTextileWriter(t *testing.T) {
	var buf bytes.Buffer
	w := &TextileWriter{w: &buf}
	w.WriteHeader("ISIL", "Source", "Name")
	w.WriteFields("DE-15", "49", "CrossRef")
	if w.Err() != nil {
		t.Fatal(w.Err())
	}
	out := buf.String()
	if !strings.Contains(out, "|_. ISIL |_. Source |_. Name |") {
		t.Errorf("header missing/wrong: %q", out)
	}
	if !strings.Contains(out, "| DE-15 | 49 | CrossRef |") {
		t.Errorf("row missing/wrong: %q", out)
	}
}

func TestTextileWriterColumnMismatch(t *testing.T) {
	var buf bytes.Buffer
	w := &TextileWriter{w: &buf}
	w.WriteHeader("A", "B")
	w.WriteFields("only-one") // wrong number of columns
	if w.Err() == nil {
		t.Error("expected error on column mismatch")
	}
}

func TestRenderSourceLink(t *testing.T) {
	data := struct{ SourceID string }{"49"}
	got, err := renderSourceLink("https://x/search?q=source_id:{{ .SourceID }}", data, "123")
	if err != nil {
		t.Fatal(err)
	}
	want := `"123":https://x/search?q=source_id:49`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if _, err := renderSourceLink("{{ .Bad", data, "1"); err == nil {
		t.Error("expected template parse error")
	}
}
