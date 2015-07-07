package holdings

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// Inaccurate but sufficient helpers for unwrapping delay values.
	day   = 24 * time.Hour
	month = 30 * day
	year  = 12 * month

	// LowDatum16 represents a lowest datum for unspecified start dates.
	// The format is YYYYvvvvvviiiiii (year-volume-issue, zero-padded).
	LowDatum16 = "0000000000000000"

	// HighDatum16 represents a lowest datum for unspecified end dates.
	// The format is YYYYvvvvvviiiiii (year-volume-issue, zero-padded).
	HighDatum16 = "ZZZZZZZZZZZZZZZZ"

	// MaxVolume is the largest year we can sensibly handle.
	MaxYear = "9999"

	// MaxVolume is the largest volume number we can sensibly handle.
	MaxVolume = "999999"

	// MaxVolume is the largest issue number we can sensibly handle.
	MaxIssue = "999999"

	// emptyLicense
	emptyLicense = License("")
)

// delayPattern is how moving walls are expressed in OVID.
var delayPattern = regexp.MustCompile(`^([-+]\d+)(M|Y)$`)

var (
	errUnknownUnit   = errors.New("unknown unit")
	errUnknownFormat = errors.New("unknown format")
	errDelayMismatch = errors.New("delay mismatch")
	errInvalidYear   = errors.New("invalid year")
	errVolumeTooBig  = errors.New("volume number too big")
	errIssueTooBig   = errors.New("issue number too big")
	errInvalidRange  = errors.New("invalid range in holdings file")
)

// ISSNPattern is the canonical form of an ISSN.
var ISSNPattern = regexp.MustCompile(`^\d\d\d\d-\d\d\d\d$`)

// Holding contains a single holding.
type Holding struct {
	EZBID        int           `xml:"ezb_id,attr" json:"ezbid"`
	Title        string        `xml:"title" json:"title"`
	Publishers   string        `xml:"publishers" json:"publishers"`
	PISSN        []string      `xml:"EZBIssns>p-issn" json:"pissn"`
	EISSN        []string      `xml:"EZBIssns>e-issn" json:"eissn"`
	Entitlements []Entitlement `xml:"entitlements>entitlement" json:"entitlements"`
}

// Entitlement holds a single OVID entitlement.
type Entitlement struct {
	Status     string `xml:"status,attr" json:"status"`
	URL        string `xml:"url" json:"url"`
	Anchor     string `xml:"anchor" json:"anchor"`
	FromYear   string `xml:"begin>year" json:"from-year"`
	FromVolume string `xml:"begin>volume" json:"from-volume"`
	FromIssue  string `xml:"begin>issue" json:"from-issue"`
	FromDelay  string `xml:"begin>delay" json:"from-delay"`
	ToYear     string `xml:"end>year" json:"to-year"`
	ToVolume   string `xml:"end>volume" json:"to-volume"`
	ToIssue    string `xml:"end>issue" json:"to-issue"`
	ToDelay    string `xml:"end>delay" json:"to-delay"`
}

// License represents a span of time, which a license covers, expressed as a
// string of the form `from:to:delay`. Both `from` and `to` are expressed as
// `YYYYvvvvvviiiiii` (year-volume-issue, zero-padded). Since we resort to
// string comparisons, `0000000000000000` and `ZZZZZZZZZZZZZZZZ` are valid
// values for unbounded start and end points in time. The delay must be
// expressed in nanoseconds, e.g. -2Y would be expressed as
// `-62208000000000000`.
type License string

// From returns the start of the license range.
func (l License) From() string {
	parts := strings.Split(string(l), ":")
	return parts[0]
}

// To returns the end of the license range.
func (l License) To() string {
	parts := strings.Split(string(l), ":")
	return parts[1]
}

// Covers returns true, if the given signature falls between the start and end
// of the license. Moving wall does not play a role here.
func (l License) Covers(signature string) bool {
	return signature >= l.From() && l.To() >= signature
}

// Delay returns the delay as a duration. This function will halt the world if
// the license has not passed basic sanity checks. Always use
// `NewLicenseFromEntitlement` to build a license.
func (l License) Delay() time.Duration {
	parts := strings.Split(string(l), ":")
	v, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Fatal(err)
	}
	return time.Duration(v)
}

// Wall returns the licence wall truncated to day. The moving wall calculation
// is based on the time given in ref.
func (l License) Wall(ref time.Time) time.Time {
	d := time.Duration(-ref.Hour()) * time.Hour
	return ref.Truncate(time.Hour).Add(d + l.Delay())
}

// Licenses holds the license ranges for an ISSN.
type Licenses map[string][]License

// Add adds a license range string to a given ISSN. Dups are ignored.
func (t Licenses) Add(issn string, license License) {
	for _, v := range t[issn] {
		if v == license {
			return
		}
	}
	t[issn] = append(t[issn], license)
}

// NewLicenseFromEntitlement creates a simple License string from the more
// complex Entitlement structure. If error is nil, the License passed the
// sanity checks.
func NewLicenseFromEntitlement(e Entitlement) (License, error) {
	if len(e.FromYear) > len(MaxYear) || len(e.ToYear) > len(MaxYear) {
		return emptyLicense, errInvalidYear
	}
	if len(e.FromVolume) > len(MaxVolume) || len(e.ToVolume) > len(MaxVolume) {
		return emptyLicense, errVolumeTooBig
	}
	if len(e.FromIssue) > len(MaxIssue) || len(e.ToIssue) > len(MaxIssue) {
		return emptyLicense, errIssueTooBig
	}

	from := CombineDatum(e.FromYear, e.FromVolume, e.FromIssue, LowDatum16)
	to := CombineDatum(e.ToYear, e.ToVolume, e.ToIssue, HighDatum16)

	if to < from {
		return emptyLicense, errInvalidRange
	}

	delay := firstNonemptyString(e.FromDelay, e.ToDelay, "-0M")

	dur, err := parseDelay(delay)
	if err != nil {
		return emptyLicense, err
	}

	return License(fmt.Sprintf("%s:%s:%d", from, to, dur.Nanoseconds())), nil
}

// CombineDatum combines year, volume and issue into a single value,
// that preserves the order, if length of year, volume and issue do not
// exceed 4, 6 and 6, respectively: YYYYvvvvvviiiiii.
func CombineDatum(year, volume, issue string, empty string) string {
	if year == "" && volume == "" && issue == "" && empty != "" {
		return empty
	}
	return fmt.Sprintf("%04s%06s%06s", year, volume, issue)
}

// parseDelay parses delay strings like '-1M', '-3Y', ... into a time.Duration.
// Will fail on on units other that M and Y.
func parseDelay(s string) (time.Duration, error) {
	var d time.Duration
	if s == "" {
		return d, nil
	}
	ms := delayPattern.FindStringSubmatch(s)
	if len(ms) != 3 {
		return d, errUnknownFormat
	}
	value, err := strconv.Atoi(ms[1])
	if err != nil {
		return d, err
	}
	switch {
	case ms[2] == "Y":
		return time.Duration(time.Duration(value) * year), nil
	case ms[2] == "M":
		return time.Duration(time.Duration(value) * month), nil
	default:
		return d, errUnknownUnit
	}
}

// firstNonemptyString returns the first value that is not the empty string.
func firstNonemptyString(v ...string) string {
	for _, s := range v {
		if s != "" {
			return s
		}
	}
	return ""
}

// ParseHoldings takes a reader and will try to return Licenses, which is just
// a map from ISSN to []License. Errors are collected and returned as slice.
func ParseHoldings(r io.Reader) (Licenses, []error) {
	decoder := xml.NewDecoder(bufio.NewReader(r))
	lmap := make(Licenses)
	var errors []error
	var tag string

	for {
		t, err := decoder.Token()
		if err != nil && err != io.EOF {
			errors = append(errors, err)
			return lmap, errors
		}
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			tag = se.Name.Local
			if tag == "holding" {
				var item Holding
				decoder.DecodeElement(&item, &se)
				var hls []License
				for _, e := range item.Entitlements {
					l, err := NewLicenseFromEntitlement(e)
					if err != nil {
						errors = append(errors, fmt.Errorf("%d => %s", item.EZBID, err))
					} else {
						hls = append(hls, l)
					}
				}

				for _, issn := range append(item.EISSN, item.PISSN...) {
					for _, l := range hls {
						lmap.Add(issn, l)
					}
				}
			}
		}
	}
	return lmap, errors
}
