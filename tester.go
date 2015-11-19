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

func KeyLength(is finc.IntermediateSchema) error {
	if len(is.RecordID) > 250 {
		return QualityIssue{Kind: KeyTooLong, Record: is}
	}
	return nil
}

func ValidURL(is finc.IntermediateSchema) error {
	for _, s := range is.URL {
		if _, err := url.Parse(s); err != nil {
			return QualityIssue{Kind: InvalidURL, Record: is, Message: s}
		}
	}
	return nil
}

func PlausiblePageCount(is finc.IntermediateSchema) error {
	if is.StartPage != "" && is.EndPage != "" {
		if s, err := strconv.Atoi(is.StartPage); err == nil {
			if e, err := strconv.Atoi(is.EndPage); err == nil {
				if e < s {
					return QualityIssue{Kind: EndPageBeforeStartPage, Record: is, Message: fmt.Sprintf("%v-%v", s, e)}
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
