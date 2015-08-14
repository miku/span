package thieme

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/miku/span"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
	"github.com/miku/span/jats"
	"golang.org/x/text/language"
)

const (
	SourceID   = "60"
	Collection = "Thieme"
	Genre      = "article"
	Format     = "ElectronicArticle"
)

type Thieme struct{}

// Iterate emits Converter elements via XML decoding.
func (s Thieme) Iterate(r io.Reader) (<-chan []span.Importer, error) {
	return span.FromXML(r, "article", func(d *xml.Decoder, se xml.StartElement) (span.Importer, error) {
		doc := new(Article)
		err := d.DecodeElement(&doc, &se)
		return doc, err
	})
}

type Article struct {
	jats.Article
}

func (article *Article) Identifiers() (jats.Identifiers, error) {
	var ids jats.Identifiers
	doi, err := article.DOI()
	if err != nil {
		return ids, err
	}
	locator := fmt.Sprintf("http://dx.doi.org/%s", doi)
	enc := fmt.Sprintf("ai-%s-%s", SourceID, base64.URLEncoding.EncodeToString([]byte(locator)))
	recordID := strings.TrimRight(enc, "=")
	return jats.Identifiers{DOI: doi, URL: locator, RecordID: recordID}, nil
}

// Authors returns the authors as slice.
func (article *Article) Authors() []finc.Author {
	var authors []finc.Author
	group := article.Front.Article.ContribGroup
	for _, contrib := range group.Contrib {
		if contrib.Type != "author" {
			continue
		}
		authors = append(authors, finc.Author{
			LastName:  contrib.StringName.Surname.Value,
			FirstName: contrib.StringName.GivenNames.Value})
	}
	return authors
}

// Languages returns a list of language in 3-letter format.
func (article *Article) Languages() []string {
	set := container.NewStringSet()
	for _, cm := range article.Front.Article.CustomMetaGroup.CustomMeta {
		if cm.Name.Value == "lang" {
			base, err := language.ParseBase(cm.Value.Value)
			if err == nil {
				set.Add(base.ISO3())
			}
		}
	}
	return set.Values()
}

// ToInternalSchema converts an article into an internal schema.
func (article *Article) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output, err := article.Article.ToIntermediateSchema()
	if err != nil {
		return output, err
	}

	ids, err := article.Identifiers()
	if err != nil {
		return output, err
	}
	output.DOI = ids.DOI

	id := ids.RecordID
	if len(id) > span.KeyLengthLimit {
		return output, span.Skip{Reason: fmt.Sprintf("id too long: %s", id)}
	}
	output.RecordID = id

	output.URL = append(output.URL, ids.URL)

	output.Authors = article.Authors()
	output.Format = Format
	output.Languages = article.Languages()
	output.MegaCollection = Collection
	output.SourceID = SourceID

	var normalized []string
	for _, issn := range output.ISSN {
		if len(issn) == 8 && !strings.Contains(issn, "-") {
			normalized = append(normalized, fmt.Sprintf("%s-%s", issn[:4], issn[4:]))
		}
	}
	output.ISSN = normalized

	// refs #5686
	for _, s := range output.Subjects {
		if s == "Book Reviews" {
			if output.ArticleTitle == "" {
				output.ArticleTitle = article.Front.Article.Product.Source.Value
			}
			break
		}
	}

	return output, nil
}
