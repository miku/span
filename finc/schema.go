// Package finc holds finc SolrSchema (SOLR) and intermediate schema related types and methods.
package finc

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	AIRecordType              = "ai"
	IntermediateSchemaVersion = "0.9"
)

var (
	NOT_ASSIGNED    = "not assigned"
	NonAlphaNumeric = regexp.MustCompile("/[^A-Za-z0-9]+/")
)

// ExportSchema encapsulate an export flavour. This will most likely be a
// struct with fields and methods relevant to the exported format. For the
// moment we assume, the output is JSON. If formats other than JSON are
// requested, move the marshalling into this interface.
type ExportSchema interface {
	// Convert takes an intermediate schema record to export. Returns an
	// error, if conversion failed.
	Convert(IntermediateSchema) error
	// Attach takes a list of strings (here: ISILs) and attaches them to the
	// current record.
	Attach([]string)
}

// Author representes an author, "inspired" by OpenURL.
type Author struct {
	ID           string `json:"x.id,omitempty"`
	Name         string `json:"rft.au,omitempty"`
	LastName     string `json:"rft.aulast,omitempty"`
	FirstName    string `json:"rft.aufirst,omitempty"`
	Initial      string `json:"rft.auinit,omitempty"`
	FirstInitial string `json:"rft.auinit1,omitempty"`
	MiddleName   string `json:"rft.auinitm,omitempty"`
	Suffix       string `json:"rft.ausuffix,omitempty"`
	Corporation  string `json:"rft.aucorp,omitempty"`
}

// String returns a formatted author string.
// TODO(miku): make this complete.
func (author *Author) String() string {
	if author.Name != "" {
		return author.Name
	}
	if author.LastName != "" {
		if author.FirstName != "" {
			return fmt.Sprintf("%s, %s", author.LastName, author.FirstName)
		}
		return author.LastName
	}
	return author.ID
}

// IntermediateSchema abstract and collects the values of various input formats.
// Goal is to simplify further processing by using a single format, from which
// the next artifacts can be derived, e.g. records for solr indices.
// This format can be viewed as a catch-all format. The dotted notation
// hints at the origin of the field, e.g. OpenURL, RIS, finc.
//
// Notes on the format:
// - The x namespace is experimental.
// - RawDate must be in ISO8601 (YYYY-MM-DD) format.
// - Version is mandatory.
// - Headings and Subjects are not bound to any format yet.
// - Use plural for slices, if possible.
type IntermediateSchema struct {
	Format         string `json:"finc.format,omitempty"`
	MegaCollection string `json:"finc.mega_collection,omitempty"`
	RecordID       string `json:"finc.record_id,omitempty"`
	SourceID       string `json:"finc.source_id,omitempty"`

	Database     string `json:"ris.db,omitempty"`
	DataProvider string `json:"ris.dp,omitempty"`
	RefType      string `json:"ris.type,omitempty"`

	ArticleNumber string `json:"rft.artnum,omitempty"`
	ArticleTitle  string `json:"rft.atitle,omitempty"`

	BookTitle    string    `json:"rft.btitle,omitempty"`
	Chronology   string    `json:"rft.chron,omitempty"`
	Edition      string    `json:"rft.edition,omitempty"`
	EISBN        []string  `json:"rft.isbn,omitempty"`
	EISSN        []string  `json:"rft.eissn,omitempty"`
	EndPage      string    `json:"rft.epage,omitempty"`
	Genre        string    `json:"rft.genre,omitempty"`
	ISBN         []string  `json:"rft.isbn,omitempty"`
	ISSN         []string  `json:"rft.issn,omitempty"`
	Issue        string    `json:"rft.issue,omitempty"`
	JournalTitle string    `json:"rft.jtitle,omitempty"`
	PageCount    string    `json:"rft.tpages,omitempty"`
	Pages        string    `json:"rft.pages,omitempty"`
	Part         string    `json:"rft.part,omitempty"`
	Places       []string  `json:"rft.place,omitempty"`
	Publishers   []string  `json:"rft.pub,omitempty"`
	Quarter      string    `json:"rft.quarter,omitempty"`
	RawDate      string    `json:"rft.date,omitempty"`
	Date         time.Time `json:"x.date,omitempty"`
	Season       string    `json:"rft.ssn,omitempty"`
	Series       string    `json:"rft.series,omitempty"`
	ShortTitle   string    `json:"rft.stitle,omitempty"`
	StartPage    string    `json:"rft.spage,omitempty"`
	Volume       string    `json:"rft.volume,omitempty"`

	Abstract  string   `json:"abstract,omitempty"`
	Authors   []Author `json:"authors,omitempty"`
	DOI       string   `json:"doi,omitempty"`
	Languages []string `json:"languages,omitempty"`
	URL       []string `json:"url,omitempty"`
	Version   string   `json:"version,omitempty"`

	ArticleSubtitle string   `json:"x.subtitle,omitempty"`
	Fulltext        string   `json:"x.fulltext,omitempty"`
	Headings        []string `json:"x.headings,omitempty"`
	Subjects        []string `json:"x.subjects,omitempty"`
	Type            string   `json:"x.type,omitempty"`
}

func NewIntermediateSchema() *IntermediateSchema {
	return &IntermediateSchema{Version: IntermediateSchemaVersion}
}

// ISSNList returns a deduplicated list of all ISSN and EISSN.
func (is *IntermediateSchema) ISSNList() []string {
	set := make(map[string]struct{})
	for _, issn := range append(is.ISSN, is.EISSN...) {
		set[issn] = struct{}{}
	}
	var issns []string
	for k := range set {
		issns = append(issns, k)
	}
	return issns
}

// Allfields returns a combination of various fields.
func (is *IntermediateSchema) Allfields() string {
	var authors []string
	for _, author := range is.Authors {
		authors = append(authors, author.String())
	}
	fields := [][]string{authors,
		is.Subjects, is.ISSN, is.EISSN, is.Publishers, is.Places, is.URL,
		{is.ArticleTitle, is.ArticleSubtitle, is.JournalTitle, is.Fulltext, is.Abstract}}
	var buf bytes.Buffer
	for _, f := range fields {
		for _, value := range f {
			for _, token := range strings.Fields(value) {
				buf.WriteString(fmt.Sprintf("%s ", strings.TrimSpace(token)))
			}
		}
	}
	return strings.TrimSpace(buf.String())
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Imprint MARC 260 a, b, c (trad.)
func (is *IntermediateSchema) Imprint() (s string) {
	var places, publisher string
	year := is.Date.Year()
	places = strings.Join(is.Places, " ; ")
	if len(is.Publishers) > 0 {
		publisher = is.Publishers[0]
	}

	mask := btoi(places != "") + 2*btoi(publisher != "") + 4*btoi(year != 0)

	switch mask {
	case 1:
		s = places
	case 2:
		s = publisher
	case 3:
		s = fmt.Sprintf("%s : %s", places, publisher)
	case 4:
		s = fmt.Sprintf("%d", year)
	case 5:
		s = fmt.Sprintf("%s : %d", places, year)
	case 6:
		s = fmt.Sprintf("%s, %d", publisher, year)
	case 7:
		s = fmt.Sprintf("%s : %s, %d", places, publisher, year)
	}
	return
}

// SortableAuthor is loosely based on solrmarcs builtin getSortableTitle
func (is *IntermediateSchema) SortableTitle() string {
	return strings.ToLower(NonAlphaNumeric.ReplaceAllString(is.ArticleTitle, ""))
}

// SortableAuthor is loosely based on solrmarcs builtin getSortableAuthor
func (is *IntermediateSchema) SortableAuthor() string {
	var buf bytes.Buffer
	for _, author := range is.Authors {
		buf.WriteString(strings.ToLower(NonAlphaNumeric.ReplaceAllString(author.String(), "")))
	}
	buf.WriteString(is.SortableTitle())
	return buf.String()
}
