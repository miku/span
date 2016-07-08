package thieme

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/miku/span/finc"
)

type Article struct {
	xml.Name                  `xml:"article"`
	Xsi                       string `xml:"xsi,attr"`
	NoNamespaceSchemaLocation string `xml:"noNamespaceSchemaLocation,attr"`
	Lang                      string `xml:"lang,attr"`
	ArticleType               string `xml:"article-type,attr"`
	Front                     struct {
		JournalMeta struct {
			JournalId         string `xml:"journal-id"`
			JournalTitleGroup struct {
				Title string `xml:"journal-title"`
			} `xml:"journal-title-group"`
			Issn      []string `xml:"issn"`
			Publisher struct {
				Name string `xml:"publisher-name"`
			} `xml:"publisher"`
		} `xml:"journal-meta"`
		ArticleMeta struct {
			Abstract struct {
				Text string `xml:",innerxml"`
			} `xml:"abstract"`
			ArticleID []struct {
				Value string `xml:",chardata"`
				Type  string `xml:"pub-id-type,attr"`
			} `xml:"article-id"`
			ArticleCategories string `xml:"article-categories"`
			TitleGroup        struct {
				ArticleTitle    string `xml:"article-title"`
				TransTitleGroup struct {
					Lang       string `xml:"lang,attr"`
					TransTitle string `xml:"trans-title"`
				} `xml:"trans-title-group"`
			} `xml:"title-group"`
			ContribGroup string `xml:"contrib-group"`
			PubDate      struct {
				Month string `xml:"month"`
				Year  string `xml:"year"`
				Day   string `xml:"day"`
			} `xml:"pub-date"`
			Volume string `xml:"volume"`
			Issue  string `xml:"issue"`
			Fpage  string `xml:"fpage"`
			Lpage  string `xml:"lpage"`
		} `xml:"article-meta"`
	} `xml:"front"`
}

func (article Article) DOI() string {
	for _, id := range article.Front.ArticleMeta.ArticleID {
		if id.Type == "doi" {
			return id.Value
		}
	}
	return ""
}

func (article Article) RecordID() string {
	return fmt.Sprintf("ai-60-%s", base64.RawURLEncoding.EncodeToString([]byte(article.DOI())))
}

func (article Article) ParseTime() (time.Time, error) {
	p := article.Front.ArticleMeta.PubDate
	var year, month, day = "1970", "1", "1"
	if p.Year != "" {
		v, err := strconv.Atoi(p.Year)
		if err == nil {
			if v > 1000 && v < 3000 {
				year = strconv.Itoa(v)
			}
		}
	}
	if p.Month != "" {
		v, err := strconv.Atoi(p.Year)
		if err == nil {
			if v > 0 && v < 13 {
				month = strconv.Itoa(v)
			}
		}
	}
	if p.Day != "" {
		v, err := strconv.Atoi(p.Year)
		if err == nil {
			if v > 0 && v < 13 {
				day = strconv.Itoa(v)
			}
		}
	}
	return time.Parse("2006-01-02", fmt.Sprintf("%04s-%02s-%02s", year, month, day))
}

func (article Article) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()
	output.RecordID = article.RecordID()
	output.SourceID = SourceID
	output.DOI = article.DOI()
	if len(output.DOI) == 0 {
		return output, fmt.Errorf("empty DOI")
	}

	output.MegaCollection = Collection
	output.Genre = Genre
	output.Format = Format

	output.URL = []string{fmt.Sprintf("http://doi.org/%s", output.DOI)}
	output.Volume = article.Front.ArticleMeta.Volume
	output.Issue = article.Front.ArticleMeta.Issue
	output.ArticleTitle = article.Front.ArticleMeta.TitleGroup.ArticleTitle
	output.JournalTitle = article.Front.JournalMeta.JournalTitleGroup.Title
	output.ISSN = article.Front.JournalMeta.Issn
	output.StartPage = article.Front.ArticleMeta.Fpage
	output.EndPage = article.Front.ArticleMeta.Lpage
	output.Abstract = article.Front.ArticleMeta.Abstract.Text
	t, err := article.ParseTime()
	if err != nil {
		return output, err
	}
	output.Date = t
	output.RawDate = output.Date.Format("2006-01-02")
	output.RefType = DefaultRefType

	if article.Front.JournalMeta.Publisher.Name != "" {
		output.Publishers = []string{article.Front.JournalMeta.Publisher.Name}
	}
	return output, nil
}
