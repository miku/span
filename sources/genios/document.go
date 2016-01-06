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
package genios

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/assetutil"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
)

const (
	SourceID = "48"

	Format = "ElectronicArticle"
	// Collection is the base name of the collection.
	Collection = "Genios"
	Genre      = "article"
	// If no abstract is found accept this number of chars from doc.Text as Abstract.
	textAsAbstractCutoff = 2000
)

type Document struct {
	ID               string   `xml:"ID,attr"`
	DB               string   `xml:"DB,attr"`
	IDNAME           string   `xml:"IDNAME,attr"`
	ISSN             string   `xml:"ISSN"`
	Source           string   `xml:"Source"`
	PublicationTitle string   `xml:"Publication-Title"`
	Title            string   `xml:"Title"`
	Year             string   `xml:"Year"`
	RawDate          string   `xml:"Date"`
	Volume           string   `xml:"Volume"`
	Issue            string   `xml:"Issue"`
	RawAuthors       []string `xml:"Authors>Author"`
	Language         string   `xml:"Language"`
	Abstract         string   `xml:"Abstract"`
	Descriptors      string   `xml:"Descriptors>Descriptor"`
	Text             string   `xml:"Text"`
	XGroup           string   `xml:"x-group"`
	XIssue           string   `xml:"x-issue"`
}

var (
	rawDateReplacer = strings.NewReplacer(`"`, "", "\n", "", "\t", "")
	collections     = assetutil.MustLoadStringMap("assets/genios/collections.json")
	// Restricts the possible languages for detection.
	acceptedLanguages = container.NewStringSet("deu", "eng")
)

type Genios struct{}

// Iterate emits Converter elements via XML decoding.
func (s Genios) Iterate(r io.Reader) (<-chan []span.Importer, error) {
	return span.FromXML(r, "Document", func(d *xml.Decoder, se xml.StartElement) (span.Importer, error) {
		doc := new(Document)
		err := d.DecodeElement(&doc, &se)
		return doc, err
	})
}

// Headings returns subject headings.
func (doc Document) Headings() []string {
	var headings []string
	fields := strings.FieldsFunc(doc.Descriptors, func(r rune) bool {
		return r == ';' || r == '/'
	})
	for _, f := range fields {
		headings = append(headings, strings.TrimSpace(f))
	}
	return headings
}

// Date returns the date as noted in the document.
func (doc Document) Date() (time.Time, error) {
	raw := strings.TrimSpace(rawDateReplacer.Replace(doc.RawDate))
	if len(raw) > 8 {
		raw = raw[:8]
	}
	return time.Parse("20060102", raw)
}

// SourceAndID will probably be a unique identifier. An ID alone might not be enough.
func (doc Document) SourceAndID() string {
	return fmt.Sprintf("%s__%s", strings.TrimSpace(doc.Source), strings.TrimSpace(doc.ID))
}

// URL returns a constructed URL at the publishers site.
func (doc Document) URL() string {
	return fmt.Sprintf("https://www.wiso-net.de/document/%s", doc.SourceAndID())
}

// isNomenNescio returns true, if the field is de-facto empty.
func isNomenNescio(s string) bool {
	t := strings.ToLower(strings.TrimSpace(s))
	return t == "n.n." || t == ""
}

// Authors returns a list of authors. Formatting is not cleaned up, so you'll
// get any combination of surname and given names.
func (doc Document) Authors() (authors []finc.Author) {
	for _, s := range doc.RawAuthors {
		fields := strings.FieldsFunc(s, func(r rune) bool {
			return r == ';' || r == '/'
		})
		for _, f := range fields {
			if isNomenNescio(f) {
				continue
			}
			authors = append(authors, finc.Author{Name: strings.TrimSpace(f)})
		}
	}
	return authors
}

// RecordID uses SourceAndID as starting point.
func (doc Document) RecordID() string {
	return fmt.Sprintf("ai-%s-%s", SourceID, base64.RawURLEncoding.EncodeToString([]byte(doc.SourceAndID())))
}

// Languages returns the given and guessed languages found in abstract and
// fulltext. Note: This is slow. Skip detection on too short strings.
func (doc Document) Languages() []string {
	set := container.NewStringSet()

	vals := []string{doc.Title, doc.Text}

	for _, s := range vals {
		if len(s) < 20 {
			continue
		}
		lang, err := span.DetectLang3(s)
		if err != nil {
			continue
		}
		if !acceptedLanguages.Contains(lang) {
			continue
		}
		if lang == "und" {
			continue
		}
		set.Add(lang)
	}

	return set.Values()
}

// ToIntermediateSchema converts a genios document into an intermediate schema document.
// Will fail/skip records with unusable dates.
func (doc Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error
	output := finc.NewIntermediateSchema()

	output.Date, err = doc.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	output.RawDate = output.Date.Format("2006-01-02")

	output.Authors = doc.Authors()

	output.URL = append(output.URL, doc.URL())

	if isNomenNescio(doc.Abstract) {
		cutoff := len(doc.Text)
		if cutoff > textAsAbstractCutoff {
			cutoff = textAsAbstractCutoff
		}
		output.Abstract = strings.TrimSpace(doc.Text[:cutoff])
	} else {
		output.Abstract = strings.TrimSpace(doc.Abstract)

	}

	output.ArticleTitle = strings.TrimSpace(doc.Title)
	output.JournalTitle = strings.TrimSpace(doc.PublicationTitle)

	if !isNomenNescio(doc.ISSN) {
		output.ISSN = append(output.ISSN, strings.TrimSpace(doc.ISSN))
	}

	if !isNomenNescio(doc.Issue) {
		output.Issue = strings.TrimSpace(doc.Issue)
	}

	if !isNomenNescio(doc.Volume) {
		output.Volume = strings.TrimSpace(doc.Volume)
	}

	output.Fulltext = doc.Text
	output.Format = Format
	output.Genre = Genre
	output.Languages = doc.Languages()
	output.Package = doc.DB
	output.MegaCollection = fmt.Sprintf("Genios (%s)", collections[doc.XGroup])
	id := doc.RecordID()
	// 250 is a limit on memcached keys; offending key was:
	// ai-48-R1JFUl9fU2NoZWliIEVsZWt0cm90ZWNobmlrIEdtYkggwr\
	// dTdGV1ZXJ1bmdzYmF1IMK3SW5kdXN0cmllLUVsZWt0cm9uaWsgwr\
	// dFbGVrdHJvbWFzY2hpbmVuYmF1IMK3SW5kdXN0cmllLVNlcnZpY2\
	// UgwrdEYW5mb3NzLVN5c3RlbXBhcnRuZXIgwrdEYW5mb3NzIERyaX\
	// ZlcyBDZW50ZXIgwrdNYXJ0aW4gU2ljaGVyaGVpdHN0ZWNobmlr
	if len(id) > span.KeyLengthLimit {
		return output, span.Skip{Reason: fmt.Sprintf("id too long: %s", id)}
	}
	output.RecordID = id
	output.SourceID = SourceID
	output.Subjects = doc.Headings()

	// keep the date indicator, so we can create an update order
	output.Indicator = doc.XIssue

	return output, nil
}
