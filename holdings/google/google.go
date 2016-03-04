package google

import (
	"bufio"
	"encoding/xml"
	"io"
	"time"

	"github.com/miku/span/holdings"
)

// Item is the main google scholar holdings container.
type Item struct {
	Title string     `xml:"title"`
	ISSN  string     `xml:"issn"`
	Covs  []Coverage `xml:"coverage"`
}

// Coverage contains coverage information for an item.
type Coverage struct {
	FromYear         string `xml:"from>year"`
	FromVolume       string `xml:"from>volume"`
	FromIssue        string `xml:"from>issue"`
	ToYear           string `xml:"to>year"`
	ToVolume         string `xml:"to>volume"`
	ToIssue          string `xml:"to>issue"`
	Comment          string `xml:"comment"`
	DaysNotAvailable int    `xml:"embargo>days_not_available"`
}

type Reader struct {
	r io.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r)}
}

func parseEmbargo(i int) time.Duration {
	var d time.Duration
	if i == 0 {
		return d
	}
	return time.Duration(i*24) * time.Hour
}

func (r Reader) ReadEntries() (holdings.Entries, error) {
	entries := make(holdings.Entries)
	decoder := xml.NewDecoder(r.r)
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
			if tag == "item" {
				var item Item
				err := decoder.DecodeElement(&item, &se)
				if err != nil {
					return entries, err
				}

				for _, cov := range item.Covs {
					entry := holdings.Entry{
						Begin: holdings.Signature{
							Date:   cov.FromYear,
							Volume: cov.FromVolume,
							Issue:  cov.FromIssue,
						},
						End: holdings.Signature{
							Date:   cov.ToYear,
							Volume: cov.ToVolume,
							Issue:  cov.ToIssue,
						},
						Embargo: parseEmbargo(cov.DaysNotAvailable),
					}
					entries[item.ISSN] = append(entries[item.ISSN], entry)
				}
			}
		}
	}
	return entries, nil
}
