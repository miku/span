// TODO.
package finc

import (
	"fmt"
	"strings"

	"github.com/segmentio/encoding/json"

	"github.com/kennygrant/sanitize"
	"github.com/miku/span/container"
)

// Solr5Vufind3 is the basic solr 5 schema as of 2016-04-14. It is based on
// VuFind 3. Same as Solr5Vufind3v12, but with fullrecord field, refs. #8031.
// TODO(martin): add support for classfinc.toml
type Solr5Vufind3 struct {
	AccessFacet          string   `json:"access_facet,omitempty"`
	AuthorFacet          []string `json:"author_facet,omitempty"`
	AuthorCorporate      []string `json:"author_corporate,omitempty"`
	Authors              []string `json:"author,omitempty"`
	AuthorSort           string   `json:"author_sort,omitempty"`
	SecondaryAuthors     []string `json:"author2,omitempty"`
	Allfields            string   `json:"allfields,omitempty"`
	DOI                  []string `json:"doi_str_mv,omitempty"` // recommended via https://vufind.org/wiki/development:architecture:solr_index_schema
	Edition              string   `json:"edition,omitempty"`
	FacetAvail           []string `json:"facet_avail"`
	FincClassFacet       []string `json:"finc_class_facet,omitempty"`
	Footnotes            []string `json:"footnote,omitempty"`
	Formats              []string `json:"format,omitempty"`
	Fullrecord           string   `json:"fullrecord,omitempty"`
	Fulltext             string   `json:"fulltext,omitempty"`
	HierarchyParentTitle []string `json:"hierarchy_parent_title,omitempty"`
	ID                   string   `json:"id,omitempty"`
	Institutions         []string `json:"institution,omitempty"`
	Imprint              string   `json:"imprint,omitempty"`
	ImprintStrMv         []string `json:"imprint_str_mv,omitempty"`
	ISSN                 []string `json:"issn,omitempty"`
	ISSNStrMv            []string `json:"issn_str_mv,omitempty"` // refs. #21393
	ISBN                 []string `json:"isbn,omitempty"`
	ISBNStrMv            []string `json:"isbn_str_mv,omitempty"` // refs. #21393
	Languages            []string `json:"language,omitempty"`
	MegaCollections      []string `json:"mega_collection,omitempty"`
	MatchStr             string   `json:"match_str"`    // do not omit, refs. #21403#note-15
	MatchStrMv           []string `json:"match_str_mv"` // do not omit, refs. #21403#note-15
	PublishDateSort      int      `json:"publishDateSort,omitempty"`
	Publishers           []string `json:"publisher,omitempty"`
	RecordID             string   `json:"record_id,omitempty"`
	RecordType           string   `json:"recordtype,omitempty"`
	Series               []string `json:"series,omitempty"`
	SourceID             string   `json:"source_id,omitempty"`
	Subtitle             string   `json:"title_sub,omitempty"`
	Title                string   `json:"title,omitempty"`
	TitleFull            string   `json:"title_full,omitempty"`
	TitleShort           string   `json:"title_short,omitempty"`
	TitleSort            string   `json:"title_sort,omitempty"`
	Topics               []string `json:"topic,omitempty"`
	URL                  []string `json:"url,omitempty"`
	PublishDate          []string `json:"publishDate,omitempty"`
	Physical             []string `json:"physical,omitempty"`
	Description          string   `json:"description"`
	Collections          []string `json:"collection"` // index/wiki/Kollektionsfacette

	ContainerIssue     string `json:"container_issue,omitempty"`
	ContainerStartPage string `json:"container_start_page,omitempty"`
	ContainerTitle     string `json:"container_title,omitempty"`
	ContainerVolume    string `json:"container_volume,omitempty"`

	FormatDe105  []string `json:"format_de105,omitempty"`
	FormatDe14   []string `json:"format_de14,omitempty"`
	FormatDe15   []string `json:"format_de15,omitempty"`
	FormatDe520  []string `json:"format_de520,omitempty"`
	FormatDe540  []string `json:"format_de540,omitempty"`
	FormatDeCh1  []string `json:"format_dech1,omitempty"`
	FormatDed117 []string `json:"format_ded117,omitempty"`
	FormatDeGla1 []string `json:"format_degla1,omitempty"`
	FormatDel152 []string `json:"format_del152,omitempty"`
	FormatDel189 []string `json:"format_del189,omitempty"`
	FormatDeZi4  []string `json:"format_dezi4,omitempty"`
	FormatDeZwi2 []string `json:"format_dezwi2,omitempty"`
	FormatFinc   []string `json:"format_finc,omitempty"`
	FormatNrw    []string `json:"format_nrw,omitempty"`
	BranchNrw    string   `json:"branch_nrw,omitempty"` // refs #11605
}

// Export fulfuls finc.Exporter interface, so we can plug this into cmd/span-export. Takes
// an intermediate schema and returns serialized JSON.
func (s *Solr5Vufind3) Export(is IntermediateSchema, withFullrecord bool) ([]byte, error) {
	if err := s.convert(is, withFullrecord); err != nil {
		return []byte{}, err
	}
	return json.Marshal(s)
}

// convert converts intermediate schema to the Solr5Vufind3. The struct fields are populated.
func (s *Solr5Vufind3) convert(is IntermediateSchema, withFullrecord bool) error {
	s.Allfields = is.Allfields()
	s.Formats = append(s.Formats, is.Format)
	s.Fullrecord = "blob:" + is.ID
	s.Fulltext = is.Fulltext
	s.ID = is.ID
	s.RecordID = is.RecordID
	s.Imprint = is.Imprint()
	s.ImprintStrMv = []string{s.Imprint}
	s.ISSN = is.ISSNList()
	s.ISSNStrMv = s.ISSN
	s.ISBN = is.ISBNList()
	s.ISBNStrMv = s.ISBN
	s.Edition = is.Edition
	for _, name := range is.MegaCollections {
		// As per 2020-06-30 try to keep tcids (sid-...) in SOLR collection
		// field, and labels in SOLR mega_collection. Except with crossref
		// (49), where we do not have tcids (yet). As of 2021-04-20, we want
		// [49] collection names in solr mega collections.
		if strings.HasPrefix(name, "sid-") {
			s.Collections = append(s.Collections, name)
		} else {
			// refs. #18495
			s.MegaCollections = append(s.MegaCollections, name)
		}
	}
	s.PublishDateSort = is.Date.Year()
	s.PublishDate = []string{is.Date.Format("2006")} // refs #18608
	s.Publishers = is.Publishers
	if withFullrecord {
		s.RecordType = IntermediateSchemaRecordType
	} else {
		s.RecordType = AIRecordType
	}

	if is.JournalTitle != "" {
		s.Series = append(s.Series, is.JournalTitle)
	}
	if is.Series != "" {
		s.Series = append(s.Series, is.Series)
	}

	s.SourceID = is.SourceID
	s.Subtitle = is.ArticleSubtitle
	s.TitleSort = is.SortableTitle()
	s.Topics = is.Subjects

	// refs. #12127
	s.URL = is.URL

	// refs. #8709
	if is.DOI != "" {
		s.DOI = []string{is.DOI}
		containsDOI := false
		for _, u := range s.URL {
			if strings.Contains(u, "doi") {
				containsDOI = true
			}
		}
		if !containsDOI {
			// refs. GH #9
			s.URL = append(s.URL, fmt.Sprintf("https://doi.org/%s", is.DOI))
		}
	}

	classes := container.NewStringSet()
	for _, s := range is.Subjects {
		for _, class := range SubjectMapping.Lookup(s, []string{}) {
			classes.Add(class)
		}
	}
	s.FincClassFacet = classes.Values()

	var sanitized string
	switch {
	// refs #13024, book title shall not shadow article title (if both are
	// given, keep the article title).
	case is.BookTitle != "" && is.ArticleTitle == "":
		sanitized = sanitize.HTML(is.BookTitle)
	default:
		sanitized = sanitize.HTML(is.ArticleTitle)
	}

	s.Title, s.TitleFull, s.TitleShort = sanitized, sanitized, sanitized

	// In intermediate schema we do not have a title yet but rft.btitle is
	// non-empty, use that.
	if s.Title == "" && is.BookTitle != "" {
		sanitized := sanitize.HTML(is.BookTitle)
		s.Title, s.TitleFull, s.TitleShort = sanitized, sanitized, sanitized
	}

	for _, lang := range is.Languages {
		s.Languages = append(s.Languages, LanguageMap.Lookup(lang, lang))
	}

	// TODO(miku): What's with author_corp_ref, https://goo.gl/sx1s3r
	// Verweisungsform aus den Normdaten für Körperschaften und Kongresse bei
	// Marc-Quellen, die entsprechend angereichert wurden (SWB, GBV)
	// grundsätzlich aber bei allen Datenquellen möglich, welche solche
	// Informationen enthalten bzw. referenzieren sollte zukünftig ggf. auch in
	// author_corporate_ref umbenannt werden"

	// Collect sanitized authors.
	var authors []string

	// Refs. https://github.com/miku/span/issues/12.
	var authorCorporate []string

	for _, author := range is.Authors {
		sanitized := AuthorReplacer.Replace(author.String())
		if sanitized == "" {
			// Refs. https://github.com/miku/span/issues/12.
			if author.Corporate != "" {
				authorCorporate = append(authorCorporate, author.Corporate)
			}
			continue
		}
		authors = append(authors, sanitized)
		s.AuthorFacet = append(s.AuthorFacet, sanitized)
	}

	if len(authorCorporate) > 0 {
		s.AuthorCorporate = authorCorporate
	}

	// refs #7092, gh #8, refs #12310
	if len(authors) > 0 {
		s.Authors = authors
		s.AuthorSort = strings.ToLower(authors[0])
	}

	s.AccessFacet = AIAccessFacet
	s.BranchNrw = s.AccessFacet // refs #11605

	// Site specific formats, TODO: fix this now.
	s.FormatDe105 = []string{FormatDe105.Lookup(is.Format, "")}
	s.FormatDe14 = []string{FormatDe14.Lookup(is.Format, "")}
	s.FormatDe15 = []string{FormatDe15.Lookup(is.Format, "")}
	s.FormatDe520 = []string{FormatDe520.Lookup(is.Format, "")}
	s.FormatDe540 = []string{FormatDe540.Lookup(is.Format, "")}
	s.FormatDeCh1 = []string{FormatDeCh1.Lookup(is.Format, "")}
	s.FormatDed117 = []string{FormatDed117.Lookup(is.Format, "")}
	s.FormatDeGla1 = []string{FormatDeGla1.Lookup(is.Format, "")}
	s.FormatDel152 = []string{FormatDel152.Lookup(is.Format, "")}
	s.FormatDel189 = []string{FormatDel189.Lookup(is.Format, "")}
	s.FormatDeZi4 = []string{FormatDeZi4.Lookup(is.Format, "")}
	s.FormatDeZwi2 = []string{FormatDeZwi2.Lookup(is.Format, "")}
	s.FormatNrw = []string{FormatNrw.Lookup(is.Format, "")}
	// TODO(miku): Is this correct?
	s.FormatFinc = []string{FormatFinc.Lookup(is.Format, "")}

	s.ContainerVolume = is.Volume
	s.ContainerIssue = is.Issue
	s.ContainerStartPage = is.StartPage
	s.ContainerTitle = is.JournalTitle

	s.Institutions = is.Labels
	s.Description = is.Abstract

	if withFullrecord {
		// refs. #8031
		b, err := json.Marshal(is)
		if err != nil {
			return err
		}
		s.Fullrecord = string(b)
	}

	// Default facet for online contents, refs #11285.
	s.FacetAvail = []string{"Online"}
	if is.OpenAccess {
		s.FacetAvail = append(s.FacetAvail, "Free")
	}

	// refs #11478
	s.Physical = []string{is.Pages}

	// refs #14215
	if is.SourceID == "48" {
		s.ISSN = []string{}
		// s.Description = "" // refs #14787
		s.Languages = []string{}
		s.Publishers = []string{}
		s.Fulltext = ""
	}

	s.Footnotes = is.Footnotes
	s.MatchStr = ""
	s.MatchStrMv = []string{}

	return nil
}
