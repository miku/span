package zvdd

import "encoding/xml"

// Mets fields, relevant for conversion, TODO(miku): complete.
type Mets struct {
	XMLName xml.Name `xml:"mets"`
	Header  struct {
		XMLName xml.Name `xml:"metsHdr"`
		Agents  []struct {
			XMLName   xml.Name `xml:"agent"`
			Role      string   `xml:"attr,ROLE"`
			Type      string   `xml:"attr,TYPE"`
			OtherType string   `xml:"attr,OTHERTYPE"`
			Name      string   `xml:">name"`
		}
	}
}
