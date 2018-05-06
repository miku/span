package ssoar

// Records was generated 2018-05-06 21:37:04 by tir on hayiti.
import (
	"encoding/xml"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span/formats/finc"
)

type Record struct {
	XMLName xml.Name `xml:"Record"`
	Text    string   `xml:",chardata"`
	Header  struct {
		Text       string `xml:",chardata"`
		Status     string `xml:"status,attr"`
		Identifier struct {
			Text string `xml:",chardata"` // oai:gesis.izsoz.de:docume...
		} `xml:"identifier"`
		Datestamp struct {
			Text string `xml:",chardata"` // 2012-08-29T21:40:31Z, 201...
		} `xml:"datestamp"`
		SetSpec []struct {
			Text string `xml:",chardata"` // com_community_10100, com_...
		} `xml:"setSpec"`
	} `xml:"header"`
	Metadata struct {
		Text   string `xml:",chardata"`
		Record struct {
			Text           string `xml:",chardata"`
			Xmlns          string `xml:"xmlns,attr"`
			Doc            string `xml:"doc,attr"`
			Xalan          string `xml:"xalan,attr"`
			Xsi            string `xml:"xsi,attr"`
			SchemaLocation string `xml:"schemaLocation,attr"`
			Leader         struct {
				Text string `xml:",chardata"` // 00000nam a2200000 u 4500,...
			} `xml:"leader"`
			Controlfield []struct {
				Text string `xml:",chardata"` // 20080514135900.0, cr|||||...
				Tag  string `xml:"tag,attr"`
			} `xml:"controlfield"`
			Datafield []struct {
				Text     string `xml:",chardata"`
				Ind2     string `xml:"ind2,attr"`
				Ind1     string `xml:"ind1,attr"`
				Tag      string `xml:"tag,attr"`
				Subfield []struct {
					Text string `xml:",chardata"` // b, http://www.ssoar.info/...
					Code string `xml:"code,attr"`
				} `xml:"subfield"`
			} `xml:"datafield"`
		} `xml:"record"`
	} `xml:"metadata"`
	About struct {
		Text string `xml:",chardata"`
	} `xml:"about"`
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
	for _, f := range r.Metadata.Record.Datafield {
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

func (r Record) Title() string {
	result := r.MustGetFirstDataField("245.a")
	subtitle := strings.TrimSpace(r.MustGetFirstDataField("245.b"))
	if subtitle != "" {
		result = result + ": " + subtitle
	}
	return result
}

func (r Record) JournalTitle() string {
	// In: Journal of Social Work Practice ; 19 (2005) 1 ; 87-101
	// In: Balzer, Wolfgang (Hg.), Pearce, David A. (Hg.), Schmidt, Heinz-Jürgen (Hg.): Reduction in science : structure, examples, philosophical problems. 1984. S. 331-357. ISBN 90-277-1811-3
	for _, s := range r.MustGetDataFields("500.a") {
		if !strings.HasPrefix(s, "In:") {
			continue
		}
		parts := strings.Split(s, ";")
		switch len(parts) {
		case 1:
			pp := strings.Split(s[4:], ":")
			switch len(pp) {
			case 1:
				ppp := strings.Split(s[4:], ".")
				return ppp[0]
			case 2:
				ppp := strings.Split(pp[1], ".")
				return ppp[0]
			case 3:
				a := pp[1]
				b := strings.Split(pp[2], ".")
				return fmt.Sprintf("%s: %s", a, b)
			default:
				return s
			}
		case 2:
			return strings.TrimSpace(parts[0][4:])
		case 3:
			return strings.TrimSpace(parts[0][4:])
		case 4:
			return strings.TrimSpace(parts[0][4:])
		default:
			log.Printf("unparsed parent title: %s", s)
		}
	}
	return ""
}

func (r Record) ID() (string, error) {
	parts := strings.Split(r.Header.Identifier.Text, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected identifier: %s", r.Header.Identifier.Text)
	} else {
		return strings.TrimSpace(parts[1]), nil
	}
	return "", fmt.Errorf("no identifier found: %s", r.Header.Identifier.Text)
}

func (r Record) FindYear() string {
	date := strings.TrimSpace(r.MustGetFirstDataField("264.c"))
	if len(date) == 4 {
		// XXX: o.J.
		if date == "o.J." {
			return fmt.Sprintf("2010-01-01")
		}
		return fmt.Sprintf("%s-01-01", date)
	}
	yp := regexp.MustCompile(`[12][6789012]\d\d`)
	for _, f := range r.MustGetDataFields("500.a") {
		matches := yp.FindAllString(f, -1)
		if len(matches) > 0 {
			return fmt.Sprintf("%s-01-01", matches[0])
		}
	}
	return "1970-01-01"
}

func (r Record) FindFormat() string {
	leader := string(r.Metadata.Record.Leader.Text)
	llen := len(string(leader))
	switch {
	case llen > 7 && string(leader[7]) == "m":
		return "eBook"
	case llen > 19 && (string(leader[19]) == " " || string(leader[19]) == "b"):
		return "eBook"
	default:
		return "ElectronicArticle"
	}
}

// stringDifference returns the numeric value of a-b of strings a and b as string.
func stringDifference(a, b string) string {
	i, err := strconv.Atoi(a)
	if err != nil {
		return ""
	}
	j, err := strconv.Atoi(b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", i-j)
}

func (r Record) FindPages() (string, string, string) {
	v := r.MustGetFirstDataField("300.a")
	p := regexp.MustCompile(`([1-9][0-9]*)`)
	matches := p.FindAllString(v, -1)
	switch {
	case len(matches) == 1:
		return matches[0], matches[0], "1"
	case len(matches) == 2:
		return matches[0], matches[1], stringDifference(matches[1], matches[0])
	default:
		// Try to find pages in 500.a
		q := regexp.MustCompile(`([1-9][0-9]{0,3})-([1-9][0-9]{0,3})`)
		for _, w := range r.MustGetDataFields("500.a") {
			sms := q.FindAllStringSubmatch(w, -1)
			if sms == nil {
				return "", "", ""
			}
			if len(sms[0]) >= 3 {
				pagecount := stringDifference(sms[0][2], sms[0][1])
				return sms[0][0], sms[0][1], pagecount
			}
		}
	}
	return "", "", ""
}

// ToIntermediateSchema converts a MarcXML-ish record.
func (r Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()
	id, err := r.ID()
	if err != nil {
		return output, err
	}
	output.RecordID = id
	output.SourceID = "30"
	output.ID = fmt.Sprintf("ai-%s-%s", output.SourceID, output.RecordID)
	output.Format = r.FindFormat()
	output.MegaCollections = []string{"SSOAR Social Science Open Access Repository"}

	switch output.Format {
	case "eBook":
		output.BookTitle = r.Title()
		output.Genre = "book"
	default:
		output.ArticleTitle = r.Title()
		output.Genre = "article"
	}
	output.JournalTitle = r.JournalTitle()
	output.URL = r.MustGetDataFields("856.u")
	for _, author := range r.MustGetDataFields("100.a") {
		output.Authors = append(output.Authors, finc.Author{Name: author})
	}
	for _, author := range r.MustGetDataFields("700.a") {
		output.Authors = append(output.Authors, finc.Author{Name: author})
	}
	output.Abstract = r.MustGetFirstDataField("520.a")
	output.Subjects = r.MustGetDataFields("650.a")

	output.RawDate = r.FindYear()
	date, err := time.Parse("2006-01-02", output.RawDate)
	if err != nil {
		log.Fatal(err)
	}
	output.Date = date
	output.Languages = r.MustGetDataFields("041.a")
	output.StartPage, output.EndPage, output.PageCount = r.FindPages()
	output.Series = r.MustGetFirstDataField("490.a")
	output.ISBN = r.MustGetDataFields("020.a")

	return output, nil
}
