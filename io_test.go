package span

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"
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
		t.Errorf(err.Error())
	}
	if err := w.Close(); err != nil {
		t.Errorf(err.Error())
	}
	s := buf.String()
	if s != "2A==" {
		t.Errorf("Read: got %v, want 2A==", s)
	}
}

func TestSavedLink(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping going out to the net")
	}
	slink := SavedLink{Link: "https://httpbin.org/bytes/1?seed=0"}
	fn, err := slink.Save()
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Logf("saved link to %v", fn)
	if err := slink.Remove(); err != nil {
		t.Errorf(err.Error())
	}
	if _, err := os.Stat("/path/to/whatever"); err == nil {
		t.Errorf("SavedLink: file exists, but should be deleted: %v", fn)
	}
}

func TestFileReader(t *testing.T) {
	var buf bytes.Buffer
	testfile := "io.go"
	n, err := io.Copy(&buf, &FileReader{Filename: testfile})
	if err != nil {
		t.Errorf(err.Error())
	}
	f, err := os.Open(testfile)
	if err != nil {
		t.Errorf(err.Error())
	}
	fi, err := f.Stat()
	if err != nil {
		t.Errorf(err.Error())
	}
	if n != fi.Size() {
		t.Errorf("FileReader: got %v, want %v", n, fi.Size())
	}
}

func TestZipContentReader(t *testing.T) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, &ZipContentReader{Filename: "fixtures/z.zip"})
	if err != nil {
		t.Errorf(err.Error())
	}
	want, got := "a\nb\n", buf.String()
	if want != got {
		t.Errorf("ZipContentReader: got %v, want %v", got, want)
	}
}
