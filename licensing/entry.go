// Package licensing implements support for KBART and ISIL attachments.
// KBART might contains special fields, that are important in certain contexts.
// Example: "Aargauer Zeitung" could not be associated with a record, because
// there is no ISSN. However, there is a string
// "https://www.wiso-net.de/dosearch?&dbShortcut=AGZ" in the record, which could
// be parsed to yield "AGZ", which could be used to relate a record to this entry
// (e.g. if the record has "AGZ" in a certain field, like x.package).
package licensing

import (
	"sort"

	"github.com/miku/span"
	"github.com/miku/span/container"
)

// Embargo is a string, that expresses a moving wall. A moving wall is a set
// period of time (usually three to five years) between a journal issue's
// publication date and its availability as archival content on [Publisher]. The
// moving wall for each journal is set by the publisher to define the portion of
// its publication history constituting its archive.
type Embargo string

// Entry contains fields about a licensed or available journal, book, article or
// other resource. First 14 columns are quite stardardized. Further columns may
// contain custom information:
//
// EZB style: own_anchor, package:collection, il_relevance, il_nationwide,
// il_electronic_transmission, il_comment, all_issns, zdb_id
//
// OCLC style: location, title_notes, staff_notes, vendor_id,
// oclc_collection_name, oclc_collection_id, oclc_entry_id, oclc_linkscheme,
// oclc_number, ACTION
//
// See also: http://www.uksg.org/kbart/s5/guidelines/data_field_labels,
// http://www.uksg.org/kbart/s5/guidelines/data_fields
type Entry struct {
	PublicationTitle                   string `csv:"publication_title"`          // "Südost-Forschungen (2014-)", "Theory of Computation"
	PrintIdentifier                    string `csv:"print_identifier"`           // "2029-8692", "9783662479841"
	OnlineIdentifier                   string `csv:"online_identifier"`          // "1533-8606", "9783834960078"
	FirstIssueDate                     string `csv:"date_first_issue_online"`    // "1901", "2008"
	FirstVolume                        string `csv:"num_first_vol_online"`       // "1",
	FirstIssue                         string `csv:"num_first_issue_online"`     // "1"
	LastIssueDate                      string `csv:"date_last_issue_online"`     // "1997", "2008"
	LastVolume                         string `csv:"num_last_vol_online"`        // "25"
	LastIssue                          string `csv:"num_last_issue_online"`      // "1"
	TitleURL                           string `csv:"title_url"`                  // "http://www.karger.com/dne", "http://link.springer.com/10.1007/978-3-658-15644-2"
	FirstAuthor                        string `csv:"first_author"`               // "Borgmann", "Wissenschaftlicher Beirat der Bundesregierung Globale Umweltveränderungen (WBGU)"
	TitleID                            string `csv:"title_id"`                   // "22540", "10.1007/978-3-658-10838-0"
	Embargo                            string `csv:"embargo_info"`               // "P12M", "P1Y", "R20Y"
	CoverageDepth                      string `csv:"coverage_depth"`             // "Volltext", "ebook"
	CoverageNotes                      string `csv:"coverage_notes"`             // ...
	PublisherName                      string `csv:"publisher_name"`             // "via Hein Online", "Springer (formerly: Kluwer)", "DUV"
	OwnAnchor                          string `csv:"own_anchor"`                 // "elsevier_2016_sax", "UNILEIP", "Wiley Custom 2015"
	PackageCollection                  string `csv:"package:collection"`         // "EBSCO:ebsco_bth", "NALAS:natli_aas2", "NALIW:sage_premier"
	InterlibraryRelevance              string `csv:"il_relevance"`               // ...
	InterlibraryNationwide             string `csv:"il_nationwide"`              // ...
	InterlibraryElectronicTransmission string `csv:"il_electronic_transmission"` // Papierkopie an Endnutzer, Elektronischer Versand an Endnutzer
	InterlibraryComment                string `csv:"il_comment"`                 // Nur im Inland, il_nationwide
	AllSerialNumbers                   string `csv:"all_issns"`                  // 1990-0104;1990-0090, undefined,
	ZDBID                              string `csv:"zdb_id"`                     // 1459367-1 (see also: http://www.zeitschriftendatenbank.de/suche/zdb-katalog.html)
	Location                           string `csv:"location"`                   // ...
	TitleNotes                         string `csv:"title_notes"`                // ...
	StaffNotes                         string `csv:"staff_notes"`                // ...
	VendorID                           string `csv:"vendor_id"`                  // ...
	OCLCCollectionName                 string `csv:"oclc_collection_name"`       // "Springer German Language eBooks 2016 - Full Set", "Wiley Online Library UBCM All Obooks"
	OCLCCollectionID                   string `csv:"oclc_collection_id"`         // "springerlink.de2011fullset", "wiley.ubcmall"
	OCLCEntryID                        string `csv:"oclc_entry_id"`              // "25106066"
	OCLCLinkScheme                     string `csv:"oclc_link_scheme"`           // "wiley.book"
	OCLCNumber                         string `csv:"oclc_number"`                // "122938128"
	Action                             string `csv:"ACTION"`                     // "raw"
}

// ISSNList returns a list of all ISSN from various fields.
func (e *Entry) ISSNList() []string {
	issns := container.NewStringSet()
	if span.ISSNPattern.MatchString(e.PrintIdentifier) {
		issns.Add(e.PrintIdentifier)
	}
	if span.ISSNPattern.MatchString(e.OnlineIdentifier) {
		issns.Add(e.OnlineIdentifier)
	}
	for _, issn := range span.ISSNPattern.FindAllString(e.OwnAnchor, -1) {
		issns.Add(issn)
	}
	v := issns.Values()
	sort.Strings(v)
	return v
}
