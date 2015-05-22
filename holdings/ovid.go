package holdings

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	day   = 24 * time.Hour
	month = 30 * day
	year  = 12 * month

	// LowDatum16 represents a lowest datum for unspecified start dates.
	LowDatum16 = "0000000000000000"
	// HighDatum16 represents a lowest datum for unspecified end dates.
	HighDatum16 = "ZZZZZZZZZZZZZZZZ"
)

// delayPattern is how moving walls are expressed in OVID format.
var delayPattern = regexp.MustCompile(`^([-+]\d+)(M|Y)$`)

var (
	errUnknownUnit   = errors.New("unknown unit")
	errUnknownFormat = errors.New("unknown format")
	errDelayMismatch = errors.New("delay mismatch")
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

// License represents a span of time, which a license covers,
// expressed as a string of the form from:to:delay.
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

func (l License) Covers(signature string) bool {
	if signature < l.From() || l.To() < signature {
		return false
	}
	return true
}

func (l License) Delay() time.Duration {
	parts := strings.Split(string(l), ":")
	v, _ := strconv.Atoi(parts[2])
	return time.Duration(v)

}

// Wall returns the licence wall truncated to day.
func (l License) Wall() time.Time {
	now := time.Now()
	d := time.Duration(-now.Hour()) * time.Hour
	return now.Truncate(time.Hour).Add(d + l.Delay())
}

// NewLicenseFromEntitlement creates a simple License string from the more complex
// Entitlement structure. If error is nil, the License passed sanity checks.
func NewLicenseFromEntitlement(e Entitlement) (License, error) {
	from := CombineDatum(e.FromYear, e.FromVolume, e.FromIssue, LowDatum16)
	to := CombineDatum(e.ToYear, e.ToVolume, e.ToIssue, HighDatum16)
	if to < from {
		return License(""), errors.New("invalid range in holdings file")
	}
	delay := firstNonemptyString(e.FromDelay, e.ToDelay)
	if delay == "" {
		delay = "-0M"
	}
	dur, err := parseDelay(delay)
	if err != nil {
		return License(""), err
	}
	return License(fmt.Sprintf("%s:%s:%d", from, to, dur.Nanoseconds())), nil
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

// CombineDatum combines year, volume and issue into a single value,
// that preserves the order, if length of year, volume and issue do not
// exceed 4, 6 and 6, respectively.
func CombineDatum(year, volume, issue string, empty string) string {
	if year == "" && volume == "" && issue == "" && empty != "" {
		return empty
	}
	return fmt.Sprintf("%04s%06s%06s", year, volume, issue)
}

// parseDelay parses delay strings like '-1M', '-3Y', ... into a time.Duration.
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
	return d, nil
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

func ParseHoldings(r io.Reader) (Licenses, []error) {
	var errors []error
	lmap := make(Licenses)
	decoder := xml.NewDecoder(bufio.NewReader(r))
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
						errors = append(errors, fmt.Errorf("in EZBID(%d): %s", item.EZBID, err))
					}
					hls = append(hls, l)
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

// slimmer processing ----

// IssnHolding maps an ISSN to a holdings.Holding struct.
// ISSN -> Holding -> []Entitlements
type IssnHolding map[string]Holding

// IsilIssnHolding maps an ISIL to an IssnHolding map.
// ISIL -> ISSN -> Holding -> []Entitlements
type IsilIssnHolding map[string]IssnHolding

// Isils returns available ISILs in this IsilIssnHolding map.
func (iih *IsilIssnHolding) Isils() (keys []string) {
	for k := range *iih {
		keys = append(keys, k)
	}
	return keys
}

// // ParseDelay parses delay strings like '-1M', '-3Y', ... into a time.Duration.
// func ParseDelay(s string) (d time.Duration, err error) {
// 	ms := delayPattern.FindStringSubmatch(s)
// 	if len(ms) != 3 {
// 		return d, errUnknownFormat
// 	}
// 	value, err := strconv.Atoi(ms[1])
// 	if err != nil {
// 		return d, err
// 	}
// 	switch {
// 	case ms[2] == "Y":
// 		d = time.Duration(time.Duration(value) * year)
// 	case ms[2] == "M":
// 		d = time.Duration(time.Duration(value) * month)
// 	default:
// 		return d, errUnknownUnit
// 	}
// 	return
// }

// Delay returns the specified delay as `time.Duration`
func (e *Entitlement) Delay() (d time.Duration, err error) {
	if e.FromDelay != "" && e.ToDelay != "" && e.FromDelay != e.ToDelay {
		return d, errDelayMismatch
	}
	if e.FromDelay != "" {
		return parseDelay(e.FromDelay)
	}
	if e.ToDelay != "" {
		return parseDelay(e.ToDelay)
	}
	return
}

// Boundary returns the last date before the moving wall restriction becomes effective.
func (e *Entitlement) Boundary() (d time.Time, err error) {
	delay, err := e.Delay()
	if err != nil {
		return d, err
	}
	return time.Now().Add(delay), nil
}

// HoldingsMap creates an ISSN[Holding] struct from a reader.
func HoldingsMap(reader io.Reader) IssnHolding {
	h := make(map[string]Holding)
	decoder := xml.NewDecoder(reader)
	var tag string
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			tag = se.Name.Local
			if tag == "holding" {
				var item Holding
				decoder.DecodeElement(&item, &se)
				for _, id := range item.EISSN {
					tid := strings.TrimSpace(id)
					if ISSNPattern.MatchString(tid) {
						h[tid] = item
					}
				}
				for _, id := range item.PISSN {
					tid := strings.TrimSpace(id)
					if ISSNPattern.MatchString(tid) {
						h[tid] = item
					}
				}
			}
		}
	}
	return h
}
