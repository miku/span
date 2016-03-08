//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                    The Finc Authors, http://finc.info
//                    Martin Czygan, <martin.czygan@uni-leipzig.de>
//
// This file is part of some open source application.
//
// Some open source application is free software: you can redistribute
// it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation, either
// version 3 of the License, or (at your option) any later version.
//
// Some open source application is distributed in the hope that it will
// be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
//
// @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
//
package jstor

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/kennygrant/sanitize"
	"github.com/miku/span"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
	"github.com/miku/span/sources/jats"
	"golang.org/x/text/language"
)

const (
	// SourceID for internal bookkeeping.
	SourceID = "55"
	// SourceName for finc.mega_collection.
	SourceName = "JSTOR"
	// Format for intermediate schema.
	Format = "ElectronicArticle"
)

var (
	// ArticleTitleBlockPatterns
	ArticleTitleBlockPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(front|back)\s*matter`),
		regexp.MustCompile(`(?i)table\s*of\s*content[s]?`),
	}
)

// Jstor source.
type Jstor struct{}

// Article with extras for this source.
type Article struct {
	jats.Article
}

// Iterate emits Converter elements via XML decoding.
func (s Jstor) Iterate(r io.Reader) (<-chan []span.Importer, error) {
	return span.FromXML(r, "article", func(d *xml.Decoder, se xml.StartElement) (span.Importer, error) {
		article := new(Article)
		err := d.DecodeElement(&article, &se)
		return article, err
	})
}

// Identifiers returns the doi and the dependent url and recordID in a struct.
// Records from this source do not need a DOI necessarily.
func (article *Article) Identifiers() (jats.Identifiers, error) {
	locator := article.Front.Article.SelfURI.Value

	// guess DOI from jstor stable URL
	var candidate, doi string
	candidate = strings.Replace(locator, "http://www.jstor.org/stable/", "", -1)
	if strings.Contains(candidate, "/") {
		doi = candidate
	}

	recordID := fmt.Sprintf("ai-%s-%s", SourceID, base64.RawURLEncoding.EncodeToString([]byte(locator)))
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

// ReviewedProduct returns the string of the reviewed thing in a best-effort way.
func (article *Article) ReviewedProduct() string {
	if len(article.Front.Article.Products) == 0 {
		return ""
	}
	if article.Front.Article.Products[0].Source.Value != "" {
		return strings.TrimSpace(sanitize.HTML(article.Front.Article.Products[0].Source.Value))
	}
	// refs. #7111
	if article.Front.Article.Products[0].StringName.Value != "" {
		return strings.TrimSpace(sanitize.HTML(article.Front.Article.Products[0].StringName.Value))
	}
	return ""
}

// ToInternalSchema converts an article into an internal schema. There are a
// couple of content-dependent choices here.
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
	output.MegaCollection = SourceName
	output.SourceID = SourceID

	var normalized []string
	for _, issn := range output.ISSN {
		if len(issn) == 8 && !strings.Contains(issn, "-") {
			normalized = append(normalized, fmt.Sprintf("%s-%s", issn[:4], issn[4:]))
		}
	}
	output.ISSN = normalized

	// refs #5686
	if output.Date.IsZero() {
		return output, span.Skip{Reason: fmt.Sprintf("zero date: %s", output.RecordID)}
	}

	// refs #5686
	for _, p := range ArticleTitleBlockPatterns {
		if p.MatchString(output.ArticleTitle) {
			return output, span.Skip{Reason: fmt.Sprintf("title blacklisted: %s", output.ArticleTitle)}
		}
	}

	// refs #5686, approx. article type distribution: https://git.io/vzlCr
	switch article.Type {
	case "book-review", "book-reviews", "Book Review":
		output.ArticleTitle = fmt.Sprintf("Review: %s", article.ReviewedProduct())
	case "misc", "other", "front-matter", "back-matter", "announcement", "font-matter", "fm", "fornt-matter":
		return output, span.Skip{Reason: fmt.Sprintf("suppressed format: %s", article.Type)}
	}

	return output, nil
}
