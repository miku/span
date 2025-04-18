package xio

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLinkReader(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping going out to the net")
	}
	r := &LinkReader{
		Link: "https://httpbin.org/bytes/1?seed=0",
	}
	var buf bytes.Buffer
	w := base64.NewEncoder(base64.StdEncoding, &buf)
	if _, err := io.Copy(w, r); err != nil {
		t.Error(err.Error())
	}
	if err := w.Close(); err != nil {
		t.Error(err.Error())
	}
	s := buf.String()
	if s != "xQ==" {
		t.Errorf("Read: got %v, want xQ==", s)
	}
}

func TestSavedLink(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping going out to the net")
	}
	slink := SavedLink{Link: "https://httpbin.org/bytes/1?seed=0"}
	fn, err := slink.Save()
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("saved link to %v", fn)
	slink.Remove()
	if _, err := os.Stat(fn); err == nil {
		t.Errorf("SavedLink: file exists, but should be deleted: %v", fn)
	}
}

func TestFileReader(t *testing.T) {
	var buf bytes.Buffer
	testfile := "io.go"
	n, err := io.Copy(&buf, &FileReader{Filename: testfile})
	if err != nil {
		t.Error(err.Error())
	}
	f, err := os.Open(testfile)
	if err != nil {
		t.Error(err.Error())
	}
	fi, err := f.Stat()
	if err != nil {
		t.Error(err.Error())
	}
	if n != fi.Size() {
		t.Errorf("FileReader: got %v, want %v", n, fi.Size())
	}
}

func TestZipContentReader(t *testing.T) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, &ZipContentReader{Filename: "../fixtures/z.zip"})
	if err != nil {
		t.Error(err.Error())
	}
	want, got := "a\nb\n", buf.String()
	if want != got {
		t.Errorf("ZipContentReader: got %v, want %v", got, want)
	}
}

func TestSavedReaders(t *testing.T) {
	sr := SavedReaders{Readers: []io.Reader{
		strings.NewReader("Hello"),
		strings.NewReader("World"),
	}}
	fn, err := sr.Save()
	if err != nil {
		t.Error(err.Error())
	}
	b, err := os.ReadFile(fn)
	if err != nil {
		t.Error(err.Error())
	}
	if string(b) != "HelloWorld" {
		t.Errorf("SavedReaders: got %v, want HelloWorld", string(b))
	}
	sr.Remove()
	if _, err := os.Stat(fn); err == nil {
		t.Errorf("SavedReaders: file exists, but should be deleted: %v", fn)
	}
}
