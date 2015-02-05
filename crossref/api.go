// crossref implements crossref related structs and transformations
package crossref

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/miku/span/finc"
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
}

// Year returns the *first* year found inside a DateField
func (d *DateField) Year() int {
	parts := d.DateParts
	if len(parts) >= 1 {
		if len(parts[0]) > 0 {
			return parts[0][0]
		}
	}
	return 0
}

// Date returns a time.Date in a best effort way
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

// StartPage returns the first page as string
func (d *Document) StartPage() (p string) {
	parts := strings.Split(d.Page, "-")
	switch len(parts) {
	case 1:
		p = parts[0]
	case 2:
		p = parts[0]
	}
	return
}

// StartPage returns the last page as string
func (d *Document) EndPage() (p string) {
	parts := strings.Split(d.Page, "-")
	switch len(parts) {
	case 1:
		p = parts[0]
	case 2:
		p = parts[1]
	}
	return
}

// CombinedTitle returns a longish title
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

// FullTitle returns everything title
func (d *Document) FullTitle() string {
	return strings.Join(append(d.Title, d.Subtitle...), " ")
}

// ShortTitle returns the first main title only
func (d *Document) ShortTitle() string {
	if len(d.Title) > 0 {
		return d.Title[0]
	} else {
		return ""
	}
}

// Transform converts a single crossref document into a finc.Schema
func Transform(doc Document) (finc.Schema, error) {
	var output finc.Schema

	if doc.URL == "" {
		return output, errors.New("input document has no URL")
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(doc.URL))
	output.ID = fmt.Sprintf("ai049%s", encoded)
	output.ISSN = doc.ISSN
	output.Publisher = doc.Publisher
	output.SourceID = "49"
	output.RecordType = "ai"
	output.Title = doc.CombinedTitle()
	output.TitleFull = doc.FullTitle()
	output.TitleShort = doc.ShortTitle()
	output.Topic = doc.Subject
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

	if doc.Issued.Year() > 0 {
		output.PublishDateSort = doc.Issued.Year()
	}

	allfields := [][]string{output.SecondaryAuthors, doc.Subject, doc.ISSN, doc.Title,
		doc.Subtitle, doc.ContainerTitle, []string{doc.Publisher, doc.URL}}

	var buf bytes.Buffer
	for _, f := range allfields {
		_, err := buf.WriteString(fmt.Sprintf("%s ", strings.Join(f, " ")))
		if err != nil {
			log.Fatal(err)
		}
	}

	output.Allfields = buf.String()
	return output, nil
}
