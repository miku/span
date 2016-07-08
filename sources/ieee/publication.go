package ieee

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

const (
	SourceID       = "89"
	Format         = "ElectronicArticle"
	Collection     = "IEEE"
	Genre          = "article"
	DefaultRefType = "EJOUR"
)

var (
	ErrNoDate       = errors.New("no date found")
	ErrNoIdentifier = errors.New("missing identifier")
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
		Issn []struct {
			Mediatype string `xml:"mediatype,attr"`
			Value     string `xml:",chardata"`
		} `xml:"issn"`
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
			Volumenum string `xml:"volumenum"`
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
				Issuenum               string `xml:"issuenum"`
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
					Day      string `xml:"day"`
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
					Keywordtype string `xml:"keywordtype,attr"`
					Keyword     []struct {
						Term string `xml:"keywordterm"`
					} `xml:"keyword"`
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

func (p Publication) PaperISSN() (issns []string) {
	for _, issn := range p.Publicationinfo.Issn {
		if strings.ToLower(issn.Mediatype) == "paper" {
			issns = append(issns, issn.Value)
		}
	}
	return
}

func (p Publication) OnlineISSN() (issns []string) {
	for _, issn := range p.Publicationinfo.Issn {
		if strings.ToLower(issn.Mediatype) == "online" {
			issns = append(issns, issn.Value)
		}
	}
	return
}

func (p Publication) Date() (time.Time, error) {
	date := p.Volume.Article.Articleinfo.Date
	y, m, d := "1970", "Jan", "1"
	if date.Year == "" {
		return time.Time{}, ErrNoDate
	} else {
		if _, err := strconv.Atoi(date.Year); err != nil {
			return time.Time{}, err
		}
		y = date.Year
	}
	if date.Month != "" {
		if v, err := strconv.Atoi(date.Month); err != nil {
			m = date.Month
		} else {
			if v > 0 && v < 13 {
				m = date.Month
			} else {
				m = "Jan"
			}
		}
	}
	if date.Day != "" {
		if v, err := strconv.Atoi(date.Day); err != nil {
			if v > 0 && v < 32 {
				m = date.Day
			} else {
				m = "1"
			}
		}
	}
	return time.Parse("2006-Jan-02", fmt.Sprintf("%s-%s-%02s", y, m, d))
}

func (p Publication) Authors() []finc.Author {
	var authors []finc.Author
	for _, author := range p.Volume.Article.Articleinfo.Authorgroup.Author {
		authors = append(authors, finc.Author{FirstName: author.Firstname, LastName: author.Surname})
	}
	return authors
}

// ToIntermediateSchema does a type conversion only.
func (p Publication) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	is := finc.NewIntermediateSchema()
	is.JournalTitle = p.Title
	is.ArticleTitle = p.Volume.Article.Title

	is.ISSN = p.PaperISSN()
	is.EISSN = p.OnlineISSN()

	if len(is.ISSN) == 0 && len(is.EISSN) == 0 {
		// TODO(miku): sort out the various types
		return is, span.Skip{Reason: "no ISSN"}
	}

	is.Abstract = p.Volume.Article.Articleinfo.Abstract

	date, err := p.Date()
	if err != nil {
		return is, span.Skip{Reason: err.Error()}
	}
	is.Date = date
	is.RawDate = date.Format("2006-01-02")

	is.Authors = p.Authors()

	is.URL = []string{}

	is.SourceID = SourceID
	is.MegaCollection = Collection
	is.Format = Format

	if p.Volume.Article.Articleinfo.Amsid != "" {
		is.URL = append(is.URL, fmt.Sprintf("http://ieeexplore.ieee.org/stamp/stamp.jsp?arnumber=%s", p.Volume.Article.Articleinfo.Amsid))
		is.RecordID = fmt.Sprintf("ai-89-%s", base64.RawURLEncoding.EncodeToString([]byte(p.Volume.Article.Articleinfo.Amsid)))
	} else {
		return is, ErrNoIdentifier
	}
	if p.Volume.Article.Articleinfo.Articledoi != "" {
		is.DOI = p.Volume.Article.Articleinfo.Articledoi
		is.URL = append(is.URL, fmt.Sprintf("http://doi.org/%s", is.DOI))
	}

	is.Volume = p.Volume.Volumeinfo.Volumenum
	is.Issue = p.Volume.Article.Articleinfo.Issuenum
	is.Pages = p.Volume.Article.Articleinfo.Numpages
	is.Publishers = []string{"IEEE"}

	is.Subjects = []string{}
	for _, kw := range p.Volume.Article.Articleinfo.Keywordset.Keyword {
		is.Subjects = append(is.Subjects, kw.Term)
	}

	is.RefType = DefaultRefType

	return is, nil
}
