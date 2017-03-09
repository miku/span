package span

import (
	"archive/zip"
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

// LinkReader implements io.Reader for a URL.
type LinkReader struct {
	Link string
	buf  bytes.Buffer
	once sync.Once
}

// fetch fills copies the content of the URL into the internal buffer.
func (r *LinkReader) fetch() (err error) {
	r.once.Do(func() {
		var resp *http.Response
		resp, err = http.Get(r.Link)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		_, err = io.Copy(&r.buf, resp.Body)
	})
	return err
}

func (r *LinkReader) Read(p []byte) (int, error) {
	if err := r.fetch(); err != nil {
		return 0, err
	}
	return r.buf.Read(p)
}

// SavedLink saves the content of a URL to a file.
type SavedLink struct {
	Link string
	f    *os.File
}

// Save link to a temporary file.
func (s *SavedLink) Save() (filename string, err error) {
	r := &LinkReader{Link: s.Link}
	s.f, err = ioutil.TempFile("", "span-")
	if err != nil {
		return "", err
	}
	defer s.f.Close()
	if _, err := io.Copy(s.f, r); err != nil {
		return "", err
	}
	log.Printf("SavedLink: stored %s at %s", s.Link, s.f.Name())
	return s.f.Name(), nil
}

// Remove remove any left over temporary file.
func (s *SavedLink) Remove() error {
	_ = os.Remove(s.f.Name())
	return nil
}

// ZipContentReader returns all files in zip concatenated.
type ZipContentReader struct {
	Filename string
	buf      bytes.Buffer
	once     sync.Once
}

func (r *ZipContentReader) fill() (err error) {
	r.once.Do(func() {
		var rc *zip.ReadCloser
		rc, err = zip.OpenReader(r.Filename)
		if err != nil {
			return
		}
		defer rc.Close()

		for _, f := range rc.File {
			var frc io.ReadCloser
			frc, err = f.Open()
			if err != nil {
				return
			}
			if _, err = io.Copy(&r.buf, frc); err != nil {
				return
			}
			if err = frc.Close(); err != nil {
				return
			}
		}
	})
	return
}

func (r *ZipContentReader) Read(p []byte) (int, error) {
	if err := r.fill(); err != nil {
		return 0, err
	}
	return r.buf.Read(p)
}

// FileReader creates a ReadCloser from a filename.
type FileReader struct {
	Filename string
	f        *os.File
}

// Read reads from the file.
func (r *FileReader) Read(p []byte) (n int, err error) {
	if r.f == nil {
		r.f, err = os.Open(r.Filename)
		if err != nil {
			return
		}
	}
	n, err = r.f.Read(p)
	if err == io.EOF {
		defer r.f.Close()
		r.f = nil
	}
	return
}

// Close closes the file.
func (r *FileReader) Close() error {
	if r.f != nil {
		return r.f.Close()
	}
	return nil
}

// ZipOrPlainLinkReader is a reader that transparently handles zipped and uncompressed
// content. If the content is zipped, all archive entries are extracted.
type ZipOrPlainLinkReader struct {
	Link string
	buf  bytes.Buffer
	once sync.Once
}

// fill fills the internal buffer.
func (r *ZipOrPlainLinkReader) fill() (err error) {
	r.once.Do(func() {
		var filename string
		link := SavedLink{Link: r.Link}
		filename, err = link.Save()
		if err != nil {
			return
		}
		defer link.Remove()

		zipReader := &ZipContentReader{Filename: filename}
		// If there is no error with zip, assume it was a zip and return.
		if _, err = io.Copy(&r.buf, zipReader); err == nil {
			return
		}
		// Error with zip? Return plain content.
		_, err = io.Copy(&r.buf, &FileReader{Filename: filename})
	})
	return err
}

func (r *ZipOrPlainLinkReader) Read(p []byte) (int, error) {
	if err := r.fill(); err != nil {
		return 0, err
	}
	return r.buf.Read(p)
}

// ReadLines returns a list of trimmed lines in a file. Empty lines are skipped.
func ReadLines(filename string) (lines []string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return
}
