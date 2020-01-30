package mediarep

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/miku/span/formats/finc"
)

// Dim was generated 2018-10-02 14:55:51 by tir on sol.
type Dim struct {
	XMLName xml.Name `xml:"Record"`
	Text    string   `xml:",chardata"`
	Header  struct {
		Text       string   `xml:",chardata"`
		Status     string   `xml:"status,attr"`
		Identifier string   `xml:"identifier"` // oai:localhost:doc/2019, o...
		Datestamp  string   `xml:"datestamp"`  // 2018-09-24T22:09:15Z, 201...
		SetSpec    []string `xml:"setSpec"`    // com_doc_344, col_doc_1951...
	} `xml:"header"`
	Metadata struct {
		Text string `xml:",chardata"`
		Dim  struct {
			Text           string `xml:",chardata"`
			Dim            string `xml:"dim,attr"`
			Doc            string `xml:"doc,attr"`
			Xsi            string `xml:"xsi,attr"`
			SchemaLocation string `xml:"schemaLocation,attr"`
			Field          []struct {
				Text      string `xml:",chardata"` // Geser, Hans, 2018-09-24T1...
				Mdschema  string `xml:"mdschema,attr"`
				Element   string `xml:"element,attr"`
				Qualifier string `xml:"qualifier,attr"`
				Lang      string `xml:"lang,attr"`
			} `xml:"field"`
		} `xml:"dim"`
	} `xml:"metadata"`
	About string `xml:"about"`
}

// FieldValue returns the first value given a mdschema and element attribute
// name, empty string if nothing is found.
func (r *Dim) FieldValue(schema, element, qualifier string) string {
	for _, f := range r.Metadata.Dim.Field {
		if f.Mdschema == schema && f.Element == element && f.Qualifier == qualifier {
			return f.Text
		}
	}
	return ""
}

// FieldValues returns the all values given a mdschema and element attribute
// name, empty slice if nothing is found.
func (r *Dim) FieldValues(schema, element, qualifier string) (result []string) {
	for _, f := range r.Metadata.Dim.Field {
		if f.Mdschema == schema && f.Element == element && f.Qualifier == qualifier {
			result = append(result, f.Text)
		}
	}
	return result
}

// PageCount returns number of pages as string or the the empty string.
func (r *Dim) PageCount() string {
	spage := r.FieldValue("local", "source", "spage")
	epage := r.FieldValue("local", "source", "epage")
	s, err := strconv.Atoi(spage)
	if err != nil {
		return ""
	}
	e, err := strconv.Atoi(epage)
	if err != nil {
		return ""
	}
	if e < s {
		return ""
	}
	return fmt.Sprintf("%d", e-s)
}

// ToIntermediateSchema converts mediarep/dim to intermediate schema
func (r *Dim) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	output.SourceID = "170"
	output.MegaCollections = []string{"sid-170-col-mediarep"}

	output.ArticleTitle = r.FieldValue("dc", "title", "")
	output.Subjects = r.FieldValues("dc", "subject", "")
	output.Publishers = r.FieldValues("dc", "publisher", "")
	output.Languages = r.FieldValues("dc", "language", "")
	output.ISBN = r.FieldValues("dc", "identifier", "isbn")
	output.ISSN = r.FieldValues("dc", "identifier", "issn")
	output.URL = r.FieldValues("dc", "identifier", "uri")
	output.Issue = r.FieldValue("local", "source", "issue")
	output.Volume = r.FieldValue("local", "source", "volume")
	output.RawDate = r.FieldValue("dc", "date", "issued")
	output.StartPage = r.FieldValue("local", "source", "spage")
	output.EndPage = r.FieldValue("local", "source", "epage")
	output.PageCount = r.PageCount()

	var date time.Time
	var err error

	layouts := []string{"2006", "2006-01", "2006-01-02"}
	for _, l := range layouts {
		date, err = time.Parse(l, output.RawDate)
		if err != nil {
			continue
		}
		break
	}

	if date.IsZero() {
		return nil, err
	}

	output.Date = date

	for _, c := range r.FieldValues("dc", "creator", "") {
		output.Authors = append(output.Authors, finc.Author{Name: c})
	}
	for _, c := range r.FieldValues("dc", "contributor", "") {
		output.Authors = append(output.Authors, finc.Author{Name: c})
	}
	return output, nil
}
