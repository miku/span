package span

import (
	"fmt"
	"strconv"

	"github.com/miku/span/finc"
)

//go:generate stringer -type=Kind
type Kind int

const (
	KeyTooLong Kind = iota
	InvalidPage
	EndPageBeforeStartPage
)

type QualityError struct {
	Kind     Kind
	RecordID string
	Message  string
}

func (e QualityError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.RecordID, e.Kind, e.Message)
}

var DefaultTests = []RecordTester{
	RecordTesterFunc(KeyLength),
	RecordTesterFunc(PlausiblePageCount),
}

func KeyLength(is finc.IntermediateSchema) error {
	if len(is.RecordID) > 250 {
		return QualityError{Kind: KeyTooLong, RecordID: is.RecordID, Message: "key too long"}
	}
	return nil
}

func PlausiblePageCount(is finc.IntermediateSchema) error {
	if is.StartPage != "" && is.EndPage != "" {
		if s, err := strconv.Atoi(is.StartPage); err == nil {
			if e, err := strconv.Atoi(is.EndPage); err == nil {
				if e < s {
					return QualityError{Kind: EndPageBeforeStartPage, RecordID: is.RecordID, Message: fmt.Sprintf("%v-%v", s, e)}
				}
			} else {
				return QualityError{Kind: InvalidPage, RecordID: is.RecordID, Message: fmt.Sprintf("end page: %v", e)}
			}
		} else {
			return QualityError{Kind: InvalidPage, RecordID: is.RecordID, Message: fmt.Sprintf("start page: %v", s)}
		}
	}
	return nil
}
