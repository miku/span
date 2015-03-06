package jats

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span/finc"
	"github.com/miku/span/holdings"
)

// SourceID for internal bookkeeping.
const SourceID = 50

// Jats source.
type Jats struct{}

// Iterate emits Converter elements via XML decoding.
func (s Jats) Iterate(r io.Reader) (chan interface{}, error) {
	ch := make(chan interface{})
	go func() {
		decoder := xml.NewDecoder(bufio.NewReader(r))
		for {
			t, _ := decoder.Token()
			if t == nil {
				break
			}
			switch se := t.(type) {
			case xml.StartElement:
				if se.Name.Local == "article" {
					doc := new(Article)
					err := decoder.DecodeElement(&doc, &se)
					if err != nil {
						log.Fatal(err)
					}
					ch <- doc
				}
			}
		}
		close(ch)
	}()
	return ch, nil
}

// Article mirrors a JATS article element. Some elements, such as
// article categories are not implmented yet.
type Article struct {
	XMLName xml.Name `xml:"article"`
	Front   struct {
		XMLName xml.Name `xml:"front"`
		Journal struct {
			ID struct {
				XMLName xml.Name `xml:"journal-id"`
				Type    string   `xml:"journal-id-type,attr"`
				Value   string   `xml:",chardata"`
			}
			ISSN []struct {
				Type  string `xml:"pub-type,attr"`
				Value string `xml:",chardata"`
			} `xml:"issn"`
			TitleGroup struct {
				XMLName          xml.Name `xml:"journal-title-group"`
				AbbreviatedTitle struct {
					XMLName xml.Name `xml:"abbrev-journal-title"`
					Title   string   `xml:",chardata"`
					Type    string   `xml:"abbrev-type,attr"`
				}
			}
			Publisher struct {
				XMLName xml.Name `xml:"publisher"`
				Name    struct {
					XMLName xml.Name `xml:"publisher-name"`
					Value   string   `xml:",chardata"`
				}
			}
		} `xml:"journal-meta"`
		Article struct {
			XMLName xml.Name `xml:"article-meta"`
			ID      []struct {
				Type  string `xml:"pub-id-type,attr"`
				Value string `xml:",chardata"`
			} `xml:"article-id"`
			TitleGroup struct {
				XMLName xml.Name `xml:"title-group"`
				Title   struct {
					XMLName xml.Name `xml:"article-title"`
					Value   string   `xml:",chardata"`
				}
				Subtitle struct {
					XMLName xml.Name `xml:"subtitle"`
					Value   string   `xml:",chardata"`
				}
			}
			ContribGroup struct {
				XMLName xml.Name `xml:"contrib-group"`
				Contrib []struct {
					Type string `xml:"contrib-type,attr"`
					Name struct {
						XMLName xml.Name `xml:"name"`
						Surname struct {
							XMLName xml.Name `xml:"surname"`
							Value   string   `xml:",chardata"`
						}
						GivenNames struct {
							XMLName xml.Name `xml:"given-names"`
							Value   string   `xml:",chardata"`
						}
					}
				} `xml:"contrib"`
			}
			PubDate struct {
				Type  string `xml:"pub-type,attr"`
				Month struct {
					XMLName xml.Name `xml:"month"`
					Value   string   `xml:",chardata"`
				}
				Year struct {
					XMLName xml.Name `xml:"year"`
					Value   string   `xml:",chardata"`
				}
				Day struct {
					XMLName xml.Name `xml:"day"`
					Value   string   `xml:",chardata"`
				}
			} `xml:"pub-date"`
			Volume struct {
				XMLName xml.Name `xml:"volume"`
				Value   string   `xml:",chardata"`
			}
			Issue struct {
				XMLName xml.Name `xml:"issue"`
				Value   string   `xml:",chardata"`
			}
			FirstPage struct {
				XMLName xml.Name `xml:"fpage"`
				Value   string   `xml:",chardata"`
			}
			LastPage struct {
				XMLName xml.Name `xml:"lpage"`
				Value   string   `xml:",chardata"`
			}
			Permissions struct {
				XMLName       xml.Name `xml:"permissions"`
				CopyrightYear struct {
					XMLName xml.Name `xml:"copyright-year"`
					Value   string   `xml:",chardata"`
				}
				CopyrightStatement struct {
					XMLName xml.Name `xml:"copyright-statement"`
					Value   string   `xml:",chardata"`
				}
			}
			Abstract struct {
				XMLName xml.Name `xml:"abstract"`
				Value   string   `xml:",innerxml"`
				Lang    string   `xml:"lang,attr"`
			}
		}
	}
}

// Authors returns the authors as slice.
// TODO(miku): get rid of cross-format dependency.
func (article *Article) Authors() []finc.Author {
	var authors []finc.Author
	group := article.Front.Article.ContribGroup
	for _, contrib := range group.Contrib {
		if contrib.Type != "author" {
			continue
		}
		authors = append(authors, finc.Author{LastName: contrib.Name.Surname.Value, FirstName: contrib.Name.GivenNames.Value})
	}
	return authors
}

// CombinedTitle returns a longish title.
func (article *Article) CombinedTitle() string {
	group := article.Front.Article.TitleGroup
	if group.Title.Value != "" {
		if group.Subtitle.Value != "" {
			return fmt.Sprintf("%s : %s", group.Title.Value, group.Subtitle.Value)
		}
		return group.Title.Value
	}
	if group.Subtitle.Value != "" {
		return group.Subtitle.Value
	}
	return ""
}

// DOI is a convenience shortcut to get the DOI.
func (article *Article) DOI() (string, error) {
	for _, id := range article.Front.Article.ID {
		if id.Type == "doi" {
			return id.Value, nil
		}
	}
	return "", fmt.Errorf("article has no DOI")
}

// ISSN returns a list of ISSNs associated with this article.
func (article *Article) ISSN() []string {
	var issns []string
	for _, issn := range article.Front.Journal.ISSN {
		issns = append(issns, issn.Value)
	}
	return issns
}

// PageCount return the number of pages as string.
func (article *Article) PageCount() string {
	first, err := strconv.Atoi(article.Front.Article.FirstPage.Value)
	if err != nil {
		return ""
	}
	last, err := strconv.Atoi(article.Front.Article.LastPage.Value)
	if err != nil {
		return ""
	}
	if last-first > 0 {
		return fmt.Sprintf("%d", last-first)
	}
	return ""
}

// defaultString returns a default if s is the empty string.
func defaultString(s, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}

func (article *Article) Abstract() string {
	return string(article.Front.Article.Abstract.Value)
}

// Date returns this articles issuing date in a best effort manner.
func (article *Article) Date() time.Time {
	pubdate := article.Front.Article.PubDate
	day := defaultString(pubdate.Day.Value, "01")
	month := defaultString(pubdate.Month.Value, "01")
	year := defaultString(pubdate.Year.Value, "1970")
	t, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%02s-%02s", year, month, day))
	if err != nil {
		// log.Printf("%+v", article)
		log.Println(err)
		t, _ = time.Parse("2006-01-02", "1970-01-01")
	}
	return t
}

func (article *Article) Year() int {
	year, err := strconv.Atoi(article.Front.Article.PubDate.Year.Value)
	if err != nil {
		return 0
	}
	return year
}

func (article *Article) Allfields() string {
	doi, _ := article.DOI()
	var authors []string
	for _, author := range article.Authors() {
		authors = append(authors, author.String())
	}
	fields := [][]string{article.ISSN(), authors, []string{article.CombinedTitle(),
		doi, article.Front.Journal.TitleGroup.AbbreviatedTitle.Title,
		article.Front.Article.Volume.Value, article.Front.Article.Issue.Value}}

	var buf bytes.Buffer
	for _, f := range fields {
		for _, value := range f {
			for _, token := range strings.Fields(value) {
				_, err := buf.WriteString(fmt.Sprintf("%s ", strings.TrimSpace(token)))
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
	return strings.TrimSpace(buf.String())
}

// ToInternalSchema converts a jats article into an internal schema.
func (article *Article) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := new(finc.IntermediateSchema)
	doi, err := article.DOI()
	if err != nil {
		return output, err
	}
	articleURL := fmt.Sprintf("http://dx.doi.org/%s", doi)

	output.RecordID = fmt.Sprintf("ai050%s", base64.StdEncoding.EncodeToString([]byte(articleURL)))
	output.URL = append(output.URL, articleURL)
	output.DOI = doi
	output.SourceID = "50"
	output.Publisher = append(output.Publisher, article.Front.Journal.Publisher.Name.Value)
	output.ArticleTitle = article.CombinedTitle()
	output.Issue = article.Front.Article.Issue.Value
	output.Volume = article.Front.Article.Volume.Value
	output.ISSN = article.ISSN()
	output.JournalTitle = article.Front.Journal.TitleGroup.AbbreviatedTitle.Title

	output.Authors = article.Authors()
	output.Abstract = article.Abstract()

	output.StartPage = article.Front.Article.FirstPage.Value
	output.EndPage = article.Front.Article.LastPage.Value
	output.Pages = fmt.Sprintf("%s-%s", output.StartPage, output.EndPage)
	output.PageCount = article.PageCount()

	output.RawDate = article.Date().Format("2006-01-02")

	output.MegaCollection = "DeGruyter SSH"
	return output, nil
}

// ToSolrSchema converts a single jats article into a basic finc solr schema.
func (article *Article) ToSolrSchema() (*finc.SolrSchema, error) {
	output := new(finc.SolrSchema)
	doi, err := article.DOI()
	if err != nil {
		return output, err
	}
	articleURL := fmt.Sprintf("http://dx.doi.org/%s", doi)

	output.ID = fmt.Sprintf("ai050%s", base64.StdEncoding.EncodeToString([]byte(articleURL)))
	output.ISSN = article.ISSN()
	output.Publisher = article.Front.Journal.Publisher.Name.Value
	output.SourceID = "50"
	output.RecordType = "ai"
	output.Title = article.CombinedTitle()
	output.TitleFull = article.CombinedTitle()
	output.TitleShort = article.Front.Article.TitleGroup.Title.Value
	// output.Topics = doc.Subject // TODO(miku): article-categories
	output.URL = articleURL

	output.HierarchyParentTitle = article.Front.Journal.TitleGroup.AbbreviatedTitle.Title
	output.Format = "ElectronicArticle"

	if len(article.Authors()) > 0 {
		output.Author = article.Authors()[0].String()
		for _, author := range article.Authors() {
			output.SecondaryAuthors = append(output.SecondaryAuthors, author.String())
		}
	}

	if article.Year() > 0 {
		output.PublishDateSort = article.Year()
	}

	output.Allfields = article.Allfields()
	output.MegaCollection = []string{"DeGruyter SSH"}
	output.Fullrecord = "blob://id/" + output.ID

	return output, nil
}

// Institutions, TODO(miku): implement lookup
func (article *Article) Institutions(iih holdings.IsilIssnHolding) []string {
	return []string{}
}
