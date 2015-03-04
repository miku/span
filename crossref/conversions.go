package crossref

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/miku/span/finc"
)

var (
	ErrMissingUrl   = errors.New("document has no URL")
	ErrMissingTitle = errors.New("document has no title")
)

// ToInternalSchema converts a crossref document into an internal finc schema.
func (doc *Document) ToInternalSchema() (*finc.InternalSchema, error) {
	output := new(finc.InternalSchema)
	if doc.URL == "" {
		return output, ErrMissingUrl
	}

	output.RecordID = doc.RecordID()
	output.URL = append(output.URL, doc.URL)
	output.DOI = doc.DOI
	output.SourceID = fmt.Sprintf("%d", SourceID)
	output.Publisher = append(output.Publisher, doc.Publisher)
	output.ArticleTitle = doc.CombinedTitle()
	if output.ArticleTitle == "" {
		return output, ErrMissingTitle
	}
	output.Issue = doc.Issue
	output.Volume = doc.Volume
	output.ISSN = doc.ISSN

	if len(doc.ContainerTitle) > 0 {
		output.JournalTitle = doc.ContainerTitle[0]
	}

	for _, author := range doc.Authors {
		output.Authors = append(output.Authors, finc.Author{FirstName: author.Given, LastName: author.Family})
	}

	pi := doc.PageInfo()
	output.StartPage = fmt.Sprintf("%d", pi.StartPage)
	output.EndPage = fmt.Sprintf("%d", pi.EndPage)
	output.Pages = pi.RawMessage
	output.PageCount = fmt.Sprintf("%d", pi.PageCount())

	output.Date = doc.Issued.Date().Format("2006-01-02")

	name, err := doc.MemberName()
	if err == nil {
		output.MegaCollection = fmt.Sprintf("%s (CrossRef)", name)
	}
	return output, nil
}

// ToSolrSchema converts a single crossref document into a basic finc solr schema.
func (doc *Document) ToSolrSchema() (*finc.SolrSchema, error) {
	output := new(finc.SolrSchema)
	if doc.URL == "" {
		return output, ErrMissingUrl
	}

	output.ID = fmt.Sprintf("ai049%s", base64.StdEncoding.EncodeToString([]byte(doc.URL)))
	output.ISSN = doc.ISSN
	output.Publisher = doc.Publisher
	output.SourceID = "49"
	output.RecordType = "ai"
	output.Title = doc.CombinedTitle()
	output.TitleFull = doc.FullTitle()
	output.TitleShort = doc.ShortTitle()
	output.Topics = doc.Subject
	output.URL = doc.URL

	if len(doc.ContainerTitle) > 0 {
		output.HierarchyParentTitle = doc.ContainerTitle[0]
	}

	if doc.Type == "journal-article" {
		output.Format = "ElectronicArticle"
	}

	for _, author := range doc.Authors {
		output.SecondaryAuthors = append(output.SecondaryAuthors, author.String())
	}

	if len(output.SecondaryAuthors) > 1 {
		output.Author = output.SecondaryAuthors[0]
	}

	if doc.Issued.Year() > 0 {
		output.PublishDateSort = doc.Issued.Year()
	}

	output.Allfields = doc.Allfields()

	name, err := doc.MemberName()
	if err == nil {
		output.MegaCollection = []string{fmt.Sprintf("%s (CrossRef)", name)}
	}

	output.Fullrecord = "blob://" + output.ID

	return output, nil
}
