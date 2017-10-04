// Package doaj maps DOAJ metadata to intermediate schema.
//
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
package doaj

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/assetutil"
	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
)

const (
	SourceIdentifier = "28"
	Collection       = "DOAJ Directory of Open Access Journals"
	Format           = "ElectronicArticle"
	Genre            = "article" // Default ris.type
	DefaultRefType   = "EJOUR"
)

var (
	LCCPatterns = assetutil.MustLoadRegexpMap("assets/finc/lcc.json")
	LanguageMap = assetutil.MustLoadStringMap("assets/doaj/language-iso-639-3.json")
	DOIPattern  = regexp.MustCompile(`10[.][a-zA-Z0-9]+/[\S]*$`)
)

// Response from elasticsearch.
type Response struct {
	ID     string   `json:"_id"`
	Index  string   `json:"_index"`
	Source Document `json:"_source"`
	Type   string   `json:"_type"`
}

// Document metadata.
type Document struct {
	BibJSON BibJSON `json:"bibjson"`
	Created string  `json:"created_date"`
	ID      string  `json:"id"`
	Index   Index   `json:"index"`
	Updated string  `json:"last_updated"`
	Type    string  // make Response.Type available here
}

// Index metadata.
type Index struct {
	Classification []string `json:"classification"`
	Country        string   `json:"country"`
	Date           string   `json:"date"`
	ISSN           []string `json:"issn"`
	Language       []string `json:"language"`
	License        []string `json:"license"`
	SchemaCode     []string `json:"schema_code"`
	SchemaSubjects []string `json:"schema_subjects"`
	Subjects       []string `json:"subject"`
}

// BibJSON contains bibliographic data.
type BibJSON struct {
	Abstract string `json:"abstract"`
	Author   []struct {
		Name string `json:"name"`
	} `json:"author"`
	EndPage    string `json:"end_page"`
	Identifier []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"identifier"`
	Journal struct {
		Country  string   `json:"country"`
		Language []string `json:"language"`
		License  []struct {
			Title      string `json:"title"`
			Type       string `json:"type"`
			OpenAccess bool   `json:"open_access"`
			URL        string `json:"url"`
		} `json:"license"`
		Number    string `json:"number"`
		Publisher string `json:"publisher"`
		Title     string `json:"title"`
		Volume    string `json:"volume"`
	} `json:"journal"`
	Link []struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"link"`
	Month     string `json:"month"`
	StartPage string `json:"start_page"`
	Subject   []struct {
		Code   string `json:"code"`
		Scheme string `json:"scheme"`
		Term   string `json:"term"`
	} `json:"subject"`
	Title string `json:"title"`
	Year  string `json:"year"`
}

// ToIntermediateSchema returns an intermediate schema from a DOAJ elasticsearch response.
func (resp *Response) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	resp.Source.Type = resp.Type
	return resp.Source.ToIntermediateSchema()

}

// Authors returns a list of authors.
func (doc Document) Authors() (authors []finc.Author) {
	for _, author := range doc.BibJSON.Author {
		authors = append(authors, finc.Author{Name: html.UnescapeString(author.Name)})
	}
	return authors
}

// Date return the document date. Journals entries usually have no date, so
// they will err.
func (doc Document) Date() (time.Time, error) {
	if doc.Index.Date != "" {
		return time.Parse("2006-01-02T15:04:05Z", doc.Index.Date)
	}
	var s string
	if y, err := strconv.Atoi(doc.BibJSON.Year); err == nil {
		s = fmt.Sprintf("%04d-01-01", y)
		if m, err := strconv.Atoi(doc.BibJSON.Month); err == nil {
			if m > 0 && m < 13 {
				s = fmt.Sprintf("%04d-%02d-01", y, m)
			}
		}
	}
	return time.Parse("2006-01-02", s)
}

// DOI returns the DOI or the empty string.
func (doc Document) DOI() string {
	for _, identifier := range doc.BibJSON.Identifier {
		if identifier.Type == "doi" {
			id := strings.TrimSpace(identifier.ID)
			if !strings.Contains(id, "http") {
				return id
			}
			return DOIPattern.FindString(id)
		}
	}
	return ""
}

// ToIntermediateSchema converts a doaj document to intermediate schema. For
// now any record, that has no usable date will be skipped.
func (doc Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error

	output := finc.NewIntermediateSchema()
	output.Date, err = doc.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	output.RawDate = output.Date.Format("2006-01-02")

	id := fmt.Sprintf("ai-%s-%s", SourceIdentifier, doc.ID)
	if len(id) > span.KeyLengthLimit {
		return output, span.Skip{Reason: fmt.Sprintf("id too long: %s", id)}
	}

	output.ArticleTitle = doc.BibJSON.Title
	output.Authors = doc.Authors()
	output.DOI = doc.DOI()
	output.Format = Format
	output.Genre = Genre
	output.ISSN = doc.Index.ISSN
	output.JournalTitle = doc.BibJSON.Journal.Title
	output.MegaCollections = []string{Collection}

	publisher := strings.TrimSpace(doc.BibJSON.Journal.Publisher)
	if publisher != "" {
		output.Publishers = append(output.Publishers, publisher)
	}

	output.RecordID = id
	output.SourceID = SourceIdentifier
	output.Volume = doc.BibJSON.Journal.Volume

	// refs. #8709
	if output.DOI != "" {
		output.URL = append(output.URL, "http://doi.org/"+output.DOI)
	}

	// refs. #8709
	if len(output.URL) == 0 {
		for _, link := range doc.BibJSON.Link {
			output.URL = append(output.URL, link.URL)
		}
	}

	// refs. #6634
	if len(output.URL) == 0 {
		output.URL = append(output.URL, "https://doaj.org/article/"+doc.ID)
	}

	output.StartPage = doc.BibJSON.StartPage
	output.EndPage = doc.BibJSON.EndPage

	if sp, err := strconv.Atoi(doc.BibJSON.StartPage); err == nil {
		if ep, err := strconv.Atoi(doc.BibJSON.EndPage); err == nil {
			output.PageCount = fmt.Sprintf("%d", ep-sp)
			output.Pages = fmt.Sprintf("%d-%d", sp, ep)
		}
	}

	subjects := container.NewStringSet()
	for _, s := range doc.Index.SchemaCode {
		class := LCCPatterns.LookupDefault(strings.Replace(s, "LCC:", "", -1), finc.NotAssigned)
		if class != finc.NotAssigned {
			subjects.Add(class)
		}
	}
	if subjects.Size() == 0 {
		output.Subjects = []string{finc.NotAssigned}
	} else {
		output.Subjects = subjects.SortedValues()
	}

	languages := container.NewStringSet()
	for _, l := range doc.Index.Language {
		languages.Add(LanguageMap.LookupDefault(l, "und"))
	}
	output.Languages = languages.Values()

	output.RefType = DefaultRefType
	return output, nil
}
