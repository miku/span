package finc

import "time"

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

	BookTitle    string   `json:"rft.btitle,omitempty"`
	Chronology   string   `json:"rft.chron,omitempty"`
	Edition      string   `json:"rft.edition,omitempty"`
	EISBN        []string `json:"rft.isbn,omitempty"`
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
}
