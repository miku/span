package exporter

import (
	"github.com/kennygrant/sanitize"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
)

// Solr4vu13v2 is the default solr 4 schema. It contains more site-specific
// formats and support for container_* fields.
type Solr4Vufind13v2 struct {
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

	// New container volumes.
	ContainerVolume    string `json:"container_volume,omitempty"`
	ContainerIssue     string `json:"container_issue,omitempty"`
	ContainerStartPage string `json:"container_start_page,omitempty"`
	ContainerTitle     string `json:"container_title,omitempty"`

	// Site specific formats, TODO(miku): abstract this away
	FormatDe105  []string `json:"format_de105,omitempty"`
	FormatDe15   []string `json:"format_de15,omitempty"`
	FormatDe520  []string `json:"format_de520,omitempty"`
	FormatDe540  []string `json:"format_de540,omitempty"`
	FormatDeCh1  []string `json:"format_dech1,omitempty"`
	FormatDed117 []string `json:"format_ded117,omitempty"`
	FormatDel152 []string `json:"format_del152,omitempty"`
	FormatDel189 []string `json:"format_del189,omitempty"`
	FormatDeZi4  []string `json:"format_dezi4,omitempty"`
	FormatDeZwi2 []string `json:"format_dezwi2,omitempty"`
	FormatDeGla1 []string `json:"format_degla1,omitempty"`
}

// Attach attaches the ISILs to a record.
func (s *Solr4Vufind13v2) Attach(isils []string) {
	s.Institutions = isils
}

// Export method from intermediate schema to solr 4/13 schema.
func (s *Solr4Vufind13v2) Convert(is finc.IntermediateSchema) error {
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
	s.RecordType = finc.AIRecordType
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
	s.FormatDe15 = []string{FormatDe15.LookupDefault(is.Format, "")}
	s.FormatDeGla1 = []string{FormatDeGla1.LookupDefault(is.Format, "")}

	s.ContainerVolume = is.Volume
	s.ContainerIssue = is.Issue
	s.ContainerStartPage = is.StartPage
	s.ContainerTitle = is.JournalTitle

	return nil
}
