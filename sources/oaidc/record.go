package oaidc

import "encoding/xml"

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
