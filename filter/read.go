package filter

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// readLines returns the trimmed, non-empty lines from r. A last line without a
// trailing newline is not included.
func readLines(r io.Reader) (lines []string, err error) {
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			return lines, nil
		}
		if err != nil {
			return nil, err
		}
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
}

// readFileLines is readLines for a file.
func readFileLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readLines(f)
}

// readLinkOrFile returns the lines of the document at link, or of file, if
// link is empty.
func readLinkOrFile(link, file string) ([]string, error) {
	// span freeze rewrites URLs to file:// URLs, which http.Get does not handle.
	if path, ok := strings.CutPrefix(link, "file://"); ok {
		link, file = "", path
	}
	switch {
	case link != "":
		b, err := fetch(link)
		if err != nil {
			return nil, err
		}
		return readLines(bytes.NewReader(b))
	case file != "":
		return readFileLines(file)
	default:
		return nil, nil
	}
}

// fetch returns the body of a URL.
func fetch(link string) ([]byte, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request to %s failed with: %s", link, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// unzipOrPlain returns the concatenated content of all files in the zip
// archive b, or b itself, if it is not a zip archive.
func unzipOrPlain(b []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return b, nil
	}
	var buf bytes.Buffer
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(&buf, rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
