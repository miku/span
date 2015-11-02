package genios

type Document struct {
	ID               string   `xml:"ID,attr"`
	DB               string   `xml:"DB,attr"`
	IDNAME           string   `xml:"IDNAME,attr"`
	ISSN             string   `xml:"ISSN"`
	Source           string   `xml:"Source"`
	PublicationTitle string   `xml:"Publication-Title"`
	Title            string   `xml:"Title"`
	Year             string   `xml:"Year"`
	RawDate          string   `xml:"Date"`
	Volume           string   `xml:"Volume"`
	Issue            string   `xml:"Issue"`
	RawAuthors       []string `xml:"Authors>Author"`
	Language         string   `xml:"Language"`
	Abstract         string   `xml:"Abstract"`
	Group            string   `xml:"x-group"`
	Descriptors      string   `xml:"Descriptors>Descriptor"`
	Text             string   `xml:"Text"`
}
