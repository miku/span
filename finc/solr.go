package finc

// SolrSchema represents a finc schema, evolving as needed.
type SolrSchema struct {
	AccessFacet          string   `json:"access_facet"`
	Allfields            string   `json:"allfields"`
	Author               string   `json:"author"`
	FincClassFacet       []string `json:"finc_class_facet"`
	Formats              []string `json:"format"`
	Fullrecord           string   `json:"fullrecord"`
	HierarchyParentTitle string   `json:"hierarchy_parent_title"`
	ID                   string   `json:"id"`
	Institutions         []string `json:"institution"`
	ISSN                 []string `json:"issn"`
	Languages            []string `json:"language"`
	MegaCollection       []string `json:"mega_collection"`
	PublishDateSort      int      `json:"publishDateSort"`
	Publishers           []string `json:"publisher"`
	RecordType           string   `json:"recordtype"`
	SecondaryAuthors     []string `json:"author2"`
	SourceID             string   `json:"source_id"`
	Title                string   `json:"title"`
	TitleFull            string   `json:"title_full"`
	TitleShort           string   `json:"title_short"`
	Topics               []string `json:"topic"`
	URL                  string   `json:"url"`
}
