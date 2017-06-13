package oai

import "encoding/xml"

// DublinCollection record.
type DublinCollection struct {
	xml.Name `xml:"Record"`
	Header   struct {
		Status     string `xml:"status,attr"`
		Identifier string `xml:"identifier"`
		DateStamp  string `xml:"datestamp"`
		SetSpec    string `xml:"setSpec"`
	} `xml:"header"`
	Metadata struct {
		Collection struct {
			DC struct {
				Title      []string `xml:"title"`
				Creator    []string `xml:"creator"`
				Publisher  []string `xml:"publisher"`
				Date       []string `xml:"date"`
				Language   []string `xml:"language"`
				Format     []string `xml:"format"`
				Subject    []string `xml:"subject"`
				Relation   []string `xml:"relation"`
				Identifier []string `xml:"identifier"`
				Rights     []string `xml:"rights"`
				Source     []string `xml:"source"`
			} `xml:"dc"`
		} `xml:"dcCollection"`
	} `xml:"metadata"`
}
