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

// datePattern matches date portions syntactically
var datePattern = regexp.MustCompile(`[12][0-9][0-9][0-9]-[01][0-9]-[0123][0-9]`)

type Record struct {
	xml.Name `xml:"Record"`
	Header   struct {
		Status     string `xml:"status,attr"`
		Identifier string `xml:"identifier"`
		Datestamp  string `xml:"datestamp"`
		SetSpec    string `xml:"setSpec"`
	} `xml:"header"`
	About string `xml:"about"`

	Dc struct {
		Title       string   `xml:"title"`
		Creator     []string `xml:"creator"`
		Subject     string   `xml:"subject"`
		Description []string `xml:"description"`
		Date        string   `xml:"date"`
		Type        string   `xml:"type"`
		Identifier  string   `xml:"identifier"`
	} `xml:"dc"`
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
	switch {
	case len(r.Dc.Date) == 4:
		return time.Parse("2006", r.Dc.Date)
	case datePattern.FindString(r.Dc.Date) != "":
		return time.Parse("2006-01-02", r.Dc.Date)
	default:
		return time.Time{}, fmt.Errorf("unknown date format: %s", r.Dc.Date)
	}
}

func (r *Record) Description() string {
	return strings.Join(r.Dc.Description, "\n")
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
	output.Abstract = r.Description()

	// output.Abstract = string(article.Front.Article.Abstract.Value)
	// output.ArticleTitle = article.CombinedTitle()
	// output.Authors = article.Authors()
	// output.Fulltext = article.Body.Section.Value
	// output.Genre = "article"
	// output.RefType = "EJOUR"
	// output.Headings = article.Headings()
	// output.ISSN = article.ISSN()
	// output.Issue = article.Front.Article.Issue.Value
	// output.JournalTitle = article.JournalTitle()
	// output.Languages = article.Languages()
	// output.Publishers = append(output.Publishers, article.Front.Journal.Publisher.Name.Value)
	// output.Subjects = article.Subjects()
	// output.Volume = article.Front.Article.Volume.Value

	// output.StartPage = article.Front.Article.FirstPage.Value
	// output.EndPage = article.Front.Article.LastPage.Value
	// output.PageCount = article.PageCount()
	// output.Pages = fmt.Sprintf("%s-%s", output.StartPage, output.EndPage)

	return output, nil
}
