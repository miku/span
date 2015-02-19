// Package finc holds finc SolrSchema (SOLR) related types and methods.
package finc

// Schema represents a finc schema, evolving as needed
type SolrSchema struct {
	RecordType           string   `json:"recordtype"`
	ID                   string   `json:"id"`
	ISSN                 []string `json:"issn"`
	SourceID             string   `json:"source_id"`
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
}

// AddInstitution adds an isil, if it is not already there.
func (s *SolrSchema) AddInstitution(isil string) {
	for _, institution := range s.Institutions {
		if institution == isil {
			return
		}
	}
	s.Institutions = append(s.Institutions, isil)
}

// AddMegaCollection adds isil, if it is not already there.
func (s *SolrSchema) AddMegaCollection(collection string) {
	for _, c := range s.MegaCollection {
		if c == collection {
			return
		}
	}
	s.MegaCollection = append(s.MegaCollection, collection)
}
