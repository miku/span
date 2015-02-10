// Package crossref implements crossref related structs and transformations.
//
// API endpoint/documentation: http://api.crossref.org
package crossref

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/finc"
	"github.com/miku/span/holdings"
)

// Author is given by family and given name
type Author struct {
	Family string `json:"family"`
	Given  string `json:"given"`
}

// String pretty print the author
func (author *Author) String() string {
	if author.Given != "" {
		return fmt.Sprintf("%s, %s", author.Family, author.Given)
	} else {
		return author.Family
	}
}

// DatePart consists of up to three int, representing year, month, day
type DatePart []int

// DateField contains two representations of one value
type DateField struct {
	Timestamp int64      `json:"timestamp"`
	DateParts []DatePart `json:"date-parts"`
}

// Document is a example API response
type Document struct {
	Prefix         string    `json:"prefix"`
	Type           string    `json:"type"`
	Volume         string    `json:"volume"`
	Deposited      DateField `json:"deposited"`
	Source         string    `json:"source"`
	Authors        []Author  `json:"author"`
	Score          float64   `json:"score"`
	Page           string    `json:"page"`
	Subject        []string  `json:"subject"`
	Title          []string  `json:"title"`
	Publisher      string    `json:"publisher"`
	ISSN           []string  `json:"ISSN"`
	Indexed        DateField `json:"indexed"`
	Issued         DateField `json:"issued"`
	Subtitle       []string  `json:"subtitle"`
	URL            string    `json:"URL"`
	Issue          string    `json:"issue"`
	ContainerTitle []string  `json:"container-title"`
	ReferenceCount int       `json:"reference-count"`
	DOI            string    `json:"DOI"`
	Member         string    `json:"member"`
}

// Year returns the first year found inside a DateField.
func (d *DateField) Year() int {
	parts := d.DateParts
	if len(parts) >= 1 {
		if len(parts[0]) > 0 {
			return parts[0][0]
		}
	}
	return 0
}

// Date returns a time.Date in a best effort way.
func (d *DateField) Date() (t time.Time) {
	if len(d.DateParts) == 0 {
		t, _ = time.Parse("2006-01-02", "0000-00-00")
	}
	p := d.DateParts[0]
	switch len(p) {
	case 1:
		t, _ = time.Parse("2006-01-02", fmt.Sprintf("%04d-01-01", p[0]))
	case 2:
		t, _ = time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-01", p[0], p[1]))
	case 3:
		t, _ = time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-%02d", p[0], p[1], p[2]))
	default:
		t, _ = time.Parse("2006-01-02", "1970-01-01")
	}
	return t
}

// CombinedTitle returns a longish title.
func (d *Document) CombinedTitle() string {
	if len(d.Title) > 0 {
		if len(d.Subtitle) > 0 {
			return fmt.Sprintf("%s : %s", d.Title[0], d.Subtitle[0])
		} else {
			return d.Title[0]
		}
	} else {
		if len(d.Subtitle) > 0 {
			return d.Subtitle[0]
		} else {
			return ""
		}
	}
}

// FullTitle returns everything title.
func (d *Document) FullTitle() string {
	return strings.Join(append(d.Title, d.Subtitle...), " ")
}

// ShortTitle returns the first main title only.
func (d *Document) ShortTitle() string {
	if len(d.Title) > 0 {
		return d.Title[0]
	} else {
		return ""
	}
}

// CoveredBy returns nil, if a given entitlement covers the current document.
// Otherwise the error con
func (d *Document) CoveredBy(e holdings.Entitlement) error {
	if e.FromYear != 0 && e.FromYear > d.Issued.Year() {
		return fmt.Errorf("from-year %d > %d", e.FromYear, d.Issued.Year())
	}
	if e.FromYear == d.Issued.Year() {
		volume, err := strconv.Atoi(d.Volume)
		if err != nil {
			return err
		}
		if e.FromVolume != 0 && e.FromVolume > volume {
			return fmt.Errorf("from-volume %d > %d", e.FromVolume, volume)
		}
		if e.FromVolume == volume {
			issue, err := strconv.Atoi(d.Issue)
			if err != nil {
				return err
			}
			if e.FromIssue != 0 && e.FromIssue > issue {
				return fmt.Errorf("from-issue %d > %d", e.FromIssue, issue)
			}
		}
	}
	if e.ToYear != 0 && e.ToYear < d.Issued.Year() {
		return fmt.Errorf("to-year %d < %d", e.ToYear, d.Issued.Year())
	}
	if e.ToYear == d.Issued.Year() {
		volume, err := strconv.Atoi(d.Volume)
		if err != nil {
			return err
		}
		if e.ToVolume != 0 && e.ToVolume < volume {
			return fmt.Errorf("to-volume %d < %d", e.ToVolume, volume)
		}
		if e.ToVolume == volume {
			issue, err := strconv.Atoi(d.Issue)
			if err != nil {
				return err
			}
			if e.ToIssue != 0 && e.ToIssue < issue {
				return fmt.Errorf("to-issue %d < %d", e.ToIssue, issue)
			}
		}
	}
	boundary, err := e.Boundary()
	if err != nil {
		return err
	}
	if d.Issued.Date().After(boundary) {
		return fmt.Errorf("moving-wall %s %s", boundary, d.Issued.Date())
	}
	return nil
}

// ParseMemberID extracts the numeric member id.
func (d *Document) ParseMemberID() (int, error) {
	fields := strings.Split(d.Member, "/")
	if len(fields) > 0 {
		id, err := strconv.Atoi(fields[len(fields)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid member: %s", d.Member)
		}
		return id, nil
	}
	return 0, fmt.Errorf("invalid member: %s", d.Member)
}

// ToSchema converts a single crossref document into a basic finc schema.
func (d *Document) ToSchema() (output finc.Schema, err error) {
	if d.URL == "" {
		return output, errors.New("input document has no URL")
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(d.URL))
	output.ID = fmt.Sprintf("ai049%s", encoded)
	output.ISSN = d.ISSN
	output.Publisher = d.Publisher
	output.SourceID = "49"
	output.RecordType = "ai"
	output.Title = d.CombinedTitle()
	output.TitleFull = d.FullTitle()
	output.TitleShort = d.ShortTitle()
	output.Topics = d.Subject
	output.URL = d.URL

	if len(d.ContainerTitle) > 0 {
		output.HierarchyParentTitle = d.ContainerTitle[0]
	}

	if d.Type == "journal-article" {
		output.Format = "ElectronicArticle"
	}

	for _, author := range d.Authors {
		output.SecondaryAuthors = append(output.SecondaryAuthors, author.String())
	}

	if d.Issued.Year() > 0 {
		output.PublishDateSort = d.Issued.Year()
	}

	allfields := [][]string{output.SecondaryAuthors, d.Subject, d.ISSN, d.Title,
		d.Subtitle, d.ContainerTitle, []string{d.Publisher, d.URL}}

	var buf bytes.Buffer
	for _, f := range allfields {
		_, err := buf.WriteString(fmt.Sprintf("%s ", strings.Join(f, " ")))
		if err != nil {
			log.Fatal(err)
		}
	}

	output.Allfields = buf.String()

	id, err := d.ParseMemberID()
	if err != nil {
		log.Println(err)
	} else {
		member, err := LookupMember(id)
		if err != nil {
			log.Println(err)
		} else {
			output.AddMegaCollection(fmt.Sprintf("%s (CrossRef)", member.PrimaryName))
		}
	}
	return output, nil
}

// Institutions returns a slice of ISILs for which this document finds
// valid entitlements in a IsilIssnHolding map.
func (d *Document) Institutions(iih holdings.IsilIssnHolding) []string {
	isils := span.NewStringSet()
	for _, isil := range iih.Isils() {
		covered := false
		for _, issn := range d.ISSN {
			h, ok := iih[isil][issn]
			if ok {
				for _, entitlement := range h.Entitlements {
					err := d.CoveredBy(entitlement)
					if err == nil {
						covered = true
					}
					if covered {
						break
					}
				}
			}
			if covered {
				break
			}
		}
		if covered {
			isils.Add(isil)
		}
	}
	return isils.Values()
}
