// crossref implements crossref related structs and transformations
package crossref

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"

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

// Document is a example API response
type Document struct {
	Prefix         string     `json:"prefix"`
	Type           string     `json:"type"`
	Volume         string     `json:"volume"`
	Deposited      []DatePart `json:"deposited"`
	Source         string     `json:"source"`
	Authors        []Author   `json:"author"`
	Score          int        `json:"score"`
	Page           string     `json:"page"`
	Subject        []string   `json:"subject"`
	Title          []string   `json:"title"`
	Publisher      string     `json:"publisher"`
	ISSN           []string   `json:"ISSN"`
	Indexed        []DatePart `json:"indexed"`
	Issued         []DatePart `json:"issued"`
	Subtitle       []string   `json:"subtitle"`
	URL            string     `json:"URL"`
	Issue          string     `json:"issue"`
	ContainerTitle []string   `json:"container-title"`
	ReferenceCount int        `json:"reference-count"`
	DOI            string     `json:"DOI"`
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

	if len(doc.Issued) > 0 {
		if len(doc.Issued[0]) > 0 {
			output.PublishDateSort = doc.Issued[0][0]
		}
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
