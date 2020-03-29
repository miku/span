package degruyter

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

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/formats/jats"
)

const (
	// SourceID for internal bookkeeping.
	SourceID = "50"
	// SourceName for finc.mega_collection.
	SourceName            = "De Gruyter Journals / Social Sciences and Humanities"
	TechnicalCollectionID = "sid-50-col-degruyterssh"
	// Format for intermediate schema.
	Format = "ElectronicArticle"
)

// Article with extras for this source.
type Article struct {
	XMLName xml.Name `xml:"article"`
	jats.Article
}

// Identifiers returns the doi and the dependent url and recordID in a struct.
// It is an error, if there is no DOI.
func (article *Article) Identifiers() (jats.Identifiers, error) {
	var ids jats.Identifiers
	doi, err := article.DOI()
	if err != nil {
		return ids, err
	}
	locator := fmt.Sprintf("http://dx.doi.org/%s", doi)
	id := fmt.Sprintf("ai-%s-%s", SourceID, base64.RawURLEncoding.EncodeToString([]byte(locator)))
	return jats.Identifiers{DOI: doi, URL: locator, ID: id}, nil
}

// ToIntermediateSchema converts a jats article into an internal schema.
func (article *Article) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output, err := article.Article.ToIntermediateSchema()
	if err != nil {
		return output, err
	}

	ids, err := article.Identifiers()
	if err != nil {
		return output, err
	}

	id := ids.ID
	if len(id) > span.KeyLengthLimit {
		return output, span.Skip{Reason: fmt.Sprintf("id too long: %s", id)}
	}
	output.ID = id
	output.RecordID = ids.DOI

	output.DOI = ids.DOI
	output.URL = append(output.URL, ids.URL)

	output.Format = Format
	output.MegaCollections = []string{SourceName, TechnicalCollectionID}
	output.SourceID = SourceID

	return output, nil
}
