package ieee

import (
	"encoding/xml"
	"io"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

type Publication struct {
	xml.Name        `xml:"publication"`
	Title           string `xml:"title"`
	Normtitle       string `xml:"normtitle"`
	Publicationinfo struct {
		Idamsid               string `xml:"idamsid"`
		Publicationtype       string `xml:"publicationtype"`
		Publicationsubtype    string `xml:"publicationsubtype"`
		Ieeeabbrev            string `xml:"ieeeabbrev"`
		Acronym               string `xml:"acronym"`
		Pubstatus             string `xml:"pubstatus"`
		Publicationopenaccess string `xml:"publicationopenaccess"`
		StandardId            string `xml:"standard_id"`
		Packagememberset      struct {
			Packagemember []string `xml:"packagemember"`
		} `xml:"packagememberset"`
		Isbn struct {
			Isbntype  string `xml:"isbntype,attr"`
			Mediatype string `xml:"mediatype,attr"`
		} `xml:"isbn"`
		Issn                string `xml:"issn"`
		Pubtopicalbrowseset string `xml:"pubtopicalbrowseset"`
		Copyrightgroup      string `xml:"copyrightgroup"`
		Publisher           string `xml:"publisher"`
		Holdstatus          string `xml:"holdstatus"`
		Confgroup           struct {
			Confdate []struct {
				Confdatetype string `xml:"confdatetype,attr"`
				Year         string `xml:"year"`
				Month        string `xml:"month"`
				Day          string `xml:"day"`
			} `xml:"confdate"`
			Conflocation  string `xml:"conflocation"`
			Confcountry   string `xml:"confcountry"`
			DoiPermission string `xml:"doi_permission"`
		} `xml:"confgroup"`
		Amsid string `xml:"amsid"`
	} `xml:"publicationinfo"`
	Volume struct {
		Volumeinfo struct {
			Year      string `xml:"year"`
			Idamsid   string `xml:"idamsid"`
			Notegroup string `xml:"notegroup"`
			Issue     struct {
				Amsid       string `xml:"amsid"`
				Issuestatus string `xml:"issuestatus"`
			} `xml:"issue"`
		} `xml:"volumeinfo"`
		Article struct {
			Title       string `xml:"title"`
			Articleinfo struct {
				Articleseqnum          string `xml:"articleseqnum"`
				Csarticlesortorder     string `xml:"csarticlesortorder"`
				Articledoi             string `xml:"articledoi"`
				Idamsid                string `xml:"idamsid"`
				Articlestatus          string `xml:"articlestatus"`
				Articleopenaccess      string `xml:"articleopenaccess"`
				Articleshowflag        string `xml:"articleshowflag"`
				Articleplagiarizedflag string `xml:"articleplagiarizedflag"`
				Articlenodoiflag       string `xml:"articlenodoiflag"`
				Articlecoverimageflag  string `xml:"articlecoverimageflag"`
				Csarticlehtmlflag      string `xml:"csarticlehtmlflag"`
				Articlereferenceflag   string `xml:"articlereferenceflag"`
				Articlepeerreviewflag  string `xml:"articlepeerreviewflag"`
				Holdstatus             string `xml:"holdstatus"`
				Articlecopyright       struct {
					Holderisieee string `xml:"holderisieee,attr"`
					Year         string `xml:"year,attr"`
				} `xml:"articlecopyright"`
				Abstract    string `xml:"abstract"`
				Authorgroup struct {
					Author []struct {
						Normname    string `xml:"normname"`
						Surname     string `xml:"surname"`
						Affiliation string `xml:"affiliation"`
						Firstname   string `xml:"firstname"`
					} `xml:"author"`
				} `xml:"authorgroup"`
				Date struct {
					Datetype string `xml:"datetype,attr"`
					Year     string `xml:"year"`
					Month    string `xml:"month"`
				} `xml:"date"`
				Numpages string `xml:"numpages"`
				Size     string `xml:"size"`
				Filename []struct {
					Docpartition string `xml:"docpartition,attr"`
					Filetype     string `xml:"filetype,attr"`
				} `xml:"filename"`
				Artpagenums struct {
					Endpage   string `xml:"endpage,attr"`
					Startpage string `xml:"startpage,attr"`
				} `xml:"artpagenums"`
				Numreferences string `xml:"numreferences"`
				Amsid         string `xml:"amsid"`
				Keywordset    struct {
					Keywordtype string   `xml:"keywordtype,attr"`
					Keyword     []string `xml:"keyword"`
				} `xml:"keywordset"`
			} `xml:"articleinfo"`
		} `xml:"article"`
	} `xml:"volume"`
}

type IEEE struct{}

// Iterate. Most data is in a single XML file.
func (s IEEE) Iterate(r io.Reader) (<-chan []span.Importer, error) {
	return span.FromXML(r, "publication", func(d *xml.Decoder, se xml.StartElement) (span.Importer, error) {
		p := new(Publication)
		err := d.DecodeElement(&p, &se)
		return p, err
	})
}

// ToIntermediateSchema does a type conversion only.
func (p Publication) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	is := finc.NewIntermediateSchema()
	is.ArticleTitle = p.Title
	return is, nil
}
