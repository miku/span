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
	"regexp"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/miku/span"
	"github.com/miku/span/assetutil"
	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
)

const (
	// SourceID for internal bookkeeping.
	SourceID = "48"
	// Format is mapped per site later.
	Format = "ElectronicArticle"
	// Collection is the base name of the collection.
	Collection = "Genios"
	// Genre default.
	Genre = "article"
	// DefaultRefType is the default ris.type.
	DefaultRefType = "EJOUR"
	// If no abstract is found accept this number of chars from doc.Text as Abstract.
	textAsAbstractCutoff = 200
	// maxAuthorLength example: document/BOND__b0604160052
	maxAuthorLength = 200
	minAuthorLength = 4
	maxTitleLength  = 2048
)

// Document was generated 2019-06-12 19:13:21 by tir on hayiti, adjusted.
type Document struct {
	XMLName     xml.Name `xml:"Document"`
	Chardata    string   `xml:",chardata"`
	ID          string   `xml:"ID,attr"`
	IDNAME      string   `xml:"IDNAME,attr"`
	DB          string   `xml:"DB,attr"`
	Abstract    string   `xml:"Abstract"` // This interview deals with...
	RawAuthors  []string `xml:"Authors>Author"`
	Descriptors string   `xml:"Descriptors>Descriptor"`
	RawDate     string   `xml:"Date"`         // 20050101, 20050501
	Issue       string   `xml:"Issue"`        // 1, 2
	ISSN        string   `xml:"ISSN"`         // 1861-1303
	ISBN        string   `xml:"ISBN"`         // n.n.
	Subtitle    string   `xml:"Subtitle"`     // n.n.
	SeriesTitle string   `xml:"Series-Title"` // n.n.
	Editors     struct {
		Text   string `xml:",chardata"`
		Editor string `xml:"Editor"` // n.n.
	} `xml:"Editors"`
	Edition          string   `xml:"Edition"`           // n.n.
	Language         string   `xml:"Language"`          // n.n.
	Page             string   `xml:"Page"`              // 9, 43, 69, 87, 99, 121, 1...
	PublicationTitle string   `xml:"Publication-Title"` // International Journal of ...
	Source           string   `xml:"Source"`            // IJAR
	Title            string   `xml:"Title"`             // "One sows the seed, but i...
	Text             string   `xml:"Text"`
	Volume           string   `xml:"Volume"`          // 1
	Year             string   `xml:"Year"`            // 2005
	Affiliation      string   `xml:"Affiliation"`     // n.n.
	DOI              string   `xml:"DOI"`             // n.n.
	PersistentLink   string   `xml:"Persistent_Link"` // https://www.wiso-net.de/d...
	Publisher        string   `xml:"Publisher"`       // Rainer Hampp Verlag
	Available        string   `xml:"available"`       // n.n.
	Copyright        string   `xml:"Copyright"`
	Modules          []string `xml:"Modules>Module"`
	XPackage         string   `xml:"X-Package"` // fachzeitschriften
}

var (
	rawDateReplacer = strings.NewReplacer(`"`, "", "\n", "", "\t", "")
	// acceptedLanguages restricts the possible languages for detection.
	acceptedLanguages = container.NewStringSet("deu", "eng")
	// dbmap maps a database name to one or more "package names"
	dbmap = assetutil.MustLoadStringSliceMap("assets/genios/dbmap.json")
	// yearPattern matches YYYY
	yearPattern = regexp.MustCompile(`[12][0-9][0-9][0-9]`)
)

// Headings returns subject headings.
func (doc Document) Headings() []string {
	var headings []string
	fields := strings.FieldsFunc(doc.Descriptors, func(r rune) bool {
		return r == ';' || r == '/'
	})
	// refs. #8009
	if len(fields) == 1 {
		fields = strings.FieldsFunc(doc.Descriptors, func(r rune) bool {
			return r == ','
		})
	}
	for _, f := range fields {
		headings = append(headings, strings.TrimSpace(f))
	}
	return headings
}

// Date returns the date as noted in the document. There might be two values:
// Date and Year. Defaults to Year, fallback to Date, refs #12193.
func (doc Document) Date() (time.Time, error) {
	rawYear := strings.TrimSpace(rawDateReplacer.Replace(doc.Year))
	if yearPattern.MatchString(rawYear) {
		// Prefer Year, refs #12193.
		return time.Parse("2006", rawYear)
	}
	// Fallback to Date, refs #12193.
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

// URL returns a constructed URL at the publishers site, refs #15177.
func (doc Document) URL() string {
	if link := strings.TrimSpace(doc.PersistentLink); link != "" {
		return link
	}
	log.Printf("%s has no persistent link, falling back", doc.FincID())
	return fmt.Sprintf("https://www.wiso-net.de/document/%s", doc.SourceAndID())
}

// isNomenNescio returns true, if the field is de-facto empty.
func isNomenNescio(s string) bool {
	t := strings.ToLower(strings.TrimSpace(s))
	return t == "n.n." || t == ""
}

// stringContainsAny returns true, if string contains any of the strings given.
func stringContainsAny(s string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

// Authors returns a list of authors. Formatting is not cleaned up, so you'll
// get any combination of surname and given names.
func (doc Document) Authors() (authors []finc.Author) {
	for _, s := range doc.RawAuthors {
		fields := strings.FieldsFunc(s, func(r rune) bool {
			// If a single field value exceeds a threshold, try
			// more delimiters, refs #12218.
			if len(s) > 60 {
				return r == ';' || r == '/' || r == ','
			}
			return r == ';' || r == '/'
		})
		for _, f := range fields {
			if isNomenNescio(f) {
				continue
			}
			name := strings.TrimSpace(f)
			// Author field sometime contains things like, &quot,
			// www.website.com, N.Y., and more weird things, skip these cases.
			if len(name) < minAuthorLength {
				continue
			}
			// Author substrings to filter out, this is just the tip of the iceberg.
			clues := []string{"www.", "http:", "&quot", "part 1 of", "part 2 of",
				"Copyright", "(c)", "All rights reserved", "he said"}
			if stringContainsAny(name, clues) {
				continue
			}
			if len(name) < maxAuthorLength {
				authors = append(authors, finc.Author{Name: name})
			}
		}
	}
	return authors
}

// ISSNList returns a list of ISSN.
func (doc Document) ISSNList() []string {
	issns := container.NewStringSet()
	for _, s := range span.ISSNPattern.FindAllString(doc.ISSN, -1) {
		issns.Add(s)
	}
	return issns.Values()
}

// FincID uses SourceAndID as starting point.
func (doc Document) FincID() string {
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
	if len(output.ArticleTitle) > maxTitleLength {
		return output, span.Skip{Reason: fmt.Sprintf("article title too long: %d", len(output.ArticleTitle))}
	}

	// TODO(miku): Find DB names where this is relevant.
	output.JournalTitle = strings.Replace(strings.TrimSpace(doc.PublicationTitle), "\n", " ", -1)

	output.ISSN = doc.ISSNList()

	if !isNomenNescio(doc.Issue) {
		output.Issue = strings.TrimSpace(doc.Issue)
	}

	if !isNomenNescio(doc.Volume) {
		output.Volume = strings.TrimSpace(doc.Volume)
	}

	output.Fulltext = "" // refs. #16743
	output.Format = Format
	output.Genre = Genre
	output.Languages = doc.Languages()

	// XXX: Package names changed around 11/2019, e.g. "Wiso Journals /
	// Wirtschaftswissenschaften", "Wiso Journals", "Wiso Journals / Recht",
	// ...

	// Lookup names like AAA, BUBH, CMW, ... in https://git.io/v2ECx - a single
	// DB may be in multiple packages.
	var packageNames = dbmap.LookupDefault(doc.DB, []string{})

	// Wrap package name in 'Genios (PACKAGENAME)'
	var packageNameVariants []string
	for _, name := range packageNames {
		packageNameVariants = append(packageNameVariants, fmt.Sprintf("Genios (%s)", name))
		packageNameVariants = append(packageNameVariants, fmt.Sprintf("Wiso Journals / %s", name))
		// Add TCID, TODO(miku): sid-48-col-wisoubl
		switch name {
		case "Fachzeitschriften":
			packageNameVariants = append(packageNameVariants, "sid-48-col-wisofachzs")
		case "Wirtschaftswissenschaften":
			packageNameVariants = append(packageNameVariants, "sid-48-col-wisowirtschaft")
		case "Sozialwissenschaften":
			packageNameVariants = append(packageNameVariants, "sid-48-col-wisosozial")
		case "Psychologie":
			packageNameVariants = append(packageNameVariants, "sid-48-col-wisopsych")
		case "Recht":
			packageNameVariants = append(packageNameVariants, "sid-48-col-wisorecht")
		}
	}

	// hack, to move Genios (LIT) further down
	sort.Sort(sort.Reverse(sort.StringSlice(packageNameVariants)))

	// Note DB name as well as package name (Wiwi, Sowi, Recht, etc.) as well
	// as kind, which - a bit confusingly - is also package in licensing terms (FZS).
	output.Packages = append([]string{doc.DB}, packageNameVariants...)

	// 2018-06-01, Modules are added, (1) add them in addition to existing
	// package names, later XXX: (2) remove own tags.
	output.Packages = append(output.Packages, doc.Modules...)

	if len(packageNameVariants) > 0 {
		output.MegaCollections = []string{packageNameVariants[0]}
	} else {
		// XXX: Log these somewhere.
		log.Printf("genios: db is not associated with package: %s, using generic default", doc.DB)
		output.MegaCollections = []string{fmt.Sprintf("Genios")}
	}

	id := doc.FincID()
	// 250 was a limit on memcached keys; offending key was:
	// ai-48-R1JFUl9fU2NoZWliIEVsZWt0cm90ZWNobmlrIEdtYkggwr\
	// dTdGV1ZXJ1bmdzYmF1IMK3SW5kdXN0cmllLUVsZWt0cm9uaWsgwr\
	// dFbGVrdHJvbWFzY2hpbmVuYmF1IMK3SW5kdXN0cmllLVNlcnZpY2\
	// UgwrdEYW5mb3NzLVN5c3RlbXBhcnRuZXIgwrdEYW5mb3NzIERyaX\
	// ZlcyBDZW50ZXIgwrdNYXJ0aW4gU2ljaGVyaGVpdHN0ZWNobmlr
	if len(id) > span.KeyLengthLimit {
		return output, span.Skip{Reason: fmt.Sprintf("id too long: %s", id)}
	}
	output.ID = id
	output.RecordID = doc.ID
	output.SourceID = SourceID
	output.Subjects = doc.Headings()

	output.RefType = DefaultRefType
	output.OpenAccess = strings.HasPrefix(doc.Available, "Open Access") // refs #15008

	return output, nil
}
