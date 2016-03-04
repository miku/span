package ovid

import (
	"bufio"
	"encoding/xml"
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/miku/span/holdings"
)

// delayPattern is how moving walls are expressed in OVID.
var delayPattern = regexp.MustCompile(`^([-+]\d+)(M|Y)$`)

var (
	Day   = 24 * time.Hour
	Month = 30 * Day
	Year  = 12 * Month
)

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

type Reader struct {
	r io.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r)}
}

// parseEmbargo parses delay strings like '-1M' or '-3Y' into a time.Duration.
func parseEmbargo(s string) time.Duration {
	var d time.Duration
	if s == "" {
		return d
	}
	ms := delayPattern.FindStringSubmatch(s)
	if len(ms) != 3 {
		return d
	}
	value, err := strconv.Atoi(ms[1])
	if err != nil {
		return d
	}
	switch {
	case ms[2] == "Y":
		return time.Duration(value) * Year
	case ms[2] == "M":
		return time.Duration(value) * Month
	default:
		return d
	}
}

func (r Reader) ReadEntries() (holdings.Entries, error) {
	entries := make(holdings.Entries)
	decoder := xml.NewDecoder(r.r)

	// collect errors, let caller decide policy
	perr := holdings.ParseError{}

	var tag string

	for {

		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return entries, err
		}
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			tag = se.Name.Local
			if tag == "holding" {
				var item Holding
				if err := decoder.DecodeElement(&item, &se); err != nil {
					perr.Errors = append(perr.Errors, err)
					continue
				}

				for _, ent := range item.Entitlements {
					entry := holdings.Entry{
						Begin: holdings.Signature{
							Date:   ent.FromYear,
							Volume: ent.FromVolume,
							Issue:  ent.FromIssue,
						},
						End: holdings.Signature{
							Date:   ent.ToYear,
							Volume: ent.ToVolume,
							Issue:  ent.ToIssue,
						},
						Embargo: parseEmbargo(ent.FromDelay),
					}
					for _, issn := range append(item.EISSN, item.PISSN...) {
						entries[issn] = append(entries[issn], entry)
					}
				}
			}
		}
	}
	if len(perr.Errors) > 0 {
		return entries, perr
	}
	return entries, nil
}
