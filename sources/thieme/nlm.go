package thieme

import (
	"encoding/xml"

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
			JournalId         string   `xml:"journal-id"`
			JournalTitleGroup string   `xml:"journal-title-group"`
			Issn              []string `xml:"issn"`
			Publisher         string   `xml:"publisher"`
		} `xml:"journal-meta"`
		ArticleMeta struct {
			ArticleId         string `xml:"article-id"`
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
			} `xml:"pub-date"`
			Volume string `xml:"volume"`
			Issue  string `xml:"issue"`
			Fpage  string `xml:"fpage"`
			Lpage  string `xml:"lpage"`
		} `xml:"article-meta"`
	} `xml:"front"`
}

func (article Article) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()
	output.ArticleTitle = article.Front.ArticleMeta.TitleGroup.ArticleTitle
	output.JournalTitle = article.Front.JournalMeta.JournalTitleGroup
	output.ISSN = article.Front.JournalMeta.Issn
	if article.Front.JournalMeta.Publisher != "" {
		output.Publishers = []string{article.Front.JournalMeta.Publisher}
	}
	return output, nil
}
