// Package finc holds finc SolrSchema (SOLR) and intermediate schema related types and methods.
package finc

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/miku/span/holdings"
	"github.com/miku/span/sets"
)

const AIRecordType = "ai"

var (
	errFromYear    = errors.New("from-year mismatch")
	errFromVolume  = errors.New("from-volume mismatch")
	errFromIssue   = errors.New("from-issue mismatch")
	errToYear      = errors.New("to-year mismatch")
	errToVolume    = errors.New("to-volume mismatch")
	errToIssue     = errors.New("to-issue mismatch")
	errMovingWall  = errors.New("moving-wall violation")
	errNotParsable = errors.New("not parsable")
)

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

type Classification struct {
	Type string   `json:"type"`
	Tags []string `json:"tags"`
}

// IntermediateSchema abstract and collects the values of various input formats.
// Goal is to simplify further processing by using a single format, from which
// the next artifacts can be derived, e.g. records for solr indices.
// This format can be viewed as a catch-all format. The dotted notation
// hints at the origin of the field, e.g. OpenURL, RIS, finc.
type IntermediateSchema struct {
	RecordID       string `json:"finc.record_id"`
	SourceID       int    `json:"finc.source_id"`
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

	RawDate    string `json:"rft.date"` // should be: ISO8601 date
	ParsedDate Date   `json:"date"`

	Place           []string         `json:"rft.place"`
	Publisher       []string         `json:"rft.pub"`
	Edition         string           `json:"rft.edition"`
	Chronology      string           `json:"rft.chron"`
	Season          string           `json:"rft.ssn"`
	Quarter         string           `json:"rft.quarter"`
	Volume          string           `json:"rft.volume"`
	Issue           string           `json:"rft.issue"`
	Part            string           `json:"rft.part"`
	StartPage       string           `json:"rft.spage"`
	EndPage         string           `json:"rft.epage"`
	Pages           string           `json:"rft.pages"`
	PageCount       string           `json:"rft.tpages"`
	ArticleNumber   string           `json:"rft.artnum"`
	ISSN            []string         `json:"rft.issn"`
	EISSN           []string         `json:"rft.eissn"`
	ISBN            []string         `json:"rft.isbn"`
	EISBN           []string         `json:"rft.isbn"`
	Classifications []Classification `json:"x.classification"`

	DOI       string   `json:"doi"`
	URL       []string `json:"url"`
	Authors   []Author `json:"authors"`
	Languages []string `json:"languages"`
	Abstract  string   `json:"abstract"`
}

// ISSNList returns a deduplicated list of all ISSNs.
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

// max of ints.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Date returns a time.Date in a best effort manner. Date parts seem to be always
// present in the source document, while timestamp is only present if
// dateparts consist of all three: year, month and day.
func (is *IntermediateSchema) Date() time.Time {
	return time.Date(max(1, is.ParsedDate.Year), time.Month(max(1, is.ParsedDate.Month)), max(1, is.ParsedDate.Day), 0, 0, 0, 0, time.UTC)
}

// CoveredBy returns nil, if a given entitlement covers the current document.
// If the given entitlement does not cover the document, the error returned
// will contain the reason as message.
// TODO(miku): better handling of 'unparseable' volume or issue strings
func (is *IntermediateSchema) CoveredBy(e holdings.Entitlement) error {
	if e.FromYear != 0 && is.ParsedDate.Year != 0 {
		if e.FromYear > is.ParsedDate.Year {
			return errFromYear
		}
		if e.FromYear == is.ParsedDate.Year {
			volume, err := strconv.Atoi(is.Volume)
			if err != nil {
				return errNotParsable
			}
			if e.FromVolume != 0 && e.FromVolume > volume {
				return errFromVolume
			}
			if e.FromVolume == volume {
				issue, err := strconv.Atoi(is.Issue)
				if err != nil {
					return errNotParsable
				}
				if e.FromIssue != 0 && e.FromIssue > issue {
					return errFromIssue
				}
			}
		}
	}

	if e.ToYear != 0 && is.ParsedDate.Year != 0 {
		if e.ToYear < is.ParsedDate.Year {
			return errToYear
		}
		if e.ToYear == is.ParsedDate.Year {
			volume, err := strconv.Atoi(is.Volume)
			if err != nil {
				return errNotParsable
			}
			if e.ToVolume != 0 && e.ToVolume < volume {
				return errToVolume
			}
			if e.ToVolume == volume {
				issue, err := strconv.Atoi(is.Issue)
				if err != nil {
					return errNotParsable
				}
				if e.ToIssue != 0 && e.ToIssue < issue {
					return errToIssue
				}
			}
		}
	}

	boundary, err := e.Boundary()
	if err != nil {
		return err
	}
	if is.Date().After(boundary) {
		return errMovingWall
	}
	return nil
}

// Institutions returns a slice of ISILs for which this document finds
// valid entitlements in a IsilIssnHolding map.
func (is *IntermediateSchema) Institutions(iih holdings.IsilIssnHolding) []string {
	isils := sets.NewString()
	for _, isil := range iih.Isils() {
		for _, issn := range is.ISSNList() {
			h, exists := iih[isil][issn]
			if !exists {
				continue
			}
			for _, entitlement := range h.Entitlements {
				err := is.CoveredBy(entitlement)
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

// ToSolrSchema converts an intermediate Schema to Solr
func (is *IntermediateSchema) ToSolrSchema() (*SolrSchema, error) {
	output := new(SolrSchema)

	output.RecordType = AIRecordType
	output.Format = is.Format

	output.MegaCollection = append(output.MegaCollection, is.MegaCollection)
	output.SourceID = is.SourceID
	output.ID = is.RecordID

	output.Title = is.ArticleTitle
	output.TitleFull = is.ArticleTitle
	output.TitleShort = is.ArticleTitle

	output.HierarchyParentTitle = is.JournalTitle
	output.PublishDateSort = is.ParsedDate.Year

	if len(is.URL) > 0 {
		output.URL = is.URL[0]
	}

	if len(is.Publisher) > 0 {
		output.Publisher = is.Publisher[0]
	}

	output.ISSN = is.ISSNList()

	if len(is.Classifications) == 1 {
		output.Topics = is.Classifications[0].Tags
	}

	for _, author := range is.Authors {
		output.SecondaryAuthors = append(output.SecondaryAuthors, author.String())
	}

	if len(output.SecondaryAuthors) > 0 {
		output.Author = output.SecondaryAuthors[0]
	}

	output.Fullrecord = "blob:" + output.ID
	return output, nil
}
