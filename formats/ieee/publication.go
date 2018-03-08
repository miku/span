package ieee

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
)

const (
	SourceID       = "89"
	Format         = "ElectronicArticle"
	Collection     = "IEEE Xplore Library"
	Genre          = "article"
	DefaultRefType = "EJOUR"
)

var (
	ErrNoDate       = errors.New("no date found")
	ErrNoIdentifier = errors.New("missing identifier")
)

// Publication represents a publication.
type Publication struct {
	xml.Name          `xml:"publication"`
	Title             string `xml:"title"`
	TitleAbbreviation string `xml:"titleabbrev"`
	Normtitle         string `xml:"normtitle"`
	Publicationinfo   struct {
		Idamsid               string   `xml:"idamsid"`
		Publicationtype       string   `xml:"publicationtype"`
		Publicationsubtype    string   `xml:"publicationsubtype"`
		Ieeeabbrev            string   `xml:"ieeeabbrev"`
		Acronym               string   `xml:"acronym"`
		Pubstatus             string   `xml:"pubstatus"`
		Publicationopenaccess string   `xml:"publicationopenaccess"`
		StandardID            string   `xml:"standard_id"`
		Packagemembers        []string `xml:"packagememberset>packagemember"`
		Isbn                  []struct {
			Isbntype  string `xml:"isbntype,attr"`
			Mediatype string `xml:"mediatype,attr"`
			Value     string `xml:",chardata"`
		} `xml:"isbn"`
		BmsProductNumber struct {
			MediaType string `xml:"mediatype,attr"`
			Value     string `xml:",chardata"`
		}
		TCN  string `xml:"tcn"`
		Issn []struct {
			Mediatype string `xml:"mediatype,attr"`
			Value     string `xml:",chardata"`
		} `xml:"issn"`
		Pubtopicalbrowse []string `xml:"pubtopicalbrowseset>pubtopicalbrowse"`
		Copyrightgroup   struct {
			Copyright []struct {
				Year   string `xml:"year"`
				Holder string `xml:"holder"`
			}
		}
		PublisherNames []string `xml:"publisher>publishername"`
		Holdstatus     string   `xml:"holdstatus"`
		Confgroup      struct {
			ConfTitle string `xml:"conftitle"`
			Confdate  []struct {
				Confdatetype string `xml:"confdatetype,attr"`
				Year         string `xml:"year"`
				Month        string `xml:"month"`
				Day          string `xml:"day"`
			} `xml:"confdate"`
			Conflocation   string `xml:"conflocation"`
			Confcountry    string `xml:"confcountry"`
			ConferenceType string `xml:"conference_type"`
			DoiPermission  string `xml:"doi_permission"`
		} `xml:"confgroup"`
		Amsid string `xml:"amsid"`
		Coden string `xml:"coden"`
	} `xml:"publicationinfo"`
	Volume struct {
		Volumeinfo struct {
			Year      string `xml:"year"`
			Idamsid   string `xml:"idamsid"`
			Notegroup string `xml:"notegroup"`
			Issue     struct {
				Amsid         string `xml:"amsid"`
				Amscreatedate string `xml:"amscreatedate"`
				Issuestatus   string `xml:"issuestatus"`
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
				ArticleLicense         string `xml:"articlelicense"`
				ArticleLicenseURI      string `xml:"article_license_uri"`
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
				Date []struct {
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
				Csarticleid   string `xml:"csarticleid"`
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

// PaperISSN returns the online ISSN found in the publication.
func (p Publication) PaperISSN() (issns []string) {
	for _, issn := range p.Publicationinfo.Issn {
		if strings.ToLower(issn.Mediatype) == "paper" {
			issns = append(issns, issn.Value)
		}
	}
	return
}

// OnlineISSN returns the online ISSN found in the publication.
func (p Publication) OnlineISSN() (issns []string) {
	for _, issn := range p.Publicationinfo.Issn {
		if strings.ToLower(issn.Mediatype) == "online" {
			issns = append(issns, issn.Value)
		}
	}
	return
}

// ISBNList returns a list of ISBN.
func (p Publication) ISBNList() (isbns []string) {
	for _, isbn := range p.Publicationinfo.Isbn {
		isbns = append(isbns, isbn.Value)
	}
	return
}

// Date returns the date of the publication. There are multiple dates, "LastEdit",
// "LastInspecUpd", "OriginalPub", "ePub" - use OriginalPub.
func (p Publication) Date() (time.Time, error) {
	if len(p.Volume.Article.Articleinfo.Date) == 0 {
		return time.Time{}, ErrNoDate
	}

	// Use the first date as default.
	date := p.Volume.Article.Articleinfo.Date[0]

	for _, dt := range p.Volume.Article.Articleinfo.Date {
		if dt.Datetype == "OriginalPub" {
			date = dt
			break
		}
	}

	y, m, d := "1970", "Jan", "1"
	if date.Year == "" {
		return time.Time{}, ErrNoDate
	}
	if _, err := strconv.Atoi(date.Year); err != nil {
		return time.Time{}, err
	}
	y = date.Year
	if date.Month != "" {
		if len(date.Month) > 3 {
			date.Month = date.Month[:3]
		}
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

	// try various patterns
	patterns := []string{
		"2006-Jan-02",
		"2006-01-02",
		"2006-1-02",
	}

	for _, p := range patterns {
		if t, err := time.Parse(p, fmt.Sprintf("%s-%s-%02s", y, m, d)); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf(fmt.Sprintf("%s", p.Volume.Article.Articleinfo.Date))
}

// Authors returns authors.
func (p Publication) Authors() []finc.Author {
	var authors []finc.Author
	for _, author := range p.Volume.Article.Articleinfo.Authorgroup.Author {
		authors = append(authors, finc.Author{FirstName: author.Firstname, LastName: author.Surname})
	}
	return authors
}

// PageCount the page count as string.
func (p Publication) PageCount() string {
	start, err := strconv.Atoi(p.Volume.Article.Articleinfo.Artpagenums.Startpage)
	if err != nil {
		return ""
	}
	end, err := strconv.Atoi(p.Volume.Article.Articleinfo.Artpagenums.Endpage)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", end-start+1)
}

// ToIntermediateSchema does a type conversion only.
func (p Publication) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	is := finc.NewIntermediateSchema()
	is.JournalTitle = p.Title
	is.ArticleTitle = p.Volume.Article.Title

	if strings.HasPrefix(is.ArticleTitle, "[") {
		return is, span.Skip{Reason: fmt.Sprintf("extra content: %s", is.ArticleTitle)}
	}

	is.ISSN = p.PaperISSN()
	is.EISSN = p.OnlineISSN()
	is.ISBN = p.ISBNList()

	is.Abstract = p.Volume.Article.Articleinfo.Abstract

	date, err := p.Date()
	if err != nil {
		log.Printf("date problem: %s: %s", err, is.ArticleTitle)
		return is, span.Skip{Reason: err.Error()}
	}
	is.Date = date
	is.RawDate = date.Format("2006-01-02")

	is.Authors = p.Authors()

	is.URL = []string{}

	is.SourceID = SourceID
	is.MegaCollections = []string{Collection}
	is.Format = Format

	if p.Volume.Article.Articleinfo.Amsid != "" {
		is.URL = append(is.URL, fmt.Sprintf("http://ieeexplore.ieee.org/stamp/stamp.jsp?arnumber=%s", p.Volume.Article.Articleinfo.Amsid))
		is.ID = fmt.Sprintf("ai-89-%s", base64.RawURLEncoding.EncodeToString([]byte(p.Volume.Article.Articleinfo.Amsid)))
		is.RecordID = p.Volume.Article.Articleinfo.Amsid
	} else {
		log.Printf("warning: no identifier: %s", is.ArticleTitle)
		return is, ErrNoIdentifier
	}
	if p.Volume.Article.Articleinfo.Articledoi != "" {
		is.DOI = p.Volume.Article.Articleinfo.Articledoi
		is.URL = append(is.URL, fmt.Sprintf("http://doi.org/%s", is.DOI))
	}

	is.Volume = p.Volume.Volumeinfo.Volumenum
	is.Issue = p.Volume.Article.Articleinfo.Issuenum

	is.StartPage = p.Volume.Article.Articleinfo.Artpagenums.Startpage
	is.EndPage = p.Volume.Article.Articleinfo.Artpagenums.Endpage
	is.Pages = fmt.Sprintf("%s-%s", is.StartPage, is.EndPage)
	is.PageCount = p.PageCount()

	is.Publishers = []string{"IEEE"}

	is.Subjects = []string{}
	for _, kw := range p.Volume.Article.Articleinfo.Keywordset.Keyword {
		term := strings.TrimSpace(kw.Term)
		if term == "" {
			continue
		}
		is.Subjects = append(is.Subjects, term)
	}

	is.RefType = DefaultRefType

	is.Packages = []string{
		p.Publicationinfo.Publicationtype,
		p.Publicationinfo.Publicationsubtype,
	}

	is.Packages = append(is.Packages, p.Publicationinfo.Packagemembers...)

	// https://supportcenter.ieee.org/app/answers/detail/a_id/1900/~/how-is-the-oapa-different-from-a-cc-by-license%3F
	if p.Volume.Article.Articleinfo.ArticleLicense == "CCBY" {
		is.OpenAccess = true
	}

	return is, nil
}
