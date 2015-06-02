package finc

import (
	"fmt"

	"github.com/kennygrant/sanitize"
	"github.com/miku/span/container"
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

// DummySchema is an example export schema, that only has one field.
type DummySchema struct {
	Title string `json:"title"`
}

// Attach is here, so it satisfies the interface, but implementation is a noop.
func (d *DummySchema) Attach(s []string) {}

// Export converts intermediate schema into this export schema.
func (d *DummySchema) Convert(is IntermediateSchema) error {
	d.Title = fmt.Sprintf("%s (%s)", is.ArticleTitle, is.JournalTitle)
	return nil
}

// Solr413Schema is the default solr 4 schema. It is based on VuFind 1.3, it
// does not contain `container_*` fields.
type Solr413Schema struct {
	AccessFacet          string   `json:"access_facet,omitempty"`
	AuthorFacet          []string `json:"author_facet"`
	Allfields            string   `json:"allfields,omitempty"`
	Author               string   `json:"author,omitempty"`
	FincClassFacet       []string `json:"finc_class_facet,omitempty"`
	Formats              []string `json:"format,omitempty"`
	Fullrecord           string   `json:"fullrecord,omitempty"`
	Fulltext             string   `json:"fulltext,omitempty"`
	HierarchyParentTitle []string `json:"hierarchy_parent_title,omitempty"`
	ID                   string   `json:"id,omitempty"`
	Institutions         []string `json:"institution,omitempty"`
	Imprint              string   `json:"imprint,omitempty"`
	ISSN                 []string `json:"issn,omitempty"`
	Languages            []string `json:"language,omitempty"`
	MegaCollections      []string `json:"mega_collection,omitempty"`
	PublishDateSort      int      `json:"publishDateSort,omitempty"`
	Publishers           []string `json:"publisher,omitempty"`
	RecordType           string   `json:"recordtype,omitempty"`
	Series               []string `json:"series,omitempty"`
	SecondaryAuthors     []string `json:"author2,omitempty"`
	SourceID             string   `json:"source_id,omitempty"`
	Subtitle             string   `json:"title_sub,omitempty"`
	Title                string   `json:"title,omitempty"`
	TitleFull            string   `json:"title_full,omitempty"`
	TitleShort           string   `json:"title_short,omitempty"`
	TitleSort            string   `json:"title_sort,omitempty"`
	Topics               []string `json:"topic,omitempty"`
	URL                  []string `json:"url,omitempty"`
	FormatDe15           []string `json:"format_de15"`
}

// Attach attaches the ISILs to a record.
func (s *Solr413Schema) Attach(isils []string) {
	s.Institutions = isils
}

// Export method from intermediate schema to solr 4/13 schema.
func (s *Solr413Schema) Convert(is IntermediateSchema) error {
	s.Allfields = is.Allfields()
	s.Formats = append(s.Formats, is.Format)
	s.Fullrecord = "blob:" + is.RecordID
	s.Fulltext = is.Fulltext
	s.HierarchyParentTitle = append(s.HierarchyParentTitle, is.JournalTitle)
	s.ID = is.RecordID
	s.Imprint = is.Imprint()
	s.ISSN = is.ISSNList()
	s.MegaCollections = append(s.MegaCollections, is.MegaCollection)
	s.PublishDateSort = is.Date.Year()
	s.Publishers = is.Publishers
	s.RecordType = AIRecordType
	s.Series = append(s.Series, is.JournalTitle)
	s.SourceID = is.SourceID
	s.Subtitle = is.ArticleSubtitle
	s.TitleSort = is.SortableTitle()
	s.Topics = is.Subjects
	s.URL = is.URL

	classes := container.NewStringSet()
	for _, s := range is.Subjects {
		for _, class := range SubjectMapping.LookupDefault(s, []string{}) {
			classes.Add(class)
		}
	}
	s.FincClassFacet = classes.Values()

	sanitized := sanitize.HTML(is.ArticleTitle)
	s.Title, s.TitleFull, s.TitleShort = sanitized, sanitized, sanitized

	for _, lang := range is.Languages {
		s.Languages = append(s.Languages, LanguageMap.LookupDefault(lang, lang))
	}

	for _, author := range is.Authors {
		s.SecondaryAuthors = append(s.SecondaryAuthors, author.String())
		s.AuthorFacet = append(s.AuthorFacet, author.String())
	}

	if len(s.SecondaryAuthors) > 0 {
		s.Author = s.SecondaryAuthors[0]
	}

	s.AccessFacet = AIAccessFacet
	s.FormatDe15 = []string{FormatSite.LookupDefault(is.Format, "")}

	return nil
}
