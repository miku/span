package doaj

import (
	"encoding/xml"
	"time"

	"github.com/miku/span/formats/finc"
)

// Record was generated 2019-03-07 22:40:57 by tir on hayiti.
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
			Type           string   `xml:"type"`        // article
			Subject        []struct {
				Text string `xml:",chardata"` // Medical physics. Medical ...
				Type string `xml:"type,attr"`
			} `xml:"subject"`
			Language   []string `xml:"language"`   // EN
			Provenance string   `xml:"provenance"` // Journal Licence: CC BY-NC...
			Source     string   `xml:"source"`     // Magnetic Resonance Insigh...
		} `xml:"dc"`
	} `xml:"metadata"`
	About string `xml:"about"`
}

// Date tries to parse the date.
func (record Record) Date() (time.Time, error) {
	// <dc:date>2012-01-01T00:00:00Z</dc:date>
	return time.Parse("2006-01-02T15:04:05Z")
}

// ToIntermediateSchema converts OAI record to intermediate schema.
func (record Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error

	output := finc.NewIntermediateSchema()
	date, err := record.Date()
	if err != nil {
		return output, err
	}
	output.Date = date
	output.RawDate = date.Format("2006-01-02")

	return output
}
