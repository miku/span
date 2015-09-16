package span

import (
	"bufio"
	"html"
	"io"
	"log"
	"strings"

	"github.com/rainycape/cld2"
	"golang.org/x/text/language"
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
		if _, err := f.Write(b[:]); err != nil {
			log.Fatal(err)
		}
		if _, err := f.Write([]byte("\n")); err != nil {
			log.Fatal(err)
		}
	}
	if err := f.Flush(); err != nil {
		log.Fatal(err)
	}
	done <- true
}

// DetectLang3 returns the best guess 3-letter language code for a given text.
func DetectLang3(text string) (string, error) {
	c := cld2.Detect(text)
	b, err := language.ParseBase(c)
	if err != nil {
		return "", err
	}
	return b.ISO3(), nil
}
