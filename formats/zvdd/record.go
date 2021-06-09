package zvdd

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/miku/span/formats/finc"
)

// SourceIdentifier for internal bookkeeping.
const (
	SourceIdentifier = "93" // TODO(miku): find correct id
	Format           = "ElectronicArticle"
	Genre            = "article"
	DefaultRefType   = "EJOUR"
	Collection       = "ZVDD"
)

// DublicCoreRecord is a sketch for highwire XML.
type DublicCoreRecord struct {
	XMLName xml.Name `xml:"record"`
	Header  struct {
		Status     string   `xml:"status,attr"`
		Identifier string   `xml:"identifier"`
		Datestamp  string   `xml:"datestamp"`
		SetSpec    []string `xml:"setSpec"`
	} `xml:"header"`
	Metadata struct {
		DC struct {
			Title       []string `xml:"title"`
			Creator     []string `xml:"creator"`
			Subject     []string `xml:"subject"`
			Publisher   []string `xml:"publisher"`
			Date        []string `xml:"date"`
			Type        []string `xml:"type"`
			Format      []string `xml:"format"`
			Identifier  []string `xml:"identifier"`
			Language    []string `xml:"language"`
			Rights      []string `xml:"rights"`
			Description []string `xml:"description"`
			Source      []string `xml:"source"`
		} `xml:"dc"`
	} `xml:"metadata"`
	About string `xml:"about"`
}

// ToIntermediateSchema sketch.
func (r DublicCoreRecord) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	// oai:www.zvdd.de:urn:nbn:de:gbv:3:1-771990
	urn := strings.Replace(r.Header.Identifier, "oai:www.zvdd.de:", "", 1)
	urnlink := fmt.Sprintf("http://nbn-resolving.de/%s", urn)
	output.URL = append(output.URL, urnlink)

	encodedIdentifier := base64.RawURLEncoding.EncodeToString([]byte(urn))
	output.ID = fmt.Sprintf("ai-%s-%s", SourceIdentifier, encodedIdentifier)
	output.RecordID = urn
	output.SourceID = SourceIdentifier
	output.Genre = Genre
	output.Format = Format
	output.RefType = DefaultRefType
	output.MegaCollections = []string{Collection}

	if len(r.Metadata.DC.Title) > 0 {
		output.ArticleTitle = r.Metadata.DC.Title[0]
	}
	for _, v := range r.Metadata.DC.Creator {
		v = strings.TrimSpace(v)
		v = strings.TrimRight(v, ",")
		output.Authors = append(output.Authors, finc.Author{Name: v})
	}
	output.Abstract = strings.Join(r.Metadata.DC.Source, "\n")

	for _, v := range r.Metadata.DC.Publisher {
		output.Publishers = append(output.Publishers, v)
	}
	for _, v := range r.Metadata.DC.Subject {
		output.Subjects = append(output.Subjects, v)
	}

	// Reach out to METS for dates.
	return output, nil
}
