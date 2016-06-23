package holdings

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var (
	ErrBeforeCoverageInterval = errors.New("before coverage interval")
	ErrAfterCoverageInterval  = errors.New("after coverage interval")
	ErrMissingValues          = errors.New("missing values")
	ErrMovingWall             = errors.New("moving wall")
	ErrUnparsedTime           = errors.New("unparsed time")
)

var (
	intPattern = regexp.MustCompile("[0-9]+")
)

// ParseError collects unmarshal errors.
type ParseError struct {
	Errors []error
}

// Error returns the number of errors encountered.
func (e ParseError) Error() string {
	return fmt.Sprintf("%d error(s) detected", len(e.Errors))
}

// Holdings can return a list of licenses for a given ISSN.
type Holdings interface {
	Licenses(string) []License
}

// License exposes methods to let clients check their own validity in the
// context of this license.
// TODO(miku): Only once method is required: Covers.
type License interface {
	// Covers looks at the static content only.
	Covers(Signature) error
}

// File is the interface implemented by all holdings file formats. It can return
// a list of entries.
type File interface {
	ReadEntries() (Entries, error)
}

// Entries holds a list of license entries keyed by ISSN. The ISSN should be
// written as 1234-567X, according to the standard (https://en.wikipedia.org/wik
// i/International_Standard_Serial_Number#Code_format). The format of the ISSN
// is an eight digit code, divided by a hyphen into two four-digit numbers.
// Not an interface since it should not be more complicated than this.
type Entries map[string][]License

// Licenses make Entries fulfill the holdings interface.
func (e Entries) Licenses(issn string) []License {
	return e[issn]
}

// Entry is a reduced holding file entry. Usually, moving wall allow the
// items, that lie earlier than the boundary. If EmbargoDisallowEarlier is
// set, the effect is reversed.
type Entry struct {
	Begin                  Signature
	End                    Signature
	Embargo                time.Duration
	EmbargoDisallowEarlier bool
}

// TimeRestricted returns an error, if the given time falls within the moving
// wall set by the Entry. The embargo is simply added to the current time,
// so it should expressed with negative values.
func (e Entry) TimeRestricted(t time.Time) error {
	if e.Embargo == 0 {
		return nil
	}
	var now = time.Now()
	if e.EmbargoDisallowEarlier {
		if t.Before(now.Add(e.Embargo)) {
			return ErrMovingWall
		}
	} else {
		if t.After(now.Add(e.Embargo)) {
			return ErrMovingWall
		}
	}
	return nil
}

// Covers returns true if the given signature lies inside the interval defined
// by the entry and the moving wall is not hit. If there is not comparable
// date, the volume and issue comparisons do not make much sense. However, if
// there is a date, we are ok with just one of volume or issue defined.
func (e Entry) Covers(s Signature) error {
	if err := e.compareDate(s); err != nil {
		return err
	}
	if err := e.compareVolume(s); err != nil {
		switch err {
		case ErrMissingValues:
			// allow missing volume
		default:
			return err
		}
	}
	if err := e.compareIssue(s); err != nil {
		switch err {
		case ErrMissingValues:
			// allow missing issue
		default:
			return err
		}
	}
	t, err := s.Time()
	if err != nil {
		return err
	}
	return e.TimeRestricted(t)
}

// compareYear returns an error, if both values are defined and disagree, or
// if too few values are defined to do a sane comparison.
func (e Entry) compareDate(s Signature) error {
	if s.Date == "" || (e.Begin.Date == "" && e.End.Date == "") {
		return ErrMissingValues
	}
	if e.Begin.Date != "" {
		if s.Date < e.Begin.Date {
			return ErrBeforeCoverageInterval
		}
	}
	if e.End.Date != "" {
		if s.Date > e.End.Date {
			return ErrAfterCoverageInterval
		}
	}
	return nil
}

// compareVolume returns an error, if both values are defined and disagree,
// otherwise we assume there is no error.
func (e Entry) compareVolume(s Signature) error {
	if s.Volume == "" || (e.Begin.Volume == "" && e.End.Volume == "") {
		return ErrMissingValues
	}
	if e.Begin.Volume != "" {
		if s.VolumeInt() < e.Begin.VolumeInt() {
			return ErrBeforeCoverageInterval
		}
	}
	if e.End.Volume != "" {
		if s.VolumeInt() > e.End.VolumeInt() {
			return ErrAfterCoverageInterval
		}
	}
	return nil
}

// compareIssue returns an error, if both values are defined and disagree,
// otherwise we assume there is no error.
func (e Entry) compareIssue(s Signature) error {
	if s.Issue == "" || (e.Begin.Issue == "" && e.End.Issue == "") {
		return nil
	}
	if e.Begin.Issue != "" {
		if s.IssueInt() < e.Begin.IssueInt() {
			return ErrBeforeCoverageInterval
		}
	}
	if e.End.Issue != "" {
		if s.IssueInt() > e.End.IssueInt() {
			return ErrAfterCoverageInterval
		}
	}
	return nil
}

// Signature is a bag of information of the record from which coverage can be
// determined. Date should comparable strings, like 2010 or 2010-12-21. The
// volume and issue are parsed as ints, with noise allowed, like "Vol. 1".
type Signature struct {
	Date   string
	Volume string
	Issue  string
}

// datePatterns are candidate patterns for parsing publishing dates.
var datePatterns = []string{
	"2006",
	"2006-",
	"2006-1",
	"2006-01",
	"2006-1-2",
	"2006-1-02",
	"2006-01-2",
	"2006-01-02",
	"2006-Jan",
	"2006-January",
	"2006-Jan-2",
	"2006-Jan-02",
	"2006-January-2",
	"2006-January-02",
	"2006-x",
	"2006-xx",
	"2006-x-x",
	"2006-x-xx",
	"2006-xx-x",
	"2006-xx-xx",
}

// Time returns a time value, if it is possible to parse it. Otherwise
// ErrUnparsedTime is returned. Extend datePatterns if necessary.
func (s Signature) Time() (time.Time, error) {
	for _, layout := range datePatterns {
		t, err := time.Parse(layout, s.Date)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrUnparsedTime
}

// VolumeInt returns the Volume in a best effort manner.
func (s Signature) VolumeInt() int {
	return findInt(s.Volume)
}

// IssueInt returns the issue as int in a best effort manner.
func (s Signature) IssueInt() int {
	return findInt(s.Issue)
}

// findInt return the first int that is found in s or 0 if there is no number.
func findInt(s string) int {
	// we expect to see a number most of the time
	if i, err := strconv.Atoi(s); err == nil {
		return int(i)
	}
	// otherwise try to parse out a number
	if m := intPattern.FindString(s); m == "" {
		return 0
	} else {
		i, _ := strconv.ParseInt(m, 10, 32)
		return int(i)
	}
}
