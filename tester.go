package span

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span/assetutil"
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
	PublicationDateTooEarly
	PublicationDateTooLate
	InvalidCollection
	RepeatedSubtitle
	CurrencyInTitle
)

var (
	// EarliestDate is the earliest publication date we accept.
	EarliestDate = time.Date(1458, 1, 1, 0, 0, 0, 0, time.UTC)
	// LatestDate represents the latest publication date we accept.
	LatestDate = time.Now().AddDate(5, 0, 0)

	// AllowedCollections
	AllowedCollections = assetutil.MustLoadStringSet("assets/qc/collections/collections.tsv", "assets/qc/collections/crossref.tsv")
)

type QualityIssue struct {
	Kind    Kind
	Record  finc.IntermediateSchema
	Message string
}

func (e QualityIssue) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Record.RecordID, e.Kind, e.Message)
}

func (e QualityIssue) TSV() string {
	return fmt.Sprintf("%s\t%s\t%s", e.Record.RecordID, e.Kind, e.Message)
}

var DefaultTests = []RecordTester{
	RecordTesterFunc(KeyLength),
	RecordTesterFunc(PlausiblePageCount),
	RecordTesterFunc(ValidURL),
	RecordTesterFunc(PlausibleDate),
	RecordTesterFunc(AllowedCollectionNames),
	RecordTesterFunc(SubtitleRepetition),
	RecordTesterFunc(NoCurrencyInTitle),
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

func PlausibleDate(is finc.IntermediateSchema) error {
	if is.Date.Before(EarliestDate) {
		return QualityIssue{Kind: PublicationDateTooEarly, Record: is, Message: is.Date.String()}
	}
	if is.Date.After(LatestDate) {
		return QualityIssue{Kind: PublicationDateTooLate, Record: is, Message: is.Date.String()}
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

func AllowedCollectionNames(is finc.IntermediateSchema) error {
	if !AllowedCollections.Contains(is.MegaCollection) {
		return QualityIssue{Kind: InvalidCollection, Record: is, Message: is.MegaCollection}
	}
	return nil
}

// SubtitleRepetition, refs #6553
func SubtitleRepetition(is finc.IntermediateSchema) error {
	if strings.Contains(is.ArticleTitle, is.ArticleSubtitle) {
		return QualityIssue{Kind: RepeatedSubtitle, Record: is,
			Message: fmt.Sprintf("%s: %s", is.ArticleTitle, is.ArticleSubtitle)}
	}
	return nil
}

// currencyPattern is a rather narrow pattern: http://rubular.com/r/WjcnjhckZq
var currencyPattern = regexp.MustCompile(`[€$¥][+-]?[0-9]{1,3}(?:[0-9]*(?:[.,][0-9]{2})?|(?:,[0-9]{3})*(?:\.[0-9]{2})?|(?:\.[0-9]{3})*(?:,[0-9]{2})?)`)

// NoCurrencyInTitle, e.g. http://katalogbeta.slub-
// dresden.de/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4Ni82ODExNDk/#tx_find
// Cartier , Marie . Baby, You Are My Religion: Women, Gay Bars, and Theology
// Before Stonewall . Gender, Theology and Spirituality. Durham, UK: Acumen,
// 2013. xii+256 pp. $90.00 (cloth); $29.95 (paper).
func NoCurrencyInTitle(is finc.IntermediateSchema) error {
	if currencyPattern.MatchString(is.ArticleTitle) {
		return QualityIssue{Kind: CurrencyInTitle, Record: is, Message: is.ArticleTitle}
	}
	return nil
}
