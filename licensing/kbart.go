// Package licensing implements support for KBART and ISIL attachments.
// KBART might contains special fields, that are important in certain contexts.
// Example: "Aargauer Zeitung" could not be associated with a record, because
// there is no ISSN. However, there is a string
// "https://www.wiso-net.de/dosearch?&dbShortcut=AGZ" in the record, which could
// be parsed to yield "AGZ", which could be used to relate a record to this entry
// (e.g. if the record has "AGZ" in a certain field, like x.package).
package licensing

import "io"

// Embargo is a string, that expresses a moving wall. A moving wall is a set
// period of time (usually three to five years) between a journal issue's
// publication date and its availability as archival content on [Publisher]. The
// moving wall for each journal is set by the publisher to define the portion of
// its publication history constituting its archive.
type Embargo string

// Entry contains fields about a licensed or available journal, book, article or
// other resource.
//
// User-defined fields:
//
// EZB style:
//
// * own_anchor
// * package:collection
// * il_relevance
// * il_nationwide
// * il_electronic_transmission
// * il_comment
// * all_issns
// * zdb_id
//
// ebooks style:
//
// * location
// * title_notes
// * staff_notes
// * vendor_id
// * oclc_collection_name
// * oclc_collection_id
// * oclc_entry_id
// * oclc_linkscheme
// * oclc_number
// * ACTION
//
// http://www.uksg.org/kbart/s5/guidelines/data_field_labels
type Entry struct {
	PublicationTitle string  // "Südost-Forschungen (2014-)", "Theory of Computation"
	PrintIdentifier  string  // "2029-8692", "9783662479841"
	OnlineIdentifier string  // "1533-8606", "9783834960078"
	FirstIssueDate   string  // "1901", "2008"
	FirstVolume      string  // "1",
	FirstIssue       string  // "1"
	LastIssueDate    string  // "1997", "2008"
	LastVolume       string  // "25"
	LastIssue        string  // "1"
	TitleURL         string  // "http://www.karger.com/dne", "http://link.springer.com/10.1007/978-3-658-15644-2"
	FirstAuthor      string  // "Borgmann", "Wissenschaftlicher Beirat der Bundesregierung Globale Umweltveränderungen (WBGU)"
	TitleID          string  // "22540", "10.1007/978-3-658-10838-0"
	Embargo          Embargo // "P12M", "P1Y", "R20Y"
	CoverageDepth    string  // "Volltext", "ebook"
	CoverageNotes    string  // ...
	PublisherName    string  // "via Hein Online", "Springer (formerly: Kluwer)", "DUV"
	// User-defined fields (EZB).
	OwnAnchor                          string // "elsevier_2016_sax", "UNILEIP", "Wiley Custom 2015"
	PackageCollection                  string // "EBSCO:ebsco_bth", "NALAS:natli_aas2", "NALIW:sage_premier"
	InterlibraryRelevance              string // ...
	InterlibraryNationwide             string // ...
	InterlibraryElectronicTransmission string // Papierkopie an Endnutzer, Elektronischer Versand an Endnutzer
	InterlibraryComment                string // Nur im Inland, il_nationwide
	AllSerialNumbers                   string // 1990-0104;1990-0090, undefined,
	ZDBID                              string // 1459367-1 (see also: http://www.zeitschriftendatenbank.de/suche/zdb-katalog.html)
	// User-defined fields (OCLC).
	Location           string // ...
	TitleNotes         string // ...
	StaffNotes         string // ...
	VendorID           string // ...
	OCLCCollectionName string // "Springer German Language eBooks 2016 - Full Set", "Wiley Online Library UBCM All Obooks"
	OCLCCollectionID   string // "springerlink.de2011fullset", "wiley.ubcmall"
	OCLCEntryID        string // "25106066"
	OCLCLinkScheme     string // "wiley.book"
	OCLCNumber         string // "122938128"
	Action             string // "raw"
}

// Holdings contains a list of entries about licenced or available content.
type Holdings struct {
	Entries []Entry
}

// ReadFrom create holdings struct from a reader. Expects a tab separated CSV with
// a single header line.
func (h *Holdings) ReadFrom(r io.Reader) (int64, error) {
	return 0, nil
}
