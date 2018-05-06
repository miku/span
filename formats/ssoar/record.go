package ssoar

// Records was generated 2018-05-06 21:37:04 by tir on hayiti.
import (
	"encoding/xml"

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

// ToIntermediateSchema
func (r Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	return nil, nil
}
