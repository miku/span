package genderopen

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
)

// Record was generated 2018-05-11 14:30:28 by tir on sol.
type Record struct {
	XMLName xml.Name `xml:"Record"`
	Text    string   `xml:",chardata"`
	Header  struct {
		Text       string `xml:",chardata"`
		Status     string `xml:"status,attr"`
		Identifier struct {
			Text string `xml:",chardata"` // oai:www.genderopen.de:255...
		} `xml:"identifier"`
		Datestamp struct {
			Text string `xml:",chardata"` // 2017-11-30T13:54:17Z, 201...
		} `xml:"datestamp"`
		SetSpec []struct {
			Text string `xml:",chardata"` // com_13579_1, col_13579_3,...
		} `xml:"setSpec"`
	} `xml:"header"`
	Metadata struct {
		Text string `xml:",chardata"`
		Dc   struct {
			Text           string `xml:",chardata"`
			OaiDc          string `xml:"oai_dc,attr"`
			Doc            string `xml:"doc,attr"`
			Xsi            string `xml:"xsi,attr"`
			Dc             string `xml:"dc,attr"`
			SchemaLocation string `xml:"schemaLocation,attr"`
			Title          struct {
				Text string `xml:",chardata"` // Ausweitung der Geschlecht...
			} `xml:"title"`
			Creator []struct {
				Text string `xml:",chardata"` // Brunner, Claudia, Döllin...
			} `xml:"creator"`
			Contributor []struct {
				Text string `xml:",chardata"` // Lakitsch, Maximilian, Ste...
			} `xml:"contributor"`
			Subject []struct {
				Text string `xml:",chardata"` // Geschlecht, Krieg, Gerech...
			} `xml:"subject"`
			Date struct {
				Text string `xml:",chardata"` // 2015, 2004, 2016, 2007, 2...
			} `xml:"date"`
			Type []struct {
				Text string `xml:",chardata"` // doc-type:bookPart, Text, ...
			} `xml:"type"`
			Identifier []struct {
				Text string `xml:",chardata"` // urn:ISBN:978-3-643-50677-...
			} `xml:"identifier"`
			Language struct {
				Text string `xml:",chardata"` // ger, ger, ger, ger, ger, ...
			} `xml:"language"`
			Rights []struct {
				Text string `xml:",chardata"` // https://creativecommons.o...
			} `xml:"rights"`
			Format struct {
				Text string `xml:",chardata"` // application/pdf, applicat...
			} `xml:"format"`
			Publisher []struct {
				Text string `xml:",chardata"` // LIT, Wien, VSA-Verlag, Ha...
			} `xml:"publisher"`
			Source struct {
				Text string `xml:",chardata"` // Lakitsch, Maximilian; Ste...
			} `xml:"source"`
			Description []struct {
				Text string `xml:",chardata"` // Nachdem kosmetische Genit...
			} `xml:"description"`
		} `xml:"dc"`
	} `xml:"metadata"`
	About struct {
		Text string `xml:",chardata"`
	} `xml:"about"`
}

func parsePages(s string) (start, end, total string) {
	p := regexp.MustCompile(`([1-9][0-9]*)-([1-9][0-9]*)`)
	match := p.FindStringSubmatch(s)
	if len(match) < 3 {
		return "", "", ""
	}
	ss, es := match[1], match[2]
	u, _ := strconv.Atoi(ss)
	v, _ := strconv.Atoi(es)
	return ss, es, fmt.Sprintf("%d", v-u)
}

func (record Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	output.SourceID = "13024"
	encodedRecordID := base64.RawURLEncoding.EncodeToString([]byte(record.Header.Identifier.Text))
	output.RecordID = encodedRecordID
	output.ID = fmt.Sprintf("ai-%s-%s", output.SourceID, output.RecordID)
	output.MegaCollections = append(output.MegaCollections, "Gender Open")
	output.Genre = "article"
	output.RefType = "EJOUR"

	output.ArticleTitle = record.Metadata.Dc.Title.Text
	output.BookTitle = record.Metadata.Dc.Source.Text
	for _, v := range record.Metadata.Dc.Creator {
		output.Authors = append(output.Authors, finc.Author{Name: v.Text})
	}
	for _, v := range record.Metadata.Dc.Identifier {
		if strings.HasPrefix(v.Text, "http") {
			output.URL = append(output.URL, v.Text)
		}
		if strings.HasPrefix(v.Text, "http://dx.doi.org/") {
			output.DOI = strings.Replace(v.Text, "http://dx.doi.org/", "", -1)
		}
	}

	for _, p := range record.Metadata.Dc.Publisher {
		output.Publishers = append(output.Publishers, p.Text)
	}
	if record.Metadata.Dc.Date.Text == "" {
		return output, span.Skip{Reason: "empty date"}
	}
	if len(record.Metadata.Dc.Date.Text) < 4 {
		return output, span.Skip{Reason: "short date"}
	}
	if record.Metadata.Dc.Date.Text != "" {
		s := record.Metadata.Dc.Date.Text[:4]
		date, err := time.Parse("2006", s)
		if err != nil {
			return output, err
		}
		output.Date = date
		output.RawDate = output.Date.Format("2006-01-02")
	}

	for _, s := range record.Metadata.Dc.Subject {
		output.Subjects = append(output.Subjects, s.Text)
	}
	start, end, total := parsePages(record.Metadata.Dc.Source.Text)
	output.StartPage = start
	output.EndPage = end
	output.PageCount = total

	return output, nil
}
