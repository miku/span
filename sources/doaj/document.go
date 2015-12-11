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
// Directory of open access journals.
package doaj

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/assetutil"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
)

const (
	// Internal bookkeeping.
	SourceID = "28"
	// Collection name
	Collection = "DOAJ Directory of Open Access Journals"
	// Format for all records
	Format = "ElectronicArticle"
	Genre  = "article"
)

var errDateMissing = errors.New("date is missing")

var (
	LCCPatterns = assetutil.MustLoadRegexpMap("assets/finc/lcc.json")
	LanguageMap = assetutil.MustLoadStringMap("assets/doaj/language-iso-639-3.json")
)

type Response struct {
	ID     string   `json:"_id"`
	Index  string   `json:"_index"`
	Source Document `json:"_source"`
	Type   string   `json:"_type"`
}

type Document struct {
	BibJson BibJSON `json:"bibjson"`
	Created string  `json:"created_date"`
	ID      string  `json:"id"`
	Index   Index   `json:"index"`
	Updated string  `json:"last_updated"`
	// make Response.Type available here
	Type string
}

type Index struct {
	Classification []string `json:"classification"`
	Country        string   `json:"country"`
	Date           string   `json:"date"`
	ISSN           []string `json:"issn"`
	Language       []string `json:"language"`
	License        []string `json:"license"`
	Publishers     []string `json:"publisher"`
	SchemaCode     []string `json:"schema_code"`
	SchemaSubjects []string `json:"schema_subjects"`
	Subjects       []string `json:"subject"`
}

type License struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}

type Journal struct {
	Country   string    `json:"country"`
	Language  []string  `json:"language"`
	License   []License `json:"license"`
	Number    string    `json:"number"`
	Publisher string    `json:"publisher"`
	Title     string    `json:"title"`
	Volume    string    `json:"volume"`
}

type Author struct {
	Name string `json:"name"`
}

type Subject struct {
	Code   string `json:"code"`
	Scheme string `json:"scheme"`
	Term   string `json:"term"`
}

type Link struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Identifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type BibJSON struct {
	Abstract   string       `json:"abstract"`
	Author     []Author     `json:"author"`
	EndPage    string       `json:"end_page"`
	Identifier []Identifier `json:"identifier"`
	Journal    Journal      `json:"journal"`
	Link       []Link       `json:"link"`
	Month      string       `json:"month"`
	StartPage  string       `json:"start_page"`
	Subject    []Subject    `json:"subject"`
	Title      string       `json:"title"`
	Year       string       `json:"year"`
}

type DOAJ struct{}

func (s DOAJ) Iterate(r io.Reader) (<-chan []span.Importer, error) {
	return span.FromLines(r, func(b []byte) (span.Importer, error) {
		resp := new(Response)
		err := json.Unmarshal(b, resp)
		if err != nil {
			return resp.Source, err
		}
		resp.Source.Type = resp.Type
		return resp.Source, nil
	})
}

func (doc Document) Authors() (authors []finc.Author) {
	for _, author := range doc.BibJson.Author {
		authors = append(authors, finc.Author{Name: author.Name})
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
	if y, err := strconv.Atoi(doc.BibJson.Year); err == nil {
		s = fmt.Sprintf("%04d-01-01", y)
		if m, err := strconv.Atoi(doc.BibJson.Month); err == nil {
			if m > 0 && m < 13 {
				s = fmt.Sprintf("%04d-%02d-01", y, m)
			}
		}
	}
	return time.Parse("2006-01-02", s)
}

func (doc Document) DOI() string {
	for _, identifier := range doc.BibJson.Identifier {
		if identifier.Type == "doi" {
			return identifier.ID
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

	id := fmt.Sprintf("ai-%s-%s", SourceID, doc.ID)
	if len(id) > span.KeyLengthLimit {
		return output, span.Skip{Reason: fmt.Sprintf("id too long: %s", id)}
	}

	output.ArticleTitle = doc.BibJson.Title
	output.Authors = doc.Authors()
	output.DOI = doc.DOI()
	output.Format = Format
	output.Genre = Genre
	output.ISSN = doc.Index.ISSN
	output.JournalTitle = doc.BibJson.Journal.Title
	output.MegaCollection = Collection

	publisher := strings.TrimSpace(doc.BibJson.Journal.Publisher)
	if publisher != "" {
		output.Publishers = append(output.Publishers, publisher)
	}

	output.RecordID = id
	output.SourceID = SourceID
	output.Volume = doc.BibJson.Journal.Volume

	for _, link := range doc.BibJson.Link {
		output.URL = append(output.URL, link.URL)
	}

	// refs. #6634
	if len(output.URL) == 0 {
		output.URL = append(output.URL, "https://doaj.org/article/"+doc.ID)
	}

	output.StartPage = doc.BibJson.StartPage
	output.EndPage = doc.BibJson.EndPage

	if sp, err := strconv.Atoi(doc.BibJson.StartPage); err == nil {
		if ep, err := strconv.Atoi(doc.BibJson.EndPage); err == nil {
			output.PageCount = fmt.Sprintf("%d", ep-sp)
			output.Pages = fmt.Sprintf("%d-%d", sp, ep)
		}
	}

	subjects := container.NewStringSet()
	for _, s := range doc.Index.SchemaCode {
		class := LCCPatterns.LookupDefault(strings.Replace(s, "LCC:", "", -1), finc.NOT_ASSIGNED)
		if class != finc.NOT_ASSIGNED {
			subjects.Add(class)
		}
	}
	if subjects.Size() == 0 {
		output.Subjects = []string{finc.NOT_ASSIGNED}
	} else {
		output.Subjects = subjects.SortedValues()
	}

	languages := container.NewStringSet()
	for _, l := range doc.Index.Language {
		languages.Add(LanguageMap.LookupDefault(l, "und"))
	}
	output.Languages = languages.Values()

	return output, nil
}
