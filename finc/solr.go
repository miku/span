package finc

// SolrSchema represents a finc schema, evolving as needed.
type SolrSchema struct {
	AccessFacet          string   `json:"access_facet,omitempty"`
	Allfields            string   `json:"allfields,omitempty"`
	Author               string   `json:"author,omitempty"`
	FincClassFacets      []string `json:"finc_class_facet,omitempty"`
	Formats              []string `json:"format,omitempty"`
	Fullrecord           string   `json:"fullrecord,omitempty"`
	Fulltext             string   `json:"fulltext,omitempty"`
	HierarchyParentTitle string   `json:"hierarchy_parent_title,omitempty"`
	ID                   string   `json:"id,omitempty"`
	Institutions         []string `json:"institution,omitempty"`
	Imprint              string   `json:"imprint,omitempty"`
	ISSN                 []string `json:"issn,omitempty"`
	Languages            []string `json:"language,omitempty"`
	MegaCollections      []string `json:"mega_collection,omitempty"`
	PublishDateSort      int      `json:"publishDateSort,omitempty"`
	Publishers           []string `json:"publisher,omitempty"`
	RecordType           string   `json:"recordtype,omitempty"`
	SecondaryAuthors     []string `json:"author2,omitempty"`
	SourceID             string   `json:"source_id,omitempty"`
	Title                string   `json:"title,omitempty"`
	TitleFull            string   `json:"title_full"`
	TitleShort           string   `json:"title_short,omitempty"`
	Topics               []string `json:"topic,omitempty"`
	URL                  []string `json:"url,omitempty"`
}
