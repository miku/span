package generic

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/miku/span/holdings"
	"github.com/miku/span/holdings/google"
	"github.com/miku/span/holdings/kbart"
	"github.com/miku/span/holdings/ovid"
)

var ErrCannotSniffFormat = errors.New("cannot sniff holding file format")

type ErrorList struct {
	errors []error
}

func (e *ErrorList) Add(err error) {
	e.errors = append(e.errors, err)
}

func (e ErrorList) Error() string {
	var errs []string
	for _, err := range e.errors {
		errs = append(errs, err.Error())
	}
	return fmt.Sprintf("Errors: %v", strings.Join(errs, ", "))
}

// New returns a holding file to read entries from given a filename.
func New(filename string) (holdings.File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return NewReader(bufio.NewReader(file))
}

// NewReader returns a holding file to read entries from. It will probe the
// the different formats and use a suitable implementation.
func NewReader(r io.Reader) (holdings.File, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var file holdings.File
	var errors ErrorList

	if len(b) == 0 {
		log.Printf("warning: file with 0 entries")
		return kbart.NewReader(bytes.NewReader(b)), nil
	}

	// probe kbart
	file = kbart.NewReader(bytes.NewReader(b))
	if entries, err := file.ReadEntries(); err != nil {
		errors.Add(fmt.Errorf("kbart: %s", err))
	} else {
		if len(entries) > 0 {
			log.Printf("holdings: kbart detected (%d)", len(entries))
			return kbart.NewReader(bytes.NewReader(b)), nil
		}
	}

	// probe ovid
	file = ovid.NewReader(bytes.NewReader(b))
	if entries, err := file.ReadEntries(); err != nil {
		errors.Add(fmt.Errorf("ovid: %s", err))
	} else {
		if len(entries) > 0 {
			log.Printf("holdings: ovid detected (%d)", len(entries))
			return ovid.NewReader(bytes.NewReader(b)), nil
		}
	}

	// probe google
	file = google.NewReader(bytes.NewReader(b))
	if entries, err := file.ReadEntries(); err != nil {
		errors.Add(fmt.Errorf("google: %s", err))
	} else {
		if len(entries) > 0 {
			log.Printf("holdings: google detected (%d)", len(entries))
			return google.NewReader(bytes.NewReader(b)), nil
		}
	}

	return file, errors
}
