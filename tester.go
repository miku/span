package span

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/miku/span/finc"
)

//go:generate stringer -type=Kind
type Kind int

const (
	KeyTooLong Kind = iota
	InvalidStartPage
	InvalidEndPage
	EndPageBeforeStartPage
	InvalidURL
	SuspiciousPageCount
)

type QualityIssue struct {
	Kind    Kind
	Record  finc.IntermediateSchema
	Message string
}

func (e QualityIssue) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Record.RecordID, e.Kind, e.Message)
}

var DefaultTests = []RecordTester{
	RecordTesterFunc(KeyLength),
	RecordTesterFunc(PlausiblePageCount),
	RecordTesterFunc(ValidURL),
}

// KeyLength checks the length of the record id.
func KeyLength(is finc.IntermediateSchema) error {
	if len(is.RecordID) > KeyLengthLimit {
		return QualityIssue{Kind: KeyTooLong, Record: is}
	}
	return nil
}

// ValidURL checks, if a URL string is parseable.
func ValidURL(is finc.IntermediateSchema) error {
	for _, s := range is.URL {
		if _, err := url.Parse(s); err != nil {
			return QualityIssue{Kind: InvalidURL, Record: is, Message: s}
		}
	}
	return nil
}

// PlausiblePageCount checks, wether the start and end page look plausible.
func PlausiblePageCount(is finc.IntermediateSchema) error {
	const (
		maxPageDigits = 6
		maxPageCount  = 20000
	)
	if len(is.StartPage) > maxPageDigits {
		return QualityIssue{Kind: InvalidStartPage, Record: is, Message: is.StartPage}
	}
	if len(is.EndPage) > maxPageDigits {
		return QualityIssue{Kind: InvalidEndPage, Record: is, Message: is.EndPage}
	}
	if is.StartPage != "" && is.EndPage != "" {
		if s, err := strconv.Atoi(is.StartPage); err == nil {
			if e, err := strconv.Atoi(is.EndPage); err == nil {
				if e < s {
					return QualityIssue{Kind: EndPageBeforeStartPage, Record: is, Message: fmt.Sprintf("%v-%v", s, e)}
				}
				if e-s > maxPageCount {
					return QualityIssue{Kind: SuspiciousPageCount, Record: is, Message: fmt.Sprintf("%v-%v", s, e)}
				}
			} else {
				return QualityIssue{Kind: InvalidEndPage, Record: is, Message: is.EndPage}
			}
		} else {
			return QualityIssue{Kind: InvalidStartPage, Record: is, Message: is.StartPage}
		}
	}
	return nil
}
