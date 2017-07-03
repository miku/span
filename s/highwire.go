package s

import (
	"encoding/xml"

	"github.com/miku/span/finc"
)

// <Record><header status=""><identifier>oai:open-archive.highwire.org:hmg:2/1/89a</identifier><datestamp>1993-01-01</datestamp><setSpec>HighWire</setSpec><setSpec>OUP</setSpec><setSpec>hmg:2:1</setSpec></header><metadata>
// <oai_dc:dc xmlns:oai_dc="http://www.openarchives.org/OAI/2.0/oai_dc/"
//            xmlns:dc="http://purl.org/dc/elements/1.1/"
//            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
//            xsi:schemaLocation="http://www.openarchives.org/OAI/2.0/oai_dc/ http://www.openarchives.org/OAI/2.0/oai_dc.xsd">
// <dc:title>A polymorphic dinucleotide repeat probe on chromosome 3p: LIB 44-36ca (D3S769)</dc:title>
// <dc:creator>Schmidt, L</dc:creator>
// <dc:creator>Li, H</dc:creator>
// <dc:creator>Wei, MH</dc:creator>
// <dc:creator>Lerman, MI</dc:creator>
// <dc:creator>Zbar, B</dc:creator>
// <dc:creator>Tory, K</dc:creator>
// <dc:subject>ARTICLES</dc:subject>
// <dc:publisher>Oxford University Press</dc:publisher>
// <dc:date>1993-01-01 00:00:00.0</dc:date>
// <dc:type>TEXT</dc:type>
// <dc:format>text/html</dc:format>
// <dc:identifier>http://hmg.oxfordjournals.org/cgi/content/short/2/1/89a</dc:identifier>
// <dc:language>en</dc:language>
// <dc:rights>Copyright (C) 1993, Oxford University Press</dc:rights>
// </oai_dc:dc>
// </metadata><about></about></Record>

// Record is a sketch for highwire XML.
type Record struct {
	XMLName xml.Name `xml:"Record"`
	Header  struct {
		Status     string   `xml:"status,attr"`
		Identifier string   `xml:"identifier"`
		Datestamp  string   `xml:"datestamp"`
		SetSpec    []string `xml:"setSpec"`
	} `xml:"header"`
	Metadata struct {
		DC struct {
			Title     []string `xml:"title"`
			Creator   []string `xml:"creator"`
			Subject   []string `xml:"subject"`
			Publisher []string `xml:"publisher"`
			Date      []string `xml:"date"`
			Type      []string `xml:"type"`
			Format    []string `xml:"format"`
			ID        []string `xml:"identifier"`
			Language  []string `xml:"language"`
			Rights    []string `xml:"rights"`
		} `xml:"dc"`
	} `xml:"metadata"`
	About string `xml:"about"`
}

// ToIntermediateSchema sketch.
func (r Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()
	if len(r.Metadata.DC.Title) > 0 {
		output.ArticleTitle = r.Metadata.DC.Title[0]
	}
	return output, nil
}
