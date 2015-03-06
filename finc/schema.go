// Package finc holds finc SolrSchema (SOLR) and intermediate schema related types and methods.
package finc

import "fmt"

// SolrSchema represents a finc schema, evolving as needed.
type SolrSchema struct {
	RecordType           string   `json:"recordtype"`
	ID                   string   `json:"id"`
	ISSN                 []string `json:"issn"`
	SourceID             string   `json:"source_id"`
	Title                string   `json:"title"`
	TitleFull            string   `json:"title_full"`
	TitleShort           string   `json:"title_short"`
	Topics               []string `json:"topic"`
	URL                  string   `json:"url"`
	Publisher            string   `json:"publisher"`
	HierarchyParentTitle string   `json:"hierarchy_parent_title"`
	Format               string   `json:"format"`
	Author               string   `json:"author"`
	SecondaryAuthors     []string `json:"author2"`
	PublishDateSort      int      `json:"publishDateSort"`
	Allfields            string   `json:"allfields"`
	Institutions         []string `json:"institution"`
	MegaCollection       []string `json:"mega_collection"`
	Fullrecord           string   `json:"fullrecord"`
}

type AuthorID struct {
	ID   string // ...
	Type string // gnd, orcid, ...
}

// Author representes an author, inspired by OpenURL.
type Author struct {
	ID           string `json:"id"` // eventuell mehrfach, mit typ
	Name         string `json:"rft.au"`
	LastName     string `json:"rft.aulast"`
	FirstName    string `json:"rft.aufirst"`
	Initial      string `json:"rft.auinit"` // was ist unterschied zu auinit1?
	FirstInitial string `json:"rft.auinit1"`
	MiddleName   string `json:"rft.auinitm"`
	Suffix       string `json:"rft.ausuffix"`
	Corporation  string `json:"rft.aucorp"`
}

// String returns a formatted author string.
// TODO(miku): make this complete.
func (author *Author) String() string {
	if author.LastName != "" {
		if author.FirstName != "" {
			return fmt.Sprintf("%s, %s", author.LastName, author.FirstName)
		}
		return author.LastName
	}
	return author.ID
}

type Date struct {
	Year      int `json:"year"`
	Month     int `json:"month"`
	Day       int `json:"year"`
	Timestamp int `json:"timestamp"`
}

// IntermediateSchema abstract and collects the values of various input formats.
// Goal is to simplify further processing by using a single format, from which
// the next artifacts can be derived, e.g. records for solr indices.
// This format can be viewed as a catch-all format. The dotted notation
// hints at the origin of the field, e.g. OpenURL, RIS, finc.
type IntermediateSchema struct {
	RecordID       string `json:"finc.record_id"`
	SourceID       string `json:"finc.source_id"`
	MegaCollection string `json:"finc.mega_collection"`
	Format         string `json:"finc.format"`

	Genre string `json:"rft.genre"`
	Type  string `json:"ris.type"` // RIS-Format? Wikipedia...

	ArticleTitle    string `json:"rft.atitle"`
	ArticleSubtitle string `json:"x.subtitle"`
	BookTitle       string `json:"rft.btitle"`
	JournalTitle    string `json:"rft.jtitle"`
	SeriesTitle     string `json:"rft.stitle"`

	Series string `json:"rft.series"`

	Database     string `json:"ris.db"`
	DataProvider string `json:"ris.dp"`

	Date          string   `json:"rft.date"` // ISO8601 date
	Place         []string `json:"rft.place"`
	Publisher     []string `json:"rft.pub"`
	Edition       string   `json:"rft.edition"`
	Chronology    string   `json:"rft.chron"`
	Season        string   `json:"rft.ssn"`
	Quarter       string   `json:"rft.quarter"`
	Volume        string   `json:"rft.volume"`
	Issue         string   `json:"rft.issue"`
	Part          string   `json:"rft.part"`
	StartPage     string   `json:"rft.spage"`
	EndPage       string   `json:"rft.epage"`
	Pages         string   `json:"rft.pages"`
	PageCount     string   `json:"rft.tpages"`
	ArticleNumber string   `json:"rft.artnum"`
	ISSN          []string `json:"rft.issn"`
	EISSN         []string `json:"rft.eissn"`
	ISBN          []string `json:"rft.isbn"`
	EISBN         []string `json:"rft.isbn"`

	DOI       string   `json:"doi"`
	URL       []string `json:"url"`
	Authors   []Author `json:"authors"`
	Languages []string `json:"languages"`
	Abstract  string   `json:"abstract"`
}
