package olms

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/miku/span"

	"github.com/miku/span/formats/finc"
)

// Record was generated 2018-03-01 19:44:04 by tir on hayiti.
type Record struct {
	XMLName xml.Name `xml:"record"`
	Text    string   `xml:",chardata"`
	Header  struct {
		Text       string `xml:",chardata"`
		Identifier struct {
			Text string `xml:",chardata"` // www.olmsonline.de:PPN5212...
		} `xml:"identifier"`
		Datestamp struct {
			Text string `xml:",chardata"` // 2012-02-01T12:29:27Z, 201...
		} `xml:"datestamp"`
		SetSpec struct {
			Text string `xml:",chardata"` // deutsche_literaturklassik...
		} `xml:"setSpec"`
	} `xml:"header"`
	Metadata struct {
		Text string `xml:",chardata"`
		Dc   struct {
			Text           string `xml:",chardata"`
			OaiDc          string `xml:"oai_dc,attr"`
			Dc             string `xml:"dc,attr"`
			Xsi            string `xml:"xsi,attr"`
			SchemaLocation string `xml:"schemaLocation,attr"`
			Title          struct {
				Text string `xml:",chardata"` // Ausgewählte Dramen und E...
			} `xml:"title"`
			Creator []struct {
				Text string `xml:",chardata"` // Fouqué, Friedrich, Campe...
			} `xml:"creator"`
			Subject struct {
				Text string `xml:",chardata"` // Deutsche_Literaturklassik...
			} `xml:"subject"`
			Publisher struct {
				Text string `xml:",chardata"` // Olms, Olms, Olms, Olms, O...
			} `xml:"publisher"`
			Date struct {
				Text string `xml:",chardata"` // 1994, 1969, 1987, 2003, 1...
			} `xml:"date"`
			Type []struct {
				Text string `xml:",chardata"` // Text, Monograph, Text, Vo...
			} `xml:"type"`
			Format []struct {
				Text string `xml:",chardata"` // image/jpeg, application/p...
			} `xml:"format"`
			Identifier []struct {
				Text string `xml:",chardata"` // http://www.olmsonline.de/...
			} `xml:"identifier"`
			Source struct {
				Text string `xml:",chardata"` // Fouqué, Friedrich: Ausge...
			} `xml:"source"`
			Relation struct {
				Text string `xml:",chardata"` // Campe, Joachim Heinrich: ...
			} `xml:"relation"`
		} `xml:"dc"`
	} `xml:"metadata"`
}

// ToIntermediateSchema converts a single record.
func (record Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	output.SourceID = "12502"
	parts := strings.Split(record.Header.Identifier.Text, ":")
	if len(parts) != 2 {
		return output, fmt.Errorf("cannot find identifier: %s", record.Header.Identifier.Text)
	}
	output.RecordID = record.Header.Identifier.Text
	output.ID = fmt.Sprintf("ai-%s-%s", output.SourceID, parts[1])
	output.MegaCollections = append(output.MegaCollections, "Olms")
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
	}

	output.Publishers = append(output.Publishers, record.Metadata.Dc.Publisher.Text)
	if record.Metadata.Dc.Date.Text == "" {
		return output, span.Skip{Reason: "empty date"}
	}
	if len(record.Metadata.Dc.Date.Text) < 4 {
		return output, span.Skip{Reason: "short date"}
	}
	if record.Metadata.Dc.Date.Text != "" {
		// <dc:date>19787</dc:date> --
		s := record.Metadata.Dc.Date.Text[:4]
		date, err := time.Parse("2006", s)
		if err != nil {
			return output, err
		}
		output.Date = date
		output.RawDate = output.Date.Format("2006-01-02")
	}

	if record.Metadata.Dc.Subject.Text != "" {
		output.Subjects = append(output.Subjects, record.Metadata.Dc.Subject.Text)
	}

	return output, nil
}
