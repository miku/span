package doaj

import (
	"encoding/xml"
	"strings"
)

// Record was generated 2019-01-21 14:34:45 by tir on hayiti.
type Record struct {
	XMLName xml.Name `xml:"Record"`
	Text    string   `xml:",chardata"`
	Header  struct {
		Text       string   `xml:",chardata"`
		Status     string   `xml:"status,attr"`
		Identifier string   `xml:"identifier"` // oai:doaj.org/article:72a8...
		Datestamp  string   `xml:"datestamp"`  // 2017-12-31T23:00:18Z, 201...
		SetSpec    []string `xml:"setSpec"`    // TENDOk1lZGljYWwgcGh5c2ljc...
	} `xml:"header"`
	Metadata struct {
		Text string `xml:",chardata"`
		Dc   struct {
			Text           string   `xml:",chardata"`
			SchemaLocation string   `xml:"schemaLocation,attr"`
			Title          string   `xml:"title"`       // Diffusion Tensor Metric M...
			Identifier     []string `xml:"identifier"`  // 1178-623X, 10.4137/MRI.S1...
			Date           string   `xml:"date"`        // 2012-01-01T00:00:00Z, 201...
			Relation       []string `xml:"relation"`    // https://doi.org/10.4137/M...
			Description    string   `xml:"description"` // MRI and Monte Carlo simul...
			Creator        []string `xml:"creator"`     // Jonathan D. Thiessen, Tre...
			Publisher      string   `xml:"publisher"`   // SAGE Publishing, Tripoli ...
			Type           string   `xml:"type"`        // article, article, article...
			Subject        []struct {
				Text string `xml:",chardata"` // Medical physics. Medical ...
				Type string `xml:"type,attr"`
			} `xml:"subject"`
			Language   []string `xml:"language"`   // EN, EN, EN, EN, EN, EN, E...
			Provenance string   `xml:"provenance"` // Journal Licence: CC BY-NC...
			Source     string   `xml:"source"`     // Magnetic Resonance Insigh...
		} `xml:"dc"`
	} `xml:"metadata"`
	About string `xml:"about"`
}

// DOI tries to extract a DOI or returns an empty string.
func (r *Record) DOI() string {
	for _, rel := range r.Metadata.Dc.Relation {
		if !strings.Contains(rel, "doi.org/") {
			continue
		}
		doi := strings.Replace(rel, "https://doi.org/", "", -1)
		return doi
	}
	return ""
}
