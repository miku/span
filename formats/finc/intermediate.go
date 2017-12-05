//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                    The Finc Authors, http://finc.info
//                    Martin Czygan, <martin.czygan@uni-leipzig.de>
//
// This file is part of some open source application.
//
// Some open source application is free software: you can redistribute
// it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation, either
// version 3 of the License, or (at your option) any later version.
//
// Some open source application is distributed in the hope that it will
// be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
//
// @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
//
package finc

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	IntermediateSchemaRecordType = "is"
	AIRecordType                 = "ai"
	IntermediateSchemaVersion    = "0.9"
)

var (
	NotAssigned     = "" // was "not assigned", refs #7092
	NonAlphaNumeric = regexp.MustCompile("/[^A-Za-z0-9]+/")
)

// Exporter implements a basic export method that serializes an intermediate schema.
type Exporter interface {
	// Export turns an intermediate schema into bytes. Lower level
	// representation than ExportSchema.Convert. Allows JSON, XML, Marc,
	// Formeta and other formats.
	Export(is IntermediateSchema, withFullrecord bool) ([]byte, error)
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
//
// * The x namespace is experimental.
// * RawDate must be in ISO8601 (YYYY-MM-DD) format.
// * Version is mandatory.
// * Headings and Subjects are not bound to any format yet.
// * Use plural for slices, if possible.
//
// TODO(miku): Clean up naming and date parsing.
type IntermediateSchema struct {
	Format          string   `json:"finc.format,omitempty"`
	MegaCollections []string `json:"finc.mega_collection,omitempty"`
	ID              string   `json:"finc.id,omitempty"`
	RecordID        string   `json:"finc.record_id,omitempty"`
	SourceID        string   `json:"finc.source_id,omitempty"`

	Database     string `json:"ris.db,omitempty"`
	DataProvider string `json:"ris.dp,omitempty"`
	RefType      string `json:"ris.type,omitempty"`

	ArticleNumber string `json:"rft.artnum,omitempty"`
	ArticleTitle  string `json:"rft.atitle,omitempty"`

	BookTitle    string   `json:"rft.btitle,omitempty"`
	Chronology   string   `json:"rft.chron,omitempty"`
	Edition      string   `json:"rft.edition,omitempty"`
	EISBN        []string `json:"rft.eisbn,omitempty"`
	EISSN        []string `json:"rft.eissn,omitempty"`
	EndPage      string   `json:"rft.epage,omitempty"`
	Genre        string   `json:"rft.genre,omitempty"`
	ISBN         []string `json:"rft.isbn,omitempty"`
	ISSN         []string `json:"rft.issn,omitempty"`
	Issue        string   `json:"rft.issue,omitempty"`
	JournalTitle string   `json:"rft.jtitle,omitempty"`
	PageCount    string   `json:"rft.tpages,omitempty"`
	Pages        string   `json:"rft.pages,omitempty"`
	Part         string   `json:"rft.part,omitempty"`
	Places       []string `json:"rft.place,omitempty"`
	Publishers   []string `json:"rft.pub,omitempty"`
	Quarter      string   `json:"rft.quarter,omitempty"`

	// TODO(miku): we do not need both dates
	RawDate string    `json:"rft.date,omitempty"`
	Date    time.Time `json:"x.date,omitempty"`

	Season     string `json:"rft.ssn,omitempty"`
	Series     string `json:"rft.series,omitempty"`
	ShortTitle string `json:"rft.stitle,omitempty"`
	StartPage  string `json:"rft.spage,omitempty"`
	Volume     string `json:"rft.volume,omitempty"`

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

	// Indicator can hold update related information, e.g. in GBI the filedate
	Indicator string `json:"x.indicator,omitempty"`
	// Packages can hold set information, e.g. in GBI the licenced package or GBI database
	Packages []string `json:"x.packages,omitempty"`
	// Labels can carry a list of marks for a given records, e.g. ISILs
	Labels []string `json:"x.labels,omitempty"`

	// OpenAccess, refs. #8986, prototype
	OpenAccess bool     `json:"x.oa,omitempty"`
	License    []string `json:"x.license,omitempty"`
}

// NewIntermediateSchema creates a new intermediate schema document with the
// current version.
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

// ISBNList returns a deduplicated list of all ISBN and EISBN.
func (is *IntermediateSchema) ISBNList() []string {
	set := make(map[string]struct{})
	for _, isbn := range append(is.ISBN, is.EISBN...) {
		set[isbn] = struct{}{}
	}
	var isbns []string
	for k := range set {
		isbns = append(isbns, k)
	}
	return isbns
}

// ParsedDate turns tries to turn a raw date string into a date.
// TODO(miku): sources need to enforce a format, maybe enforce it here, too?
func (is *IntermediateSchema) ParsedDate() time.Time {
	t, _ := time.Parse("2006-01-02", is.RawDate)
	return t
}

// Allfields returns a combination of various fields.
func (is *IntermediateSchema) Allfields() string {
	var authors []string
	for _, author := range is.Authors {
		authors = append(authors, author.String())
	}

	fields := [][]string{
		// multivalued
		authors,
		is.EISBN,
		is.EISSN,
		is.ISBN,
		is.ISSN,
		is.Places,
		is.Publishers,
		is.Subjects,
		is.URL,
		{
			// single-valued
			is.Abstract,
			is.ArticleSubtitle,
			is.ArticleTitle,
			is.BookTitle,
			is.Edition,
			is.Fulltext,
			is.JournalTitle,
			is.Series,
			is.ShortTitle,
		}}

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

// SortableTitle is loosely based on getSortableTitle in SOLRMARC.
func (is *IntermediateSchema) SortableTitle() string {
	switch {
	case is.BookTitle != "":
		return strings.ToLower(NonAlphaNumeric.ReplaceAllString(is.BookTitle, ""))
	default:
		return strings.ToLower(NonAlphaNumeric.ReplaceAllString(is.ArticleTitle, ""))
	}
}

// SortableAuthor is loosely based on getSortableAuthor in SOLRMARC.
func (is *IntermediateSchema) SortableAuthor() string {
	var buf bytes.Buffer
	for _, author := range is.Authors {
		buf.WriteString(strings.ToLower(NonAlphaNumeric.ReplaceAllString(author.String(), "")))
	}
	buf.WriteString(is.SortableTitle())
	return buf.String()
}

// StrippedSchema is a snippet of an IntermediateSchema.
type StrippedSchema struct {
	DOI      string   `json:"doi"`
	Labels   []string `json:"x.labels"`
	SourceID string   `json:"finc.source_id"`
}
