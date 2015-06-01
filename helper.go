package span

import (
	"bufio"
	"html"
	"io"
	"strings"
)

// UnescapeTrim unescapes HTML character references and trims the space of a given string.
func UnescapeTrim(s string) string {
	return strings.TrimSpace(html.UnescapeString(s))
}

// ByteSink is a fan in writer for a byte channel.
// A newline is appended after each object.
func ByteSink(w io.Writer, out chan []byte, done chan bool) {
	f := bufio.NewWriter(w)
	for b := range out {
		f.Write(b[:])
		f.Write([]byte("\n"))
	}
	f.Flush()
	done <- true
}
