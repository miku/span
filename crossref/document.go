// Package crossref implements crossref related structs and transformations.
//
// API endpoint/documentation: http://api.crossref.org
package crossref

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/finc"
	"github.com/miku/span/holdings"
)

// Author is given by family and given name.
type Author struct {
	Family string `json:"family"`
	Given  string `json:"given"`
}

// String pretty prints the author.
func (author *Author) String() string {
	if author.Given != "" {
		if author.Family != "" {
			return fmt.Sprintf("%s, %s", author.Family, author.Given)
		} else {
			return author.Given
		}
	}
	return author.Family
}

// DatePart consists of up to three int, representing year, month, day.
type DatePart []int

// DateField contains two representations of one value.
type DateField struct {
	DateParts []DatePart `json:"date-parts"`
	Timestamp int64      `json:"timestamp"`
}

// Document is a example works API response.
type Document struct {
	Authors        []Author  `json:"author"`
	ContainerTitle []string  `json:"container-title"`
	Deposited      DateField `json:"deposited"`
	DOI            string    `json:"DOI"`
	Indexed        DateField `json:"indexed"`
	ISSN           []string  `json:"ISSN"`
	Issue          string    `json:"issue"`
	Issued         DateField `json:"issued"`
	Member         string    `json:"member"`
	Page           string    `json:"page"`
	Prefix         string    `json:"prefix"`
	Publisher      string    `json:"publisher"`
	ReferenceCount int       `json:"reference-count"`
	Score          float64   `json:"score"`
	Source         string    `json:"source"`
	Subject        []string  `json:"subject"`
	Subtitle       []string  `json:"subtitle"`
	Title          []string  `json:"title"`
	Type           string    `json:"type"`
	URL            string    `json:"URL"`
	Volume         string    `json:"volume"`
}

// PageInfo holds various page related data.
type PageInfo struct {
	RawMessage string
	StartPage  int
	EndPage    int
}

// PageCount returns the number of pages, or zero if this cannot be determined.
func (pi *PageInfo) PageCount() int {
	if pi.StartPage != 0 && pi.EndPage != 0 {
		count := pi.EndPage - pi.StartPage
		if count > 0 {
			return count
		}
	}
	return 0
}

// ParsePageValue parses a page specfication in a best effort manner into a PageInfo struct.
func (doc *Document) PageInfo() PageInfo {
	pi := PageInfo{RawMessage: doc.Page}
	parts := strings.Split(doc.Page, "-")
	if len(parts) != 2 {
		return pi
	}
	spage, err := strconv.Atoi(parts[0])
	if err != nil {
		return pi
	}
	pi.StartPage = spage

	epage, err := strconv.Atoi(parts[1])
	if err != nil {
		return pi
	}
	pi.EndPage = epage
	return pi
}

// Year returns the first year found inside a DateField.
func (d *DateField) Year() int {
	if len(d.DateParts) >= 1 && len(d.DateParts[0]) > 0 {
		return d.DateParts[0][0]
	}
	return 0
}

// Date returns a time.Date in a best effort way. Date parts are always
// present in the source document, while timestamp is only present if
// dateparts consist of year, month and day.
func (d *DateField) Date() (t time.Time) {
	if len(d.DateParts) == 0 {
		t, _ = time.Parse("2006-01-02", "0000-00-00")
		return t
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
func (doc *Document) CombinedTitle() string {
	if len(doc.Title) > 0 {
		if len(doc.Subtitle) > 0 {
			return fmt.Sprintf("%s : %s", doc.Title[0], doc.Subtitle[0])
		} else {
			return doc.Title[0]
		}
	} else {
		if len(doc.Subtitle) > 0 {
			return doc.Subtitle[0]
		} else {
			return ""
		}
	}
}

// FullTitle returns everything title.
func (doc *Document) FullTitle() string {
	return strings.Join(append(doc.Title, doc.Subtitle...), " ")
}

// ShortTitle returns the first main title only.
func (doc *Document) ShortTitle() string {
	if len(doc.Title) > 0 {
		return doc.Title[0]
	} else {
		return ""
	}
}

// Allfields returns a combination of various fields.
func (doc *Document) Allfields() string {
	var authors []string
	for _, author := range doc.Authors {
		authors = append(authors, author.String())
	}

	fields := [][]string{authors,
		doc.Subject, doc.ISSN, doc.Title, doc.Subtitle, doc.ContainerTitle,
		[]string{doc.Publisher, doc.URL}}

	var buf bytes.Buffer
	for _, f := range fields {
		for _, value := range f {
			for _, token := range strings.Fields(value) {
				_, err := buf.WriteString(fmt.Sprintf("%s ", strings.TrimSpace(token)))
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
	return strings.TrimSpace(buf.String())
}

// MemberName resolves the primary name of the member.
func (doc *Document) MemberName() (name string, err error) {
	id, err := doc.ParseMemberID()
	if err != nil {
		return
	}
	name, err = LookupMemberName(id)
	return
}

// CoveredBy returns nil, if a given entitlement covers the current document.
func (doc *Document) CoveredBy(e holdings.Entitlement) error {
	if e.FromYear != 0 && doc.Issued.Year() != 0 {
		if e.FromYear > doc.Issued.Year() {
			return fmt.Errorf("from-year %d > %d", e.FromYear, doc.Issued.Year())
		}
		if e.FromYear == doc.Issued.Year() {
			volume, err := strconv.Atoi(doc.Volume)
			if err != nil {
				return err
			}
			if e.FromVolume != 0 && e.FromVolume > volume {
				return fmt.Errorf("from-volume %d > %d", e.FromVolume, volume)
			}
			if e.FromVolume == volume {
				issue, err := strconv.Atoi(doc.Issue)
				if err != nil {
					return err
				}
				if e.FromIssue != 0 && e.FromIssue > issue {
					return fmt.Errorf("from-issue %d > %d", e.FromIssue, issue)
				}
			}
		}
	}

	if e.ToYear != 0 && doc.Issued.Year() != 0 {
		if e.ToYear < doc.Issued.Year() {
			return fmt.Errorf("to-year %d < %d", e.ToYear, doc.Issued.Year())
		}
		if e.ToYear == doc.Issued.Year() {
			volume, err := strconv.Atoi(doc.Volume)
			if err != nil {
				return err
			}
			if e.ToVolume != 0 && e.ToVolume < volume {
				return fmt.Errorf("to-volume %d < %d", e.ToVolume, volume)
			}
			if e.ToVolume == volume {
				issue, err := strconv.Atoi(doc.Issue)
				if err != nil {
					return err
				}
				if e.ToIssue != 0 && e.ToIssue < issue {
					return fmt.Errorf("to-issue %d < %d", e.ToIssue, issue)
				}
			}
		}
	}

	boundary, err := e.Boundary()
	if err != nil {
		return err
	}
	if doc.Issued.Date().After(boundary) {
		return fmt.Errorf("moving-wall violation")
	}
	return nil
}

// ParseMemberID extracts the numeric member id.
func (doc *Document) ParseMemberID() (int, error) {
	fields := strings.Split(doc.Member, "/")
	if len(fields) > 0 {
		id, err := strconv.Atoi(fields[len(fields)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid member: %s", doc.Member)
		}
		return id, nil
	}
	return 0, fmt.Errorf("invalid member: %s", doc.Member)
}

// ToSchema converts a crossref document into an intermediate schema.
func (doc *Document) ToSchema() (output finc.Schema, err error) {
	if doc.URL == "" {
		return output, errors.New("input document has no URL")
	}

	output.RecordID = fmt.Sprintf("ai049%s", base64.StdEncoding.EncodeToString([]byte(doc.URL)))

	// TODO(miku): how to handle multiple ISSNs?
	// output.ISSN = doc.ISSN

	output.URL = doc.URL
	output.DOI = doc.DOI
	output.SourceID = "49"
	output.Publisher = doc.Publisher
	output.ArticleTitle = doc.CombinedTitle()
	output.Issue = doc.Issue
	output.Volume = doc.Volume

	if len(doc.ContainerTitle) > 0 {
		output.JournalTitle = doc.ContainerTitle[0]
	}

	// authors
	for _, author := range doc.Authors {
		output.Authors = append(output.Authors, finc.Author{FirstName: author.Given, LastName: author.Family})
	}

	// pages
	pi := doc.PageInfo()
	output.StartPage = fmt.Sprintf("%d", pi.StartPage)
	output.EndPage = fmt.Sprintf("%d", pi.EndPage)
	output.Pages = pi.RawMessage
	output.PageCount = fmt.Sprintf("%d", pi.PageCount())

	// date
	output.Date = doc.Issued.Date().Format("2006-01-02")

	name, err := doc.MemberName()
	if err == nil {
		output.MegaCollection = fmt.Sprintf("%s (CrossRef)", name)
	}
	return output, nil
}

// ToSolrSchema converts a single crossref document into a basic finc solr schema.
func (doc *Document) ToSolrSchema() (output finc.SolrSchema, err error) {
	if doc.URL == "" {
		return output, errors.New("input document has no URL")
	}

	output.ID = fmt.Sprintf("ai049%s", base64.StdEncoding.EncodeToString([]byte(doc.URL)))
	output.ISSN = doc.ISSN
	output.Publisher = doc.Publisher
	output.SourceID = "49"
	output.RecordType = "ai"
	output.Title = doc.CombinedTitle()
	output.TitleFull = doc.FullTitle()
	output.TitleShort = doc.ShortTitle()
	copy(output.Topics, doc.Subject)
	output.URL = doc.URL

	if len(doc.ContainerTitle) > 0 {
		output.HierarchyParentTitle = doc.ContainerTitle[0]
	}

	if doc.Type == "journal-article" {
		output.Format = "ElectronicArticle"
	}

	for _, author := range doc.Authors {
		output.SecondaryAuthors = append(output.SecondaryAuthors, author.String())
	}

	if len(output.SecondaryAuthors) > 1 {
		output.Author = output.SecondaryAuthors[0]
	}

	if doc.Issued.Year() > 0 {
		output.PublishDateSort = doc.Issued.Year()
	}

	output.Allfields = doc.Allfields()

	name, err := doc.MemberName()
	if err == nil {
		output.AddMegaCollection(fmt.Sprintf("%s (CrossRef)", name))
	}

	// marshal additional/intermediate representation info into fullrecord
	// TODO(miku): find a less ugly way, e.g. external storage?
	schema, err := doc.ToSchema()
	if err != nil {
		return output, err
	}

	b, err := json.Marshal(schema)
	if err != nil {
		return output, err
	}
	output.Fullrecord = string(b)

	return output, nil
}

// Institutions returns a slice of ISILs for which this document finds
// valid entitlements in a IsilIssnHolding map.
func (doc *Document) Institutions(iih holdings.IsilIssnHolding) []string {
	isils := span.NewStringSet()
	for _, isil := range iih.Isils() {
		for _, issn := range doc.ISSN {
			h, exists := iih[isil][issn]
			if !exists {
				continue
			}
			for _, entitlement := range h.Entitlements {
				err := doc.CoveredBy(entitlement)
				if err != nil {
					continue
				}
				isils.Add(isil)
				break
			}
		}
	}
	values := isils.Values()
	sort.Strings(values)
	return values
}
