package span

import (
	"archive/zip"
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

// LinkReader implements io.Reader for a URL.
type LinkReader struct {
	Link string
	buf  bytes.Buffer
	once sync.Once
}

// fill copies the content of the URL into the internal buffer.
func (r *LinkReader) fill() (err error) {
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
	if err := r.fill(); err != nil {
		return 0, err
	}
	return r.buf.Read(p)
}

// SavedLink saves the content of a URL to a file.
type SavedLink struct {
	Link string
	f    *os.File
}

// Save link to a temporary file, return the filename.
func (s *SavedLink) Save() (filename string, err error) {
	r := &LinkReader{Link: s.Link}
	s.f, err = ioutil.TempFile("", "span-")
	if err != nil {
		return
	}
	defer s.f.Close()
	if _, err = io.Copy(s.f, r); err != nil {
		return
	}
	return s.f.Name(), nil
}

// Remove remove any left over temporary file.
func (s *SavedLink) Remove() {
	_ = os.Remove(s.f.Name())
}

// ZipContentReader returns the concatenated content of all files in a zip archive
// given by its filename. All content is temporarily stored in memory, so this
// type should only be used with smaller archives.
type ZipContentReader struct {
	Filename string
	buf      bytes.Buffer
	once     sync.Once
}

// fill populates the internal buffer with the content of all archive members.
func (r *ZipContentReader) fill() (err error) {
	r.once.Do(func() {
		var zrc *zip.ReadCloser
		if zrc, err = zip.OpenReader(r.Filename); err != nil {
			return
		}
		defer zrc.Close()

		for _, f := range zrc.File {
			var rc io.ReadCloser
			if rc, err = f.Open(); err != nil {
				return
			}
			if _, err = io.Copy(&r.buf, rc); err != nil {
				return
			}
			err = rc.Close()
		}
	})
	return
}

// Read returns the content of all archive members.
func (r *ZipContentReader) Read(p []byte) (int, error) {
	if err := r.fill(); err != nil {
		return 0, err
	}
	return r.buf.Read(p)
}

// FileReader creates a ReadCloser from a filename. If postpones error handling up
// until the first read.
type FileReader struct {
	Filename string
	f        *os.File
}

func (r *FileReader) openFile() (err error) {
	if r.f != nil {
		return nil
	}
	r.f, err = os.Open(r.Filename)
	return err
}

// Read reads from the file.
func (r *FileReader) Read(p []byte) (n int, err error) {
	if err = r.openFile(); err != nil {
		return
	}
	n, err = r.f.Read(p)
	if err == io.EOF {
		defer r.f.Close()
		r.f = nil
	}
	return
}

// Close closes the file.
func (r *FileReader) Close() (err error) {
	if r.f != nil {
		err = r.f.Close()
		r.f = nil
	}
	return
}

// ZipOrPlainLinkReader is a reader that transparently handles zipped and uncompressed
// content, given a URL as string.
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
		if filename, err = link.Save(); err != nil {
			return
		}
		defer link.Remove()

		zipReader := &ZipContentReader{Filename: filename}
		if _, err = io.Copy(&r.buf, zipReader); err == nil {
			// If there is no error with zip, assume it was a zip and return.
			return
		}
		// Error with zip? Return plain content.
		_, err = io.Copy(&r.buf, &FileReader{Filename: filename})
	})
	return err
}

// Read implements the reader interface.
func (r *ZipOrPlainLinkReader) Read(p []byte) (int, error) {
	if err := r.fill(); err != nil {
		return 0, err
	}
	return r.buf.Read(p)
}

// SavedReaders takes a list of readers and persists their content in temporary file.
// SavedReaders takes a list of readers and persists their content in a temporary file.
type SavedReaders struct {
	Readers []io.Reader
	f       *os.File
}

// Save saves all readers to a temporary file and returns the filename.
func (r *SavedReaders) Save() (filename string, err error) {
	r.f, err = ioutil.TempFile("", "span-")
	if err != nil {
		return
	}
	if _, err = io.Copy(r.f, io.MultiReader(r.Readers...)); err != nil {
		return
	}
	if err = r.f.Close(); err != nil {
		return
	}
	filename = r.f.Name()
	return
}

// Remove remove any left over temporary file.
func (r *SavedReaders) Remove() {
	_ = os.Remove(r.f.Name())
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

// WriteCounter counts the number of bytes written through it.
type WriteCounter struct {
	w     io.Writer
	count uint64
}

// Write increments the total byte count.
func (w *WriteCounter) Write(p []byte) (int, error) {
	atomic.AddUint64(&w.count, uint64(len(p)))
	return len(p), nil
}

// Count returns the number of bytes written.
func (w *WriteCounter) Count() uint64 {
	return atomic.LoadUint64(&w.count)
}
