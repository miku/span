package ceeol

import (
	"encoding/xml"
	"fmt"
	"strings"

	"time"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
	"github.com/shantanubhadoria/go-roman/roman"
)

const (
	SourceIdentifier      = "53"
	Format                = "ElectronicArticle"
	Genre                 = "article"
	DefaultRefType        = "EJOUR"
	Collection            = "CEEOL Central and Eastern European Online Library"
	TechnicalCollectionID = "sid-53-col-ceeol"

	// BaseURL is prepended to ArticleURL values that only contain a path.
	BaseURL = "https://www.ceeol.com"
)

// Article from CEEOL, refs #9398.
type Article struct {
	XMLName                 xml.Name `xml:"Article"`
	UniqueID                string   `xml:"UniqueID"`
	ISSN                    string   `xml:"ISSN"`
	EISSN                   string   `xml:"eISSN"`
	PublicationTitle        string   `xml:"PublicationTitle"`
	PublicationTitleEnglish string   `xml:"PublicationTitleEnglish"`
	ArticleTitle            string   `xml:"ArticleTitle"`
	ArticleTitleEnglish     string   `xml:"ArticleTitleEnglish"`
	IsOpenAccess            string   `xml:"IsOpenAccess"`
	PublicationYear         string   `xml:"PublicationYear"`
	Volume                  string   `xml:"Volume"`
	Issue                   string   `xml:"Issue"`
	StartPage               string   `xml:"StartPage"`
	ArticleURL              string   `xml:"ArticleURL"`
	Authors                 []string `xml:"Authors>Author"`
	Languages               []string `xml:"Languages>Language"`
	ArticleSubtitle         string   `xml:"ArticleSubtitle"`
	PublicationSubtitle     string   `xml:"PublicationSubtitle"`
	EndPage                 string   `xml:"EndPage"`
	PageCount               string   `xml:"PageCount"`
	SubjectTerms            []string `xml:"SubjectTerms>SubjectTerm"`
	Publisher               string   `xml:"Publisher"`
	PublisherEnglish        string   `xml:"PublisherEnglish"`
	Keywords                string   `xml:"Keywords"`
	Description             string   `xml:"Description"`
	FileID                  string   `xml:"FileID"`
}

func normalizeString(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

// absoluteURL returns an absolute link to an article. Depending on the
// snapshot, ArticleURL contains either a full URL, which is passed through
// unchanged, or just a path, like "/search/article-detail?id=6765", which gets
// BaseURL prepended.
func absoluteURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	return BaseURL + "/" + strings.TrimLeft(s, "/")
}

// ToIntermediateSchema converts an article to intermediate schema.
func (article *Article) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()
	if strings.TrimSpace(article.ArticleTitleEnglish) == "" ||
		normalizeString(article.ArticleTitleEnglish) == normalizeString(article.ArticleTitle) {
		output.ArticleTitle = strings.TrimSpace(article.ArticleTitle)
	} else {
		output.ArticleTitle = fmt.Sprintf("%s [%s]",
			strings.TrimSpace(article.ArticleTitle),
			strings.TrimSpace(article.ArticleTitleEnglish))
	}
	output.ISSN = append(output.ISSN, article.ISSN)
	output.EISSN = append(output.EISSN, article.EISSN)
	v, err := roman.ToIndoArabic(article.Volume)
	if err != nil {
		output.Volume = article.Volume
	} else {
		output.Volume = fmt.Sprintf("%d", v)
	}
	if strings.TrimSpace(article.PublicationTitleEnglish) == "" ||
		normalizeString(article.PublicationTitleEnglish) == normalizeString(article.PublicationTitle) {
		output.JournalTitle = strings.TrimSpace(article.PublicationTitle)
	} else {
		output.JournalTitle = fmt.Sprintf("%s [%s]", strings.TrimSpace(article.PublicationTitle),
			strings.TrimSpace(article.PublicationTitleEnglish))
	}
	if article.IsOpenAccess != "0" {
		output.OpenAccess = true
	}
	output.Issue = article.Issue
	output.StartPage = article.StartPage
	output.EndPage = article.EndPage
	output.PageCount = article.PageCount
	output.Abstract = article.Description
	output.Publishers = append(output.Publishers, article.Publisher)
	if article.PublisherEnglish != "" && article.PublisherEnglish != article.Publisher {
		output.Publishers = append(output.Publishers, article.PublisherEnglish)
	}
	for _, author := range article.Authors {
		name := strings.TrimSpace(author)
		if len(name) < 4 {
			continue
		}
		// Simple blacklist, refs #9398.
		if strings.HasPrefix(name, "No Author Specified") ||
			strings.HasPrefix(name, "Miscellaneous, Miscellaneous") ||
			strings.HasPrefix(name, "Anonymous, Anonymous") ||
			strings.HasPrefix(name, "Various, Authors") ||
			strings.HasPrefix(name, "TOL, TOL") {
			continue
		}
		output.Authors = append(output.Authors, finc.Author{Name: name})
	}
	output.RawDate = fmt.Sprintf("%s-01-01", article.PublicationYear)
	output.Date, err = time.Parse("2006-01-02", fmt.Sprintf("%s-01-01", article.PublicationYear))
	if err != nil {
		return nil, fmt.Errorf("ceeol: invalid date: %w", err)
	}
	output.Subjects = article.SubjectTerms
	if u := absoluteURL(article.ArticleURL); u != "" {
		output.URL = append(output.URL, u)
	}
	output.RecordID = article.UniqueID
	output.ID = fmt.Sprintf("ai-%s-%s", SourceIdentifier, article.UniqueID)
	output.SourceID = SourceIdentifier
	output.Format = Format
	output.Genre = Genre
	output.MegaCollections = []string{Collection, TechnicalCollectionID}
	output.RefType = DefaultRefType
	for _, lang := range article.Languages {
		if isocode := span.LanguageIdentifier(lang); isocode != "" {
			output.Languages = append(output.Languages, isocode)
		}
	}
	return output, nil
}
