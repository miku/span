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

// AddInstitution adds an isil, if it is not already there.
func (s *SolrSchema) AddInstitution(isil string) {
	for _, institution := range s.Institutions {
		if institution == isil {
			return
		}
	}
	s.Institutions = append(s.Institutions, isil)
}

// AddMegaCollection adds isil, if it is not already there.
func (s *SolrSchema) AddMegaCollection(collection string) {
	for _, c := range s.MegaCollection {
		if c == collection {
			return
		}
	}
	s.MegaCollection = append(s.MegaCollection, collection)
}

// Author representes an author, inspired by OpenURL.
type Author struct {
	ID           string `json:"id"`
	Name         string `json:"rft.au"`
	LastName     string `json:"rft.aulast"`
	FirstName    string `json:"rft.aufirst"`
	Initial      string `json:"rft.auinit"`
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

// Schema is an intermediate format inspired by a few existing formats, e.g. OpenURL.
type Schema struct {
	RecordID       string `json:"finc.record_id"`
	SourceID       string `json:"finc.source_id"`
	MegaCollection string `json:"finc.mega_collection"`

	Genre  string `json:"rft.genre"`
	Type   string `json:"ris.type"`
	Format string `json:"finc.format"`

	ArticleTitle string `json:"rft.atitle"`
	BookTitle    string `json:"rft.btitle"`
	JournalTitle string `json:"rft.jtitle"`
	SeriesTitle  string `json:"rft.stitle"`

	Series string `json:"rft.series"`

	Database     string `json:"ris.db"`
	DataProvider string `json:"ris.dp"`

	Date          string   `json:"rft.date"`
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
	ISBN          string   `json:"rft.isbn"`
	EISBN         []string `json:"rft.isbn"`

	DOI      string   `json:"doi"`
	URL      []string `json:"url"`
	Authors  []Author `json:"authors"`
	Language string   `json:"language"`
	Abstract string   `json:"abstract"`
}
