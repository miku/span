// span-folio talks to FOLIO API to fetch ISIL, collections and other
// information relevant to attachments.  Docs:
// https://s3.amazonaws.com/foliodocs/api/mod-finc-config/p/fincConfigMetadataCollections.html
//
// Get metadata collections per ISIL, each "fincConfigMetadataCollections",
// "FilterToCollections", "Filter".
//
// Tenant specific filter. The "finc-select" is special views. Whitelist,
// blacklist filter. EZB holdings is a whitelist. Blacklist predatory journals.
// Filter. How does a filter rule.
//
// Every filter on each collection. Workflow field (testing, approved).
//
// Detour: Regular expressions in RM.
//
// Technical collection identifier to ISIL (tcid => ISIL).
package main

import "flag"

var (
	muFolio = flag.String("folio", "https://example.com/finc-config/metadata-collections", "folio endpoint") // Maybe 100K at once.
	// TODO: Add config file location, also: unify config file handling.
)

func main() {
	flag.Parse()
	// We need the all ISIL, first.
	// Do one query per ISIL.
	// Iterate over all collections.
}

// MetadataCollections folio endpoint response.
type MetadataCollections struct {
	FincConfigMetadataCollections []struct {
		CollectionId string        `json:"collectionId"`
		ContentFiles []interface{} `json:"contentFiles"`
		Description  string        `json:"description"`
		FacetLabel   string        `json:"facetLabel"`
		FreeContent  string        `json:"freeContent"`
		Id           string        `json:"id"`
		Label        string        `json:"label"`
		Lod          struct {
			Note        string `json:"note"`
			Publication string `json:"publication"`
		} `json:"lod"`
		MdSource struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"mdSource"`
		Metadata struct {
			CreatedDate string `json:"createdDate"`
			UpdatedDate string `json:"updatedDate"`
		} `json:"metadata"`
		MetadataAvailable   string        `json:"metadataAvailable"`
		PermittedFor        []string      `json:"permittedFor"`
		SelectedBy          []string      `json:"selectedBy"`
		SolrMegaCollections []string      `json:"solrMegaCollections"`
		Tickets             []interface{} `json:"tickets"`
		UsageRestricted     string        `json:"usageRestricted"`
	} `json:"fincConfigMetadataCollections"`
	TotalRecords int64 `json:"totalRecords"`
}
