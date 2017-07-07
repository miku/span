package highwire

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

// SourceIdentifier for internal bookkeeping.
const (
	SourceIdentifier = "200"
	Format           = "ElectronicArticle"
	Genre            = "article"
	DefaultRefType   = "EJOUR"
)

// Record is a sketch for highwire XML.
type Record struct {
	XMLName xml.Name `xml:"Record"`
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
		} `xml:"dc"`
	} `xml:"metadata"`
	About string `xml:"about"`
}

// ToIntermediateSchema sketch.
func (r Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	encodedIdentifier := base64.RawURLEncoding.EncodeToString([]byte(r.Header.Identifier))
	output.RecordID = fmt.Sprintf("ai-%s-%s", SourceIdentifier, encodedIdentifier)
	output.SourceID = SourceIdentifier
	output.Genre = Genre
	output.Format = Format
	output.RefType = DefaultRefType

	if len(r.Metadata.DC.Publisher) > 0 {
		output.MegaCollection = fmt.Sprintf("%s (HighWire)", r.Metadata.DC.Publisher[0])
	}

	if len(r.Metadata.DC.Title) > 0 {
		output.ArticleTitle = r.Metadata.DC.Title[0]
	}
	for _, v := range r.Metadata.DC.Creator {
		v = strings.TrimSpace(v)
		v = strings.TrimRight(v, ",")
		output.Authors = append(output.Authors, finc.Author{Name: v})
	}
	for _, v := range r.Metadata.DC.Identifier {
		if strings.HasPrefix(v, "http") {
			output.URL = append(output.URL, v)
		}
		if strings.HasPrefix(v, "http://dx.doi.org/") {
			output.DOI = strings.Replace(v, "http://dx.doi.org/", "", -1)
		}
	}
	output.Abstract = strings.Join(r.Metadata.DC.Description, "\n")

	for _, v := range r.Metadata.DC.Publisher {
		output.Publishers = append(output.Publishers, v)
	}
	for _, v := range r.Metadata.DC.Language {
		if isocode := span.LanguageIdentifier(v); isocode != "" {
			output.Languages = append(output.Languages, isocode)
		}
	}
	for _, v := range r.Metadata.DC.Subject {
		output.Subjects = append(output.Subjects, v)
	}
	if len(r.Metadata.DC.Date) > 0 && len(r.Metadata.DC.Date[0]) >= 10 {
		// 1993-01-01 00:00:00.0
		t, err := time.Parse("2006-01-02", r.Metadata.DC.Date[0][:10])
		if err != nil {
			return output, err
		}
		output.Date = t
		output.RawDate = t.Format("2006-01-02")
	} else {
		return output, fmt.Errorf("could not parse date: %v", r.Metadata.DC.Date)
	}

	return output, nil
}
