package ios

//	Copyright 2023 by Leipzig University Library, http://ub.uni-leipzig.de
//	                  The Finc Authors, http://finc.info
//	                  Martin Czygan, <martin.czygan@uni-leipzig.de>
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

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/miku/span"
	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/formats/jats"
	"golang.org/x/text/language"
)

const (
	// SourceID for internal bookkeeping, refs #24731
	SourceID = "219"
	// SourceName for finc.mega_collection.
	SourceName = "IOS Press"
	// Format for intermediate schema.
	Format = "ElectronicArticle"
)

var (
	// ArticleTitleBlockPatterns
	ArticleTitleBlockPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(front|back)\s*matter`),
		regexp.MustCompile(`(?i)table\s*of\s*content[s]?`),
	}
	DOIPattern = regexp.MustCompile(`10\.[0-9]+\/\S+`)
)

type Article struct {
	XMLName xml.Name `xml:"article"`
	jats.Article
}

// Identifiers returns identifiers.
func (article *Article) Identifiers() (jats.Identifiers, error) {
	doi, err := article.DOI()
	if err != nil {
		return jats.Identifiers{}, err
	}
	locator := fmt.Sprintf("https://doi.org/%s", doi)
	id := fmt.Sprintf("ai-%s-%s", SourceID, base64.RawURLEncoding.EncodeToString([]byte(doi)))
	return jats.Identifiers{DOI: doi, URL: locator, ID: id}, nil
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

// ToIntermediateSchema converts an article into an internal schema. There are a
// couple of content-dependent choices here.
func (article *Article) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output, err := article.Article.ToIntermediateSchema()
	if err != nil {
		// TODO: in this case, conversion cannot fail
		return output, err
	}
	ids, err := article.Identifiers()
	if err == jats.ErrNoDOI {
		return output, span.Skip{
			Reason: fmt.Sprintf("missing doi, ids: %v", article.Front.Article.ID),
		}
	}
	if err != nil {
		return output, err
	}
	output.DOI = ids.DOI
	id := ids.ID
	if len(id) > span.KeyLengthLimit {
		return output, span.Skip{Reason: fmt.Sprintf("id too long: %s", id)}
	}
	if len(id) == 0 {
		return nil, span.Skip{Reason: fmt.Sprintf("no doi or identifier: %v", article)}
	}
	output.ID = id
	output.RecordID = ids.DOI
	output.URL = append(output.URL, ids.URL)
	output.Authors = article.Authors()
	output.Format = Format
	output.Languages = article.Languages()
	output.MegaCollections = []string{SourceName}
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
		return output, span.Skip{Reason: fmt.Sprintf("zero date: %s", output.ID)}
	}
	// refs #5686
	for _, p := range ArticleTitleBlockPatterns {
		if p.MatchString(output.ArticleTitle) {
			return output, span.Skip{Reason: fmt.Sprintf("title blacklisted: %s", output.ArticleTitle)}
		}
	}
	// refs #5686, approx. article type distribution: https://git.io/vzlCr
	// switch article.Type {
	// case "book-review", "book-reviews", "Book Review":
	// 	output.ArticleTitle = fmt.Sprintf("Review: %s", article.ReviewedProduct())
	// case "misc", "other", "front-matter", "back-matter", "announcement", "font-matter", "fm", "fornt-matter":
	// 	return output, span.Skip{Reason: fmt.Sprintf("suppressed format: %s", article.Type)}
	// }
	return output, nil
}
