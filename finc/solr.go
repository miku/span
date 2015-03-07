package finc

// SolrSchema represents a finc schema, evolving as needed.
type SolrSchema struct {
	RecordType           string   `json:"recordtype"`
	ID                   string   `json:"id"`
	ISSN                 []string `json:"issn"`
	SourceID             int      `json:"source_id"`
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
