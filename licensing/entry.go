// Package licensing implements support for KBART and ISIL attachments.
// KBART might contains special fields, that are important in certain contexts.
// Example: "Aargauer Zeitung" could not be associated with a record, because
// there is no ISSN. However, there is a string
// "https://www.wiso-net.de/dosearch?&dbShortcut=AGZ" in the record, which could
// be parsed to yield "AGZ", which could be used to relate a record to this entry
// (e.g. if the record has "AGZ" in a certain field, like x.package).
package licensing

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/container"
)

// DateGranularity indicates how complete a date is.
type DateGranularity byte

const (
	GRANULARITY_YEAR DateGranularity = iota
	GRANULARITY_MONTH
	GRANULARITY_DAY
)

var (
	ErrBeforeFirstIssueDate = errors.New("before first issue date")
	ErrAfterLastIssueDate   = errors.New("after last issue date")
	ErrBeforeFirstVolume    = errors.New("before first volume")
	ErrAfterLastVolume      = errors.New("after last volume")
	ErrBeforeFirstIssue     = errors.New("before first issue")
	ErrAfterLastIssue       = errors.New("after last issue")
	ErrInvalidDate          = errors.New("invalid date")

	intPattern = regexp.MustCompile("[0-9]+")
)

// dateWithGranularity groups layout and granularity.
type dateWithGranularity struct {
	layout      string
	granularity DateGranularity
}

// datePatterns are candidate patterns for parsing dates.
var datePatterns = []dateWithGranularity{
	{"2006-", GRANULARITY_YEAR},
	{"2006-01-02", GRANULARITY_DAY},
	{"2006-01-02T15:04:05Z", GRANULARITY_DAY},
	{"2006-01-2", GRANULARITY_DAY},
	{"2006-01", GRANULARITY_MONTH},
	{"2006-1-02", GRANULARITY_DAY},
	{"2006-1-2", GRANULARITY_DAY},
	{"2006-1", GRANULARITY_MONTH},
	{"2006-Jan-02", GRANULARITY_DAY},
	{"2006-Jan-2", GRANULARITY_DAY},
	{"2006-Jan", GRANULARITY_MONTH},
	{"2006-January-02", GRANULARITY_DAY},
	{"2006-January-2", GRANULARITY_DAY},
	{"2006-January", GRANULARITY_MONTH},
	{"2006-x-x", GRANULARITY_YEAR},
	{"2006-x-xx", GRANULARITY_YEAR},
	{"2006-x", GRANULARITY_YEAR},
	{"2006-xx-x", GRANULARITY_YEAR},
	{"2006-xx-xx", GRANULARITY_YEAR},
	{"2006-xx", GRANULARITY_YEAR},
	{"2006", GRANULARITY_YEAR},
	{"200601", GRANULARITY_MONTH},
	{"20060102", GRANULARITY_DAY},
}

// Entry contains fields about a licensed or available journal, book, article or
// other resource. First 14 columns are quite stardardized. Further columns may
// contain custom information:
//
// EZB style: own_anchor, package:collection, il_relevance, il_nationwide,
// il_electronic_transmission, il_comment, all_issns, zdb_id
//
// OCLC style: location, title_notes, staff_notes, vendor_id,
// oclc_collection_name, oclc_collection_id, oclc_entry_id, oclc_linkscheme,
// oclc_number, ACTION
//
// See also: http://www.uksg.org/kbart/s5/guidelines/data_field_labels,
// http://www.uksg.org/kbart/s5/guidelines/data_fields
type Entry struct {
	PublicationTitle                   string `csv:"publication_title"`          // "Südost-Forschungen (2014-)", "Theory of Computation"
	PrintIdentifier                    string `csv:"print_identifier"`           // "2029-8692", "9783662479841"
	OnlineIdentifier                   string `csv:"online_identifier"`          // "1533-8606", "9783834960078"
	FirstIssueDate                     string `csv:"date_first_issue_online"`    // "1901", "2008"
	FirstVolume                        string `csv:"num_first_vol_online"`       // "1",
	FirstIssue                         string `csv:"num_first_issue_online"`     // "1"
	LastIssueDate                      string `csv:"date_last_issue_online"`     // "1997", "2008"
	LastVolume                         string `csv:"num_last_vol_online"`        // "25"
	LastIssue                          string `csv:"num_last_issue_online"`      // "1"
	TitleURL                           string `csv:"title_url"`                  // "http://www.karger.com/dne", "http://link.springer.com/10.1007/978-3-658-15644-2"
	FirstAuthor                        string `csv:"first_author"`               // "Borgmann", "Wissenschaftlicher Beirat der Bundesregierung Globale Umweltveränderungen (WBGU)"
	TitleID                            string `csv:"title_id"`                   // "22540", "10.1007/978-3-658-10838-0"
	Embargo                            string `csv:"embargo_info"`               // "P12M", "P1Y", "R20Y"
	CoverageDepth                      string `csv:"coverage_depth"`             // "Volltext", "ebook"
	CoverageNotes                      string `csv:"coverage_notes"`             // ...
	PublisherName                      string `csv:"publisher_name"`             // "via Hein Online", "Springer (formerly: Kluwer)", "DUV"
	OwnAnchor                          string `csv:"own_anchor"`                 // "elsevier_2016_sax", "UNILEIP", "Wiley Custom 2015"
	PackageCollection                  string `csv:"package:collection"`         // "EBSCO:ebsco_bth", "NALAS:natli_aas2", "NALIW:sage_premier"
	InterlibraryRelevance              string `csv:"il_relevance"`               // ...
	InterlibraryNationwide             string `csv:"il_nationwide"`              // ...
	InterlibraryElectronicTransmission string `csv:"il_electronic_transmission"` // "Papierkopie an Endnutzer", "Elektronischer Versand an Endnutzer"
	InterlibraryComment                string `csv:"il_comment"`                 // "Nur im Inland", "il_nationwide"
	AllSerialNumbers                   string `csv:"all_issns"`                  // "1990-0104;1990-0090", "undefined"
	ZDBID                              string `csv:"zdb_id"`                     // "1459367-1" (see also: http://www.zeitschriftendatenbank.de/suche/zdb-katalog.html)
	Location                           string `csv:"location"`                   // ...
	TitleNotes                         string `csv:"title_notes"`                // ...
	StaffNotes                         string `csv:"staff_notes"`                // ...
	VendorID                           string `csv:"vendor_id"`                  // ...
	OCLCCollectionName                 string `csv:"oclc_collection_name"`       // "Springer German Language eBooks 2016 - Full Set", "Wiley Online Library UBCM All Obooks"
	OCLCCollectionID                   string `csv:"oclc_collection_id"`         // "springerlink.de2011fullset", "wiley.ubcmall"
	OCLCEntryID                        string `csv:"oclc_entry_id"`              // "25106066"
	OCLCLinkScheme                     string `csv:"oclc_link_scheme"`           // "wiley.book"
	OCLCNumber                         string `csv:"oclc_number"`                // "122938128"
	Action                             string `csv:"ACTION"`                     // "raw"

	// cache data, that needs to be parsed, for performance
	parsed struct {
		FirstIssueDate time.Time
		LastIssueDate  time.Time
	}
}

// ISSNList returns a list of normalized ISSN from various fields.
func (entry *Entry) ISSNList() []string {
	issns := container.NewStringSet()
	for _, issn := range []string{entry.PrintIdentifier, entry.OnlineIdentifier} {
		s := NormalizeSerialNumber(issn)
		if span.ISSNPattern.MatchString(s) {
			issns.Add(s)
		}
	}
	for _, issn := range FindSerialNumbers(entry.AllSerialNumbers) {
		issns.Add(issn)
	}
	return issns.SortedValues()
}

// Covers is a generic method to determine, whether a given date, volume or
// issue is covered by this entry. It takes into account moving walls. If
// values are not defined, we mostly assume they are not constrained.
func (entry *Entry) Covers(date, volume, issue string) error {
	t, g, err := parseWithGranularity(date)
	if err != nil {
		return err
	}
	// XXX: Containment and embargo should be one thing.
	if err := entry.containsDateTime(t, g); err != nil {
		return err
	}
	if err := Embargo(entry.Embargo).Compatible(t); err != nil {
		return err
	}

	if entry.parsed.FirstIssueDate.Year() == t.Year() {
		if entry.FirstVolume != "" && volume != "" && findInt(volume) < findInt(entry.FirstVolume) {
			return ErrBeforeFirstVolume
		}
		if entry.FirstIssue != "" && issue != "" && findInt(issue) < findInt(entry.FirstIssue) {
			return ErrBeforeFirstIssue
		}
	}

	if entry.parsed.LastIssueDate.Year() == t.Year() {
		if entry.LastVolume != "" && volume != "" && findInt(volume) > findInt(entry.LastVolume) {
			return ErrAfterLastVolume
		}
		if entry.LastIssue != "" && issue != "" && findInt(issue) > findInt(entry.LastIssue) {
			return ErrAfterLastIssue
		}
	}
	return nil
}

// begin parses left boundary of license interval, returns a date far in the
// past if it is not defined.
func (entry *Entry) begin() time.Time {
	if entry.parsed.FirstIssueDate.IsZero() {
		entry.parsed.FirstIssueDate = time.Date(1, time.January, 1, 0, 0, 0, 1, time.UTC)
		for _, dfmt := range datePatterns {
			if t, err := time.Parse(dfmt.layout, entry.FirstIssueDate); err == nil {
				entry.parsed.FirstIssueDate = t
				break
			}
		}
	}
	return entry.parsed.FirstIssueDate
}

// beginGranularity returns the begin date with a given granularity.
func (entry *Entry) beginGranularity(g DateGranularity) time.Time {
	t := entry.begin()
	switch g {
	case GRANULARITY_YEAR:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	case GRANULARITY_MONTH:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		return t
	}
}

// end parses right boundary of license interval, returns a date far in the future
// if it is not defined.
func (entry *Entry) end() time.Time {
	if entry.parsed.LastIssueDate.IsZero() {
		entry.parsed.LastIssueDate = time.Date(2364, time.January, 1, 0, 0, 0, 1, time.UTC)
		for _, dfmt := range datePatterns {
			if t, err := time.Parse(dfmt.layout, entry.LastIssueDate); err == nil {
				entry.parsed.LastIssueDate = t
				break
			}
		}
	}
	return entry.parsed.LastIssueDate
}

// endGranularity returns the end date with a given granularity.
func (entry *Entry) endGranularity(g DateGranularity) time.Time {
	t := entry.end()
	switch g {
	case GRANULARITY_YEAR:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	case GRANULARITY_MONTH:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		return t
	}
}

// containsDateTime returns nil, if the given time lies between this entries
// dates. If the given time is the zero value, it will be contained by any
// interval.
func (entry *Entry) containsDateTime(t time.Time, g DateGranularity) error {
	if t.IsZero() {
		return nil
	}
	if t.Before(entry.beginGranularity(g)) {
		// XXX: This has nothing to do with issue.
		return ErrBeforeFirstIssueDate
	}
	if t.After(entry.endGranularity(g)) {
		return ErrAfterLastIssueDate
	}
	return nil
}

// containsDate return nil, if the given date (as string), lies between this
// entries issue dates. The empty string is interpreted as being inside all
// intervals.
func (entry *Entry) containsDate(s string) (err error) {
	if s == "" {
		return nil
	}
	t, g, err := parseWithGranularity(s)
	if err != nil {
		return err
	}
	return entry.containsDateTime(t, g)
}

// NormalizeSerialNumber tries to transform the input into 1234-567X standard form.
func NormalizeSerialNumber(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if len(s) == 8 {
		return fmt.Sprintf("%s-%s", s[:4], s[4:])
	}
	return s
}

// FindSerialNumbers returns ISSN in standard form in a given string.
func FindSerialNumbers(s string) []string {
	return span.ISSNPattern.FindAllString(s, -1)
}

// parseWithGranularity tries to parse a string into a time. If successful,
// also return the granularity. Any value that is not recorgnized results in an
// ErrInvalidDate.
func parseWithGranularity(s string) (t time.Time, g DateGranularity, err error) {
	if s == "" {
		return time.Time{}, GRANULARITY_DAY, ErrInvalidDate
	}
	for _, dfmt := range datePatterns {
		t, err = time.Parse(dfmt.layout, s)
		if err != nil {
			continue
		}
		g = getGranularity(dfmt.layout)
		return
	}
	return t, g, ErrInvalidDate
}

// getGranularity returns the granularity for given date layout, if nothing
// matches assume the finest granualarity.
func getGranularity(layout string) DateGranularity {
	for _, dfmt := range datePatterns {
		if dfmt.layout == layout {
			return dfmt.granularity
		}
	}
	return GRANULARITY_DAY
}

// findInt return the first int that is found in s or 0 if there is no number.
func findInt(s string) int {
	if s == "" {
		return 0
	}
	// We expect to see a number most of the time.
	if i, err := strconv.Atoi(s); err == nil {
		return int(i)
	}
	// Otherwise try to parse out a number.
	m := intPattern.FindString(s)
	if m == "" {
		return 0
	}
	i, _ := strconv.ParseInt(m, 10, 32)
	return int(i)
}
