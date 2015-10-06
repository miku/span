//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                 by The Finc Authors, http://finc.info
//                 by Martin Czygan, <martin.czygan@uni-leipzig.de>
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
package thieme

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

const (
	SourceID = "60"
	// Format   = "ElectronicArticle"
	Collection = "Thieme"
	Genre      = "article"
)

func leftPad(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

type Thieme struct{}

// Iterate emits Converter elements via XML decoding.
func (s Thieme) Iterate(r io.Reader) (<-chan []span.Importer, error) {
	return span.FromXML(r, "record", func(d *xml.Decoder, se xml.StartElement) (span.Importer, error) {
		doc := new(Document)
		err := d.DecodeElement(&doc, &se)
		return doc, err
	})
}

type Document struct {
	Identifier struct {
		Value string `xml:",chardata"`
	} `xml:"header>identifier"`
	Metadata struct {
		ArticleSet struct {
			Article struct {
				VernacularTitle struct {
					VernacularLanguage struct {
						Abstract struct {
							Value string `xml:",chardata"`
						} `xml:"abstract"`
						AuthorList struct {
							Author []struct {
								FirstName struct {
								} `xml:"firstname"`
								LastName struct {
								} `xml:"lastname"`
							} `xml:"author"`
						} `xml:"authorlist"`
					} `xml:"vernacularlanguage"`
				} `xml:"vernaculartitle"`
				Title struct {
					Value string `xml:",chardata"`
				} `xml:"articletitle"`
				Journal struct {
					Publisher struct {
						Name string `xml:",chardata"`
					} `xml:"publishername"`
					Title struct {
						Value string `xml:",chardata"`
					} `xml:"journaltitle"`
					ISSN struct {
						Value string `xml:",chardata"`
					} `xml:"issn"`
					EISSN struct {
						Value string `xml:",chardata"`
					} `xml:"e-issn"`
					Volume struct {
						Value string `xml:",chardata"`
					} `xml:"volume"`
					Issue struct {
						Value string `xml:",chardata"`
					} `xml:"issue"`
					PubDate struct {
						Year struct {
							Value string `xml:",chardata"`
						} `xml:"year"`
						Month struct {
							Value string `xml:",chardata"`
						} `xml:"month"`
						Day struct {
							Value string `xml:",chardata"`
						} `xml:"day"`
					} `xml:"pubdate"`
				} `xml:"journal"`
			} `xml:"article"`
		} `xml:"articleset"`
	} `xml:"metadata"`
}

func (doc Document) Date() (time.Time, error) {
	pd := doc.Metadata.ArticleSet.Article.Journal.PubDate

	if pd.Month.Value == "0" {
		pd.Month.Value = "01"
	}
	if pd.Day.Value == "0" {
		pd.Day.Value = "01"
	}

	if pd.Year.Value != "" && pd.Month.Value != "" && pd.Day.Value != "" {
		s := fmt.Sprintf("%s-%s-%s", leftPad(pd.Year.Value, "0", 4),
			leftPad(pd.Month.Value, "0", 2), leftPad(pd.Day.Value, "0", 2))
		return time.Parse("2006-01-02", s)
	}
	if pd.Year.Value != "" && pd.Month.Value != "" {
		s := fmt.Sprintf("%s-%s-01", leftPad(pd.Year.Value, "0", 4),
			leftPad(pd.Month.Value, "0", 2))
		return time.Parse("2006-01-02", s)
	}
	if pd.Year.Value != "" {
		s := fmt.Sprintf("%s-01-01", leftPad(pd.Year.Value, "0", 4))
		return time.Parse("2006-01-02", s)
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func (doc Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	output.RecordID = doc.Identifier.Value
	output.SourceID = SourceID
	output.MegaCollection = Collection
	output.Genre = Genre

	date, err := doc.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	output.Date = date

	journal := doc.Metadata.ArticleSet.Article.Journal
	if journal.Publisher.Name != "" {
		output.Publishers = append(output.Publishers, journal.Publisher.Name)
	}

	if journal.Title.Value == "" {
		return output, span.Skip{Reason: fmt.Sprintf("NO_JTITLE %s", output.RecordID)}
	}

	output.JournalTitle = journal.Title.Value
	output.ISSN = append(output.ISSN, journal.ISSN.Value)
	output.EISSN = append(output.EISSN, journal.ISSN.Value)
	output.Volume = journal.Volume.Value
	output.Issue = journal.Issue.Value

	article := doc.Metadata.ArticleSet.Article
	if article.Title.Value == "" {
		return output, span.Skip{Reason: fmt.Sprintf("NO_ATITLE %s", output.RecordID)}
	}

	output.ArticleTitle = article.Title.Value
	output.Abstract = article.VernacularTitle.VernacularLanguage.Abstract.Value

	return output, nil
}
