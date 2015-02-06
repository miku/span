package finc

// Schema represents a finc schema, evolving as needed
type Schema struct {
	RecordType           string   `json:"recordtype"`
	ID                   string   `json:"id"`
	ISSN                 []string `json:"issn"`
	Publisher            string   `json:"publisher"`
	SourceID             string   `json:"source_id"`
	Title                string   `json:"title"`
	TitleFull            string   `json:"title_full"`
	TitleShort           string   `json:"title_short"`
	Topic                []string `json:"topic"`
	URL                  string   `json:"url"`
	HierarchyParentTitle string   `json:"hierarchy_parent_title"`
	Format               string   `json:"format"`
	SecondaryAuthors     []string `json:"author2"`
	PublishDateSort      int      `json:"publishDateSort"`
	Allfields            string   `json:"allfields"`
	Institution          []string `json:"institution"`
}

func (s *Schema) AddInstitution(isil string) {
	for _, institution := range s.Institution {
		if institution == isil {
			return
		}
	}
	s.Institution = append(s.Institution, isil)
}
