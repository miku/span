// Package holdings contains wrappers for various holding file formats.
package holdings

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"
)

// DelayPattern is how moving walls are expressed in OVID format
var DelayPattern = regexp.MustCompile(`^-(\d+)(M|Y)$`)

// Holding contains a single holding
type Holding struct {
	EZBID        int           `xml:"ezb_id,attr" json:"ezbid"`
	Title        string        `xml:"title" json:"title"`
	Publishers   string        `xml:"publishers" json:"publishers"`
	PISSN        []string      `xml:"EZBIssns>p-issn" json:"pissn"`
	EISSN        []string      `xml:"EZBIssns>e-issn" json:"eissn"`
	Entitlements []Entitlement `xml:"entitlements>entitlement" json:"entitlements"`
}

// Entitlement holds a single OVID entitlement
type Entitlement struct {
	Status     string `xml:"status,attr" json:"status"`
	URL        string `xml:"url" json:"url"`
	Anchor     string `xml:"anchor" json:"anchor"`
	FromYear   int    `xml:"begin>year" json:"from-year"`
	FromVolume int    `xml:"begin>volume" json:"from-volume"`
	FromIssue  int    `xml:"begin>issue" json:"from-issue"`
	FromDelay  string `xml:"begin>delay" json:"from-delay"`
	ToYear     int    `xml:"end>year" json:"to-year"`
	ToVolume   int    `xml:"end>volume" json:"to-volume"`
	ToIssue    int    `xml:"end>issue" json:"to-issue"`
	ToDelay    string `xml:"end>delay" json:"to-delay"`
}

// IssnHolding maps an ISSN to a holdings.Holding struct
type IssnHolding map[string]Holding

// IsilIssnHolding maps an ISIL to an IssnHolding map
type IsilIssnHolding map[string]IssnHolding

// Isils returns available ISILs in this IsilIssnHolding map
func (iih *IsilIssnHolding) Isils() []string {
	var keys []string
	for k, _ := range *iih {
		keys = append(keys, k)
	}
	return keys
}

// ParseDelay parses delay strings like '-1M', '-3Y', ... into a time.Duration
func ParseDelay(s string) (d time.Duration, err error) {
	ms := DelayPattern.FindStringSubmatch(s)
	if len(ms) == 3 {
		value, err := strconv.Atoi(ms[1])
		if err != nil {
			return d, err
		}
		switch {
		case ms[2] == "Y":
			d, err = time.ParseDuration(fmt.Sprintf("-%dh", value*8760))
		case ms[2] == "M":
			d, err = time.ParseDuration(fmt.Sprintf("-%dh", value*720))
		default:
			return d, fmt.Errorf("unknown unit: %s", ms[2])
		}
	} else {
		return d, fmt.Errorf("unknown format: %s", s)
	}
	return d, nil
}

// Delay returns the specified delay as `time.Duration`
func (e *Entitlement) Delay() (d time.Duration, err error) {
	if e.FromDelay != "" {
		return ParseDelay(e.FromDelay)
	}
	if e.ToDelay != "" {
		return ParseDelay(e.ToDelay)
	}
	return d, nil
}

// Boundary returns the last date before the moving wall restriction becomes effective
func (e *Entitlement) Boundary() (d time.Time, err error) {
	delay, err := e.Delay()
	if err != nil {
		return d, err
	}
	return time.Now().Add(delay), nil
}

// HoldingsMap creates an ISSN[Holding] struct from a reader
func HoldingsMap(reader io.Reader) (h IssnHolding) {
	h = make(map[string]Holding)
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
					h[id] = item
				}
				for _, id := range item.PISSN {
					h[id] = item
				}
			}
		}
	}
	return h
}
