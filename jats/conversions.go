package jats

import (
	"encoding/base64"
	"fmt"

	"github.com/miku/span/finc"
)

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
	output.Publisher = append(output.Publisher, article.Front.Journal.Publisher.Name.Value)
	output.ArticleTitle = article.CombinedTitle()
	output.Issue = article.Front.Article.Issue.Value
	output.Volume = article.Front.Article.Volume.Value
	output.ISSN = article.ISSN()
	output.JournalTitle = article.Front.Journal.TitleGroup.AbbreviatedTitle.Title

	output.Authors = article.Authors()

	output.StartPage = article.Front.Article.FirstPage.Value
	output.EndPage = article.Front.Article.LastPage.Value
	output.Pages = fmt.Sprintf("%s-%s", output.StartPage, output.EndPage)
	output.PageCount = article.PageCount()

	output.Date = article.Date().Format("2006-01-02")

	output.MegaCollection = "DeGruyter SSH"
	return output, nil
}

// ToSolrSchema converts a single jats article into a basic finc solr schema.
func (article *Article) ToSolrSchema() (output finc.SolrSchema, err error) {
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
