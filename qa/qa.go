// Package qa implements quality assurance helpers.
package qa

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

var (
	// EarliestDate is the earliest publication date we accept.
	EarliestDate = time.Date(1458, 1, 1, 0, 0, 0, 0, time.UTC)
	// LatestDate represents the latest publication date we accept. Five years into the future.
	LatestDate = time.Now().AddDate(5, 0, 0)

	ErrInvalidEndPage          = errors.New("broken end page")
	ErrInvalidStartPage        = errors.New("broken start page")
	ErrEndPageBeforeStartPage  = errors.New("end page before start page")
	ErrSuspiciousPageCount     = errors.New("suspicious page count")
	ErrInvalidURL              = errors.New("invalid URL")
	ErrKeyTooLong              = fmt.Errorf("record id exceeds key limit of %s", span.KeyLengthLimit)
	ErrPublicationDateTooEarly = errors.New("publication date too early")
	ErrRepeatedSubtitle        = errors.New("repeated subtitle")
	ErrCurrencyInTitle         = errors.New("currenty in title")
	ErrExcessivePunctuation    = errors.New("excessive punctuation")
	ErrNoPublisher             = errors.New("no publisher")
	ErrShortAuthorName         = errors.New("very short author name")
	ErrEtAlAuthorName          = errors.New("et al in author name")
	ErrNAInAuthorName          = errors.New("NA in author name")
	ErrWhitespaceAuthor        = errors.New("whitespace author")
	ErrHTMLEntityInAuthorName  = errors.New("html entity in author name")
	ErrRepeatedSlashInDOI      = errors.New("repeated slash in DOI")
	ErrNoURL                   = errors.New("record has no URL")
	ErrNonCanonicalISSN        = errors.New("non-canonical ISSN")

	// currencyPattern is a rather narrow pattern:
	// http://rubular.com/r/WjcnjhckZq, used by NoCurrencyInTitle
	currencyPattern = regexp.MustCompile(`[€$¥][+-]?[0-9]{1,3}(?:[0-9]*(?:[.,][0-9]{2})?|(?:,[0-9]{3})*(?:\.[0-9]{2})?|(?:\.[0-9]{3})*(?:,[0-9]{2})?)`)
	// suspiciousPatterns, used by NoExcessivePunctuation
	suspiciousPatterns = []string{"?????", "!!!!!", "....."}
	// issnPattern looks for a canonical ISSN
	issnPattern = regexp.MustCompile(`^[0-9]{4,4}-[0-9]{3,3}[0-9X]$`)
	// htmlEntityPattern looks for leftover entities: http://rubular.com/r/flzmBzpShX
	htmlEntityPattern = regexp.MustCompile(`&(?:[a-z\d]+|#\d+|#x[a-f\d]+);`)
)

var TestSuite = []Tester{
	TesterFunc(TestKeyLength),
	TesterFunc(TestPageCount),
	TesterFunc(TestURL),
	TesterFunc(TestDate),
	TesterFunc(TestSubtitleRepetition),
	TesterFunc(TestCurrencyInTitle),
	TesterFunc(TestExcessivePunctuation),
	TesterFunc(TestPublisher),
	TesterFunc(TestFeasibleAuthor),
	TesterFunc(TestRepeatedSlashInDOI),
	TesterFunc(TestHasURL),
	TesterFunc(TestCanonicalISSN),
}

// Tester is a intermediate record checker.
type Tester interface {
	TestRecord(finc.IntermediateSchema) error
}

// TesterFunc makes a function satisfy an interface.
type TesterFunc func(finc.IntermediateSchema) error

// TestRecord delegates test to the given func.
func (f TesterFunc) TestRecord(is finc.IntermediateSchema) error {
	return f(is)
}

type Issue struct {
	Err    error                   `json:"err"`
	Record finc.IntermediateSchema `json:"record"`
}

func (i Issue) Error() string {
	return fmt.Sprintf("%s: %+v", i.Error, i.Record)
}

func (i Issue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"err":    i.Err.Error(),
		"record": i.Record,
	})
}

// TestKeyLength checks the length of the record id. memcachedb limits is 250 bytes.
func TestKeyLength(is finc.IntermediateSchema) error {
	if len(is.RecordID) > span.KeyLengthLimit {
		return Issue{Err: ErrKeyTooLong, Record: is}
	}
	return nil
}

// TestURL check if URLs are in order.
func TestURL(is finc.IntermediateSchema) error {
	for _, s := range is.URL {
		if _, err := url.Parse(s); err != nil {
			return Issue{Err: ErrInvalidURL, Record: is}
		}
	}
	return nil
}

// TestDate checks for suspicious dates, refs. #5686.
func TestDate(is finc.IntermediateSchema) error {
	if is.Date.Before(EarliestDate) {
		return Issue{Err: ErrPublicationDateTooEarly, Record: is}
	}
	if is.Date.After(LatestDate) {
		return Issue{Err: ErrPublicationDateTooEarly, Record: is}
	}
	return nil
}

// TestPageCount checks, wether the start and end page look plausible.
func TestPageCount(is finc.IntermediateSchema) error {
	const (
		maxPageDigits = 6
		maxPageCount  = 20000
	)
	if len(is.StartPage) > maxPageDigits {
		return Issue{Err: ErrInvalidStartPage, Record: is}
	}
	if len(is.EndPage) > maxPageDigits {
		return Issue{Err: ErrInvalidEndPage, Record: is}
	}
	if is.StartPage != "" && is.EndPage != "" {
		if s, err := strconv.Atoi(is.StartPage); err == nil {
			if e, err := strconv.Atoi(is.EndPage); err == nil {
				if e < s {
					return Issue{Err: ErrEndPageBeforeStartPage, Record: is}
				}
				if e-s > maxPageCount {
					return Issue{Err: ErrSuspiciousPageCount, Record: is}
				}
			} else {
				return Issue{Err: ErrInvalidEndPage, Record: is}
			}
		} else {
			return Issue{Err: ErrInvalidStartPage, Record: is}
		}
	}
	return nil
}

// TestSubtitleRepetition, refs #6553.
func TestSubtitleRepetition(is finc.IntermediateSchema) error {
	if is.ArticleSubtitle != "" && strings.Contains(is.ArticleTitle, is.ArticleSubtitle) {
		return Issue{Err: ErrRepeatedSubtitle, Record: is}
	}
	return nil
}

// TestCurrencyInTitle, e.g. http://goo.gl/HACBcW
// Cartier , Marie . Baby, You Are My Religion: Women, Gay Bars, and Theology
// Before Stonewall . Gender, Theology and Spirituality. Durham, UK: Acumen,
// 2013. xii+256 pp. $90.00 (cloth); $29.95 (paper).
func TestCurrencyInTitle(is finc.IntermediateSchema) error {
	if currencyPattern.MatchString(is.ArticleTitle) {
		return Issue{Err: ErrCurrencyInTitle, Record: is}
	}
	return nil
}

// TestExcessivePunctuation should detect things like this title:
// CrossRef????????????? https://goo.gl/AD0V1o
func TestExcessivePunctuation(is finc.IntermediateSchema) error {
	for _, p := range suspiciousPatterns {
		if strings.Contains(is.ArticleTitle, p) {
			return Issue{Err: ErrExcessivePunctuation, Record: is}
		}
	}
	return nil
}

// TestPublisher tests, whether a publisher is given.
func TestPublisher(is finc.IntermediateSchema) error {
	// This rule is does not apply for this source.
	if is.SourceID == "48" {
		return nil
	}
	switch len(is.Publishers) {
	case 0:
		return Issue{Err: ErrNoPublisher, Record: is}
	case 1:
		if is.Publishers[0] == "" {
			return Issue{Err: ErrNoPublisher, Record: is}
		}
	default:
		for _, p := range is.Publishers {
			if p == "" {
				return Issue{Err: ErrNoPublisher, Record: is}
			}
		}
	}
	return nil
}

// TestFeasibleAuthor checks for a few suspicious authors patterns, refs. #4892, #4940, #5895.
func TestFeasibleAuthor(is finc.IntermediateSchema) error {
	for _, author := range is.Authors {
		s := author.String()
		if len(s) < 5 {
			return Issue{Err: ErrShortAuthorName, Record: is}
		}
		lower := strings.ToLower(s)
		if strings.HasPrefix(lower, "et al") {
			return Issue{Err: ErrEtAlAuthorName, Record: is}
		}
		if strings.Contains(lower, "&na;") {
			return Issue{Err: ErrNAInAuthorName, Record: is}
		}
		if len(s) > 0 && strings.TrimSpace(s) == "" {
			return Issue{Err: ErrWhitespaceAuthor, Record: is}
		}
		if htmlEntityPattern.MatchString(s) {
			return Issue{Err: ErrHTMLEntityInAuthorName, Record: is}
		}
	}
	return nil
}

// TestRepeatedSlashInDOI checks a DOI for repeated slashes, refs. #6312.
func TestRepeatedSlashInDOI(is finc.IntermediateSchema) error {
	if strings.Contains(is.DOI, "//") {
		return Issue{Err: ErrRepeatedSlashInDOI, Record: is}
	}
	return nil
}

// TestHasURL checks for a value in URL. This is no URL validation.
func TestHasURL(is finc.IntermediateSchema) error {
	if len(is.URL) == 0 {
		return Issue{Err: ErrNoURL, Record: is}
	}
	return nil
}

// TestCanonicalISSN checks for the canonical ISSN format 1234-567X.
func TestCanonicalISSN(is finc.IntermediateSchema) error {
	for _, issn := range append(is.ISSN, is.EISSN...) {
		if !issnPattern.MatchString(issn) {
			return Issue{Err: ErrNonCanonicalISSN, Record: is}
		}
	}
	return nil
}
