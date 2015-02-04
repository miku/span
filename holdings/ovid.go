package holdings

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// Holding contains a single holding
type Holding struct {
	EZBID        int           `xml:"ezb_id,attr"`
	Title        string        `xml:"title"`
	Publishers   string        `xml:"publishers"`
	PISSN        []string      `xml:"EZBIssns>p-issn"`
	EISSN        []string      `xml:"EZBIssns>e-issn"`
	Entitlements []Entitlement `xml:"entitlements>entitlement"`
}

// Entitlement holds a single OVID entitlement
type Entitlement struct {
	Status     string `xml:"status,attr"`
	URL        string `xml:"url"`
	Anchor     string `xml:"anchor"`
	FromYear   int    `xml:"begin>year"`
	FromVolume int    `xml:"begin>volume"`
	FromIssue  int    `xml:"begin>issue"`
	FromDelay  string `xml:"begin>delay"`
	ToYear     int    `xml:"end>year"`
	ToVolume   int    `xml:"end>volume"`
	ToIssue    int    `xml:"end>issue"`
	ToDelay    string `xml:"end>delay"`
}

// String returns a string representation of an Entitlement
func (e *Entitlement) String() string {
	delay, _ := e.Delay()
	unescaped, _ := url.QueryUnescape(e.URL)
	effective, _ := e.Effective()
	return fmt.Sprintf("<Entitlement status=%s url=%s range=%d/%d/%d-%d/%d/%d effective=%s delay=%0.2f>",
		e.Status, unescaped, e.FromYear, e.FromVolume, e.FromIssue, e.ToYear, e.ToVolume, e.ToIssue, effective, delay.Hours())
}

// Parse '1M', '3Y', ... into a duration
func ParseDelay(s string) (d time.Duration, err error) {
	r := regexp.MustCompile(`(\d+)(M|Y)`)
	ms := r.FindStringSubmatch(s)
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
	}
	return d, err
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

// Effective returns the last allowed date before the moving wall
func (e *Entitlement) Effective() (d time.Time, err error) {
	delay, err := e.Delay()
	if err != nil {
		return d, err
	}
	return time.Now().Add(delay), nil
}

// ISSNSet returns the ISSNs contained in an OVID holding file
func ISSNSet(reader io.Reader) map[string]bool {
	issns := make(map[string]bool)
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
					issns[id] = true
				}
				for _, id := range item.PISSN {
					issns[id] = true
				}
			}
		default:
		}
	}
	return issns
}
