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
	SourceID   = "60"
	Format     = "ElectronicArticle"
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
	Identifier string `xml:"header>identifier"`
	Metadata   struct {
		ArticleSet struct {
			Article struct {
				Journal struct {
					PublisherName string `xml:"PublisherName"`
					JournalTitle  string `xml:"JournalTitle"`
					ISSN          string `xml:"Issn"`
					EISSN         string `xml:"E-Issn"`
					Volume        string `xml:"Volume"`
					Issue         string `xml:"Issue"`
					PubDate       struct {
						Year  string `xml:"Year"`
						Month string `xml:"Month"`
						Day   string `xml:"Day"`
					} `xml:"PubDate"`
				} `xml:"Journal"`
				AuthorList struct {
					Author []struct {
						FirstName string `xml:"FirstName"`
						LastName  string `xml:"LastName"`
					} `xml:"Author"`
				} `xml:"AuthorList"`
				ArticleIdList []struct {
					ArticleId struct {
						OpenAccess string `xml:"OpenAccess,attr"`
						IdType     string `xml:"IdType,attr"`
						Id         string `xml:",chardata"`
					}
				} `xml:"ArticleIdList"`
				ArticleType        string   `xml:"ArticleType"`
				ArticleTitle       string   `xml:"ArticleTitle"`
				VernacularTitle    string   `xml:"VernacularTitle"`
				FirstPage          string   `xml:"FirstPage"`
				LastPage           string   `xml:"LastPage"`
				VernacularLanguage string   `xml:"VernacularLanguage"`
				Language           string   `xml:"Language"`
				Subject            []string `xml:"subject"`
				Links              []string `xml:"Links>Link"`
				History            []struct {
					PubDate struct {
						Status string `xml:"PubStatus,attr"`
						Year   string `xml:"Year"`
						Month  string `xml:"Month"`
						Day    string `xml:"Day"`
					} `xml:"PubDate"`
				} `xml:"History"`
				Abstract           string `xml:"Abstract"`
				VernacularAbstract string `xml:"VernacularAbstract"`
				Format             struct {
					HTML string `xml:"html,attr"`
					PDF  string `xml:"pdf,attr"`
				} `xml:"format"`
				CopyrightInformation string `xml:"CopyrightInformation"`
			} `xml:"Article"`
		} `xml:"ArticleSet"`
	} `xml:"metadata"`
}

func (doc Document) Date() (time.Time, error) {
	pd := doc.Metadata.ArticleSet.Article.Journal.PubDate

	if pd.Month == "0" {
		pd.Month = "01"
	}
	if pd.Day == "0" {
		pd.Day = "01"
	}

	if pd.Year != "" && pd.Month != "" && pd.Day != "" {
		s := fmt.Sprintf("%s-%s-%s", leftPad(pd.Year, "0", 4),
			leftPad(pd.Month, "0", 2), leftPad(pd.Day, "0", 2))
		return time.Parse("2006-01-02", s)
	}
	if pd.Year != "" && pd.Month != "" {
		s := fmt.Sprintf("%s-%s-01", leftPad(pd.Year, "0", 4),
			leftPad(pd.Month, "0", 2))
		return time.Parse("2006-01-02", s)
	}
	if pd.Year != "" {
		s := fmt.Sprintf("%s-01-01", leftPad(pd.Year, "0", 4))
		return time.Parse("2006-01-02", s)
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func (doc Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	output.RecordID = doc.Identifier
	output.SourceID = SourceID
	output.MegaCollection = Collection
	output.Genre = Genre

	date, err := doc.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	output.Date = date

	journal := doc.Metadata.ArticleSet.Article.Journal
	if journal.PublisherName != "" {
		output.Publishers = append(output.Publishers, journal.PublisherName)
	}

	if journal.JournalTitle == "" {
		return output, span.Skip{Reason: fmt.Sprintf("NO_JTITLE %s", output.RecordID)}
	}

	output.JournalTitle = journal.JournalTitle
	output.ISSN = append(output.ISSN, journal.ISSN)
	output.EISSN = append(output.EISSN, journal.EISSN)
	output.Volume = journal.Volume
	output.Issue = journal.Issue

	article := doc.Metadata.ArticleSet.Article

	output.Abstract = article.Abstract
	if output.Abstract == "" {
		output.Abstract = article.VernacularAbstract
	}

	for _, link := range article.Links {
		output.URL = append(output.URL, link)
	}

	output.Subjects = article.Subject

	if article.Language != "" {
		output.Languages = append(output.Languages, strings.ToLower(article.Language))
	} else {
		if article.VernacularLanguage != "" {
			output.Languages = append(output.Languages, strings.ToLower(article.VernacularLanguage))
		}
	}

	return output, nil
}
