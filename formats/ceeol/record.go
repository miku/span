package ceeol

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
)

// identifierRegexp, since ID is buried in 856.u link.
var identifierRegexp = regexp.MustCompile("id=([0-9]*)")

// Record for MARC-XML data, Ceeol style.
type Record struct {
	XMLName        xml.Name `xml:"record"`
	Text           string   `xml:",chardata"`
	Xmlns          string   `xml:"xmlns,attr"`
	Doc            string   `xml:"doc,attr"`
	Xalan          string   `xml:"xalan,attr"`
	Xsi            string   `xml:"xsi,attr"`
	SchemaLocation string   `xml:"schemaLocation,attr"`
	Leader         struct {
		Text string `xml:",chardata"`
	} `xml:"leader"`
	Controlfield []struct {
		Text string `xml:",chardata"`
		Tag  string `xml:"tag,attr"`
	} `xml:"controlfield"`
	Datafield []struct {
		Text     string `xml:",chardata"`
		Ind2     string `xml:"ind2,attr"`
		Ind1     string `xml:"ind1,attr"`
		Tag      string `xml:"tag,attr"`
		Subfield []struct {
			Text string `xml:",chardata"`
			Code string `xml:"code,attr"`
		} `xml:"subfield"`
	} `xml:"datafield"`
}

func (r Record) MustGetControlField(tag string) string {
	v, err := r.GetControlField(tag)
	if err != nil {
		panic(err)
	}
	return v
}

func (r Record) GetControlField(tag string) (string, error) {
	for _, f := range r.Controlfield {
		if f.Tag == tag {
			return f.Text, nil
		}
	}
	return "", fmt.Errorf("undefined tag: %s", tag)
}

func (r Record) MustGetFirstDataField(spec string) string {
	value, err := r.GetFirstDataField(spec)
	if err != nil {
		panic(err)
	}
	return value
}

func (r Record) GetFirstDataField(spec string) (string, error) {
	values, err := r.GetDataFields(spec)
	if err != nil {
		return "", err
	}
	if len(values) == 0 {
		return "", nil
	}
	return values[0], nil
}

func (r Record) MustGetDataFields(spec string) []string {
	result, err := r.GetDataFields(spec)
	if err != nil {
		panic(err)
	}
	return result
}

func (r Record) GetDataFields(spec string) (result []string, err error) {
	parts := strings.Split(spec, ".")
	if len(parts) != 2 {
		return result, fmt.Errorf("spec must be of the form tag.subfield, like 245.a")
	}
	tag, subfield := parts[0], parts[1]
	for _, f := range r.Datafield {
		if f.Tag == tag {
			for _, sf := range f.Subfield {
				if sf.Code == subfield {
					result = append(result, sf.Text)
				}
			}
		}
	}
	return result, nil
}

// ID returns the identifier, the identifier might be hidden in the URL (e.g.
// <marc:subfield
// code="u">https://www.ceeol.com/search/book-detail?id=279462</marc:subfield>),
// was <UniqueID>279462</UniqueID>.
func (r Record) ID() (string, error) {
	values, err := r.GetDataFields("856.u")
	if err != nil {
		return "", err
	}
	for _, v := range values {
		matches := identifierRegexp.FindStringSubmatch(v)
		if len(matches) < 2 {
			return "", fmt.Errorf("value without id: %v", v)
		}
		id := matches[1]
		return id, nil
	}
	return "", fmt.Errorf("missing identifier")
}

// Title returns some title.
func (r Record) Title() (string, error) {
	title, err := r.GetFirstDataField("245.a")
	if err != nil {
		return "", err
	}
	subtitle, err := r.GetFirstDataField("245.b")
	if err != nil {
		return "", err
	}
	title = strings.TrimSpace(title)
	subtitle = strings.TrimSpace(subtitle)
	if subtitle == "" {
		return title, nil
	}
	return fmt.Sprintf("%s: %s", title, subtitle), nil
}

// RecordFormat returns "book", "article" or "unknown".
func (r Record) RecordFormat() string {
	v, _ := r.GetDataFields("022.a")
	if len(v) > 0 {
		return "article"
	}
	v, _ = r.GetDataFields("020.a")
	if len(v) > 0 {
		return "book"
	}
	return "unknown"
}

// ISSN returns a list of ISSN.
func (r Record) ISSN() (values []string) {
	values, _ = r.GetDataFields("022.a")
	return values
}

// ISBN returns a list of ISBN.
func (r Record) ISBN() (values []string) {
	values, _ = r.GetDataFields("020.a")
	return values
}

func (r Record) Abstract() string {
	values, _ := r.GetDataFields("520.a")
	return strings.Join(values, "\n")
}

func (r Record) Authors() (authors []finc.Author) {
	values, _ := r.GetDataFields("100.a")
	for _, value := range values {
		authors = append(authors, finc.Author{Name: value})
	}
	values, _ = r.GetDataFields("700.a")
	for _, value := range values {
		authors = append(authors, finc.Author{Name: value})
	}
	return authors
}

func (r Record) SubjectHeadings() []string {
	values, _ := r.GetDataFields("650.a")
	return values
}

// ToIntermediateSchema converts CEEOL marcxml data into intermediate schema.
func (r Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()
	v, err := r.ID()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	output.RecordID = v

	v, err = r.Title()
	if err != nil {
		return output, span.Skip{
			Reason: fmt.Sprintf("missing title: %s", output.RecordID)}
	}

	switch r.RecordFormat() {
	case "book":
		output.BookTitle = v
	default:
		output.ArticleTitle = v
	}

	output.Format = Format
	output.SourceID = SourceIdentifier
	output.Genre = Genre
	output.RefType = DefaultRefType
	output.MegaCollections = []string{Collection}
	output.ISSN = r.ISSN()
	output.ISBN = r.ISBN()
	output.Abstract = r.Abstract()
	output.Authors = r.Authors()
	output.Subjects = r.SubjectHeadings()

	return output, nil
}
