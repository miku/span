package jats

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/miku/span/finc"
)

// Article mirrors a JATS article element. Some elements, such as
// article categories are not implmented yet.
type Article struct {
	XMLName xml.Name `xml:"article"`
	Front   struct {
		JournalMeta struct {
			JournalID struct {
				Type string `xml:"journal-id-type,attr"`
				ID   string `xml:",chardata"`
			} `xml:"journal-id"`
			ISSN []struct {
				Type  string `xml:"pub-type,attr"`
				Value string `xml:",chardata"`
			} `xml:"issn"`
			JournalTitleGroup struct {
				AbbreviatedJournalTitle struct {
					Value string `xml:",chardata"`
					Type  string `xml:"abbrev-type,attr"`
				} `xml:"abbrev-journal-title"`
			} `xml:"journal-title-group"`
			Publisher struct {
				Name struct {
					Value string `xml:",chardata"`
				} `xml:"publisher-name"`
			} `xml:"publisher"`
		} `xml:"journal-meta"`
		ArticleMeta struct {
			ArticleID []struct {
				Type  string `xml:"pub-id-type,attr"`
				Value string `xml:",chardata"`
			} `xml:"article-id"`
			TitleGroup struct {
				Title struct {
					Value string `xml:",chardata"`
				} `xml:"article-title"`
				Subtitle struct {
					Value string `xml:",chardata"`
				} `xml:"subtitle"`
			} `xml:"title-group"`
			ContribGroup struct {
				Contrib []struct {
					Type string `xml:"contrib-type,attr"`
					Name struct {
						Surname struct {
							Value string `xml:",chardata"`
						} `xml:"surname"`
						GivenNames struct {
							Value string `xml:",chardata"`
						} `xml:"given-names"`
					} `xml:"name"`
				} `xml:"contrib"`
			} `xml:"contrib-group"`
			PubDate struct {
				Type  string `xml:"pub-type,attr"`
				Month struct {
					Value string `xml:",chardata"`
				} `xml:"month"`
				Year struct {
					Value string `xml:",chardata"`
				} `xml:"year"`
				Day struct {
					Value string `xml:",chardata"`
				} `xml:"day"`
			} `xml:"pub-date"`
			Volume struct {
				Value string `xml:",chardata"`
			} `xml:"volume"`
			Issue struct {
				Value string `xml:",chardata"`
			} `xml:"issue"`
			FirstPage struct {
				Value string `xml:",chardata"`
			} `xml:"fpage"`
			LastPage struct {
				Value string `xml:",chardata"`
			} `xml:"lpage"`
			Permissions struct {
				CopyrightYear struct {
					Value string `xml:",chardata"`
				} `xml:"copyright-year"`
				CopyrightStatement struct {
					Value string `xml:",chardata"`
				} `xml:"copyright-statement"`
			} `xml:"permissions"`
		} `xml:"article-meta"`
	} `xml:"front"`
}

func (article *Article) Authors() []string {
	var authors []string
	group := article.Front.ArticleMeta.ContribGroup
	for _, contrib := range group.Contrib {
		if contrib.Type != "author" {
			continue
		}
		authors = append(authors, fmt.Sprintf("%s, %s", contrib.Name.Surname.Value, contrib.Name.GivenNames.Value))
	}
	return authors
}

// CombinedTitle returns a longish title.
func (article *Article) CombinedTitle() string {
	group := article.Front.ArticleMeta.TitleGroup
	if group.Title.Value != "" {
		if group.Subtitle.Value != "" {
			return fmt.Sprintf("%s : %s", group.Title.Value, group.Subtitle.Value)
		}
		return group.Title.Value
	}
	if group.Subtitle.Value != "" {
		group.Subtitle.Value
	}
	return ""
}

func (article *Article) DOI() (string, error) {
	for _, id := range article.Front.ArticleMeta.ArticleID {
		if id.Type == "doi" {
			return id.Value, nil
		}
	}
	return "", fmt.Errorf("article has no DOI")
}

func (article *Article) ISSN() []string {
	var issns []string
	for _, issn := range article.Front.JournalMeta.ISSN {
		issns = append(issns, issn.Value)
	}
	return issns
}

func (article *Article) PageCount() string {
	first, err := strconv.Atoi(article.Front.ArticleMeta.FirstPage.Value)
	if err != nil {
		return ""
	}
	last, err := strconv.Atoi(article.Front.ArticleMeta.LastPage.Value)
	if err != nil {
		return ""
	}
	if last-first > 0 {
		return fmt.Sprintf("%d", last-first)
	}
	return ""
}

func defaultString(s, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}

func (article *Article) Date() time.Time {
	pubdate := article.Front.ArticleMeta.PubDate
	day := defaultString(pubdate.Day, "01")
	month := defaultString(pubdate.Month, "01")
	year := defaultString(pubdate.Year, "1970")
	t, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", year, month, day))
	if err != nil {
		log.Fatal(err)
	}
	return t
}

// ToSchema converts a jats article into an intermediate schema.
func (article *Article) ToSchema() (output finc.Schema, err error) {

	doi, err := article.DOI()
	if err != nil {
		return output, err
	}
	articleURL := fmt.Sprintf("http://dx.doi.org/%s", doi)

	output.RecordID = fmt.Sprintf("ai050%s", base64.StdEncoding.EncodeToString([]byte(articleURL)))
	output.URL = append(output.URL, articleURL)
	output.DOI = doi
	output.SourceID = "50"
	output.Publisher = append(output.Publisher, article.Front.JournalMeta.Publisher.Name.Value)
	output.ArticleTitle = article.CombinedTitle()
	output.Issue = article.Front.ArticleMeta.Issue.Value
	output.Volume = article.Front.ArticleMeta.Volume.Value
	output.ISSN = article.ISSN()
	output.JournalTitle = article.Front.JournalMeta.JournalTitleGroup.AbbreviatedJournalTitle.Value

	output.Authors = article.Authors()

	output.StartPage = article.Front.ArticleMeta.FirstPage.Value
	output.EndPage = article.Front.ArticleMeta.EndPage.Value
	output.Pages = fmt.Sprintf("%s-%s", output.StartPage, output.EndPage)
	output.PageCount = article.PageCount()

	output.Date = article.Date().Format("2006-01-02")

	output.MegaCollection = "DeGruyter SSH"
	return output, nil
}
