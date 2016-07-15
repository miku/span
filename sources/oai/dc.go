package oai

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

var (
	// datePattern matches date portions syntactically
	datePattern = regexp.MustCompile(`[12][0-9][0-9][0-9]-[01][0-9]-[0123][0-9]`)
	// http://stackoverflow.com/questions/27910/finding-a-doi-in-a-document-or-page
	doiPattern = regexp.MustCompile(`10[.][0-9]{4,}.+`)
)

type Record struct {
	XMLName xml.Name `xml:"Record"`
	Header  struct {
		Status     string `xml:"status,attr"`
		Identifier string `xml:"identifier"`
		Datestamp  string `xml:"datestamp"`
		SetSpec    string `xml:"setSpec"`
	} `xml:"header"`
	About string `xml:"about"`

	Metadata struct {
		// http://www.openarchives.org/OAI/2.0/oai_dc.xsd
		Dc struct {
			Title       []string `xml:"title"`
			Creator     []string `xml:"creator"`
			Subject     []string `xml:"subject"`
			Description []string `xml:"description"`
			Publisher   []string `xml:"publisher"`
			Contributor []string `xml:"contributor"`
			Date        []string `xml:"date"`
			Type        []string `xml:"type"`
			Identifier  []string `xml:"identifier"`
			Language    []string `xml:"language"`
			Rights      []string `xml:"rights"`
			Format      []string `xml:"format"`
			Source      []string `xml:"source"`
			Relation    []string `xml:"relation"`
			Coverage    []string `xml:"coverage"`
		} `xml:"dc"`
	} `xml:"metadata"`
}

type DublinCore struct{}

// Iterate emits Converter elements via XML decoding.
func (s DublinCore) Iterate(r io.Reader) (<-chan []span.Importer, error) {
	return span.FromXML(r, "Record", func(d *xml.Decoder, se xml.StartElement) (span.Importer, error) {
		doc := new(Record)
		err := d.DecodeElement(&doc, &se)
		return doc, err
	})
}

func (r *Record) Date() (time.Time, error) {
	if len(r.Metadata.Dc.Date) == 0 {
		return time.Time{}, fmt.Errorf("record has no date")
	}
	date := datePattern.FindString(r.Metadata.Dc.Date[0])
	switch {
	case len(r.Metadata.Dc.Date[0]) == 4:
		return time.Parse("2006", r.Metadata.Dc.Date[0])
	case date != "":
		return time.Parse("2006-01-02", date)
	default:
		return time.Time{}, fmt.Errorf("unknown date format: %s", r.Metadata.Dc.Date[0])
	}
}

func (r *Record) Description() string {
	return strings.Join(r.Metadata.Dc.Description, "\n")
}

func (r *Record) Title() string {
	if len(r.Metadata.Dc.Title) > 0 {
		return r.Metadata.Dc.Title[0]
	}
	return ""
}

func (r *Record) Authors() []finc.Author {
	var authors []finc.Author
	for _, c := range r.Metadata.Dc.Creator {
		authors = append(authors, finc.Author{Name: c})
	}
	for _, c := range r.Metadata.Dc.Contributor {
		authors = append(authors, finc.Author{Name: c})
	}
	return authors
}

func (r *Record) DOI() string {
	for _, id := range r.Metadata.Dc.Identifier {
		if doi := doiPattern.FindString(id); doi != "" {
			return doi
		}
	}
	return ""
}

func (r *Record) Links() []string {
	var links []string
	for _, id := range r.Metadata.Dc.Identifier {
		if strings.HasPrefix(id, "http") {
			links = append(links, id)
		}
	}
	return links
}

// ToInternalSchema converts a jats article into an internal schema.
// This is a basic implementation, different source might implement their own.
func (r *Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	date, err := r.Date()
	if err != nil {
		return output, nil
	}

	output.Date = date
	output.RawDate = output.Date.Format("2006-01-02")
	output.Abstract = strings.TrimSpace(r.Description())
	output.ArticleTitle = r.Title()
	output.Publishers = r.Metadata.Dc.Publisher
	output.Authors = r.Authors()
	output.URL = r.Links()
	output.Subjects = r.Metadata.Dc.Subject
	// TODO(miku): normalize
	output.Languages = r.Metadata.Dc.Language
	output.DOI = r.DOI()

	return output, nil
}
