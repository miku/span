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
package crossref

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/assetutil"
	"github.com/miku/span/formats/finc"
)

const (
	// Internal bookkeeping.
	SourceID = "49"
)

var (
	errNoDate = errors.New("date is missing")
	errNoURL  = errors.New("URL is missing")
)

var (
	DefaultFormat = "ElectronicArticle"

	// Load assets
	Formats  = assetutil.MustLoadStringMap("assets/crossref/formats.json")
	Genres   = assetutil.MustLoadStringMap("assets/crossref/genres.json")
	RefTypes = assetutil.MustLoadStringMap("assets/crossref/reftypes.json")

	// AuthorReplacer is a special cleaner for author names.
	AuthorReplacer = strings.NewReplacer("#", "", "--", "", "*", "", "|", "", "&NA;", "", "\u0026NA;", "", "\u0026", "")

	// ArticleTitleBlocker will trigger skips, if article title matches exactly.
	ArticleTitleBlocker = []string{"Titelei", "Front Matter", "Advertisement", "Advertisement:"}

	// ArticleTitleCleanerPatterns removes matching parts.
	ArticleTitleCleanerPatterns = []*regexp.Regexp{
		// refs. #5827
		regexp.MustCompile(`[?]{6,}`),
	}

	// Future ends soon.
	Future = time.Now().Add(time.Hour * 24 * 365 * 2)
)

// BulkResponse for a bulk request containing multiple items.
type BulkResponse struct {
	Status         string `json:"status"`
	MessageType    string `json:"message-type"`
	MessageVersion string `json:"message-version"`
	Message        struct {
		NextCursor   string     `json:"next-cursor"`
		TotalResults int        `json:"total-results"`
		Items        []Document `json:"items"`
	} `json:"message"`
}

// DatePart consists of up to three int, representing year, month, day.
type DatePart []int

// DateField contains two representations of one value.
type DateField struct {
	DateParts []DatePart `json:"date-parts"`
	Timestamp int64      `json:"timestamp"`
}

// Document is a example 'works' API response - message part only.
type Document struct {
	Author []struct {
		Family string `json:"family"`
		Given  string `json:"given"`
	} `json:"author"`
	ContainerTitle []string  `json:"container-title"`
	Deposited      DateField `json:"deposited"`
	DOI            string    `json:"DOI"`
	Indexed        DateField `json:"indexed"`
	ISSN           []string  `json:"ISSN"`
	Issue          string    `json:"issue"`
	Issued         DateField `json:"issued"`
	Member         string    `json:"member"`
	Page           string    `json:"page"`
	Prefix         string    `json:"prefix"`
	PublishedPrint DateField `json:"published-print"`
	Publisher      string    `json:"publisher"`
	ReferenceCount int       `json:"reference-count"`
	Score          float64   `json:"score"`
	Source         string    `json:"source"`
	Subjects       []string  `json:"subject"`
	Subtitle       []string  `json:"subtitle"`
	Title          []string  `json:"title"`
	Type           string    `json:"type"`
	URL            string    `json:"URL"`
	Volume         string    `json:"volume"`
}

// PageInfo holds various page related data.
type PageInfo struct {
	RawMessage string
	StartPage  int
	EndPage    int
}

// PageCount returns the number of pages, or zero if this cannot be determined.
func (pi *PageInfo) PageCount() int {
	if pi.StartPage != 0 && pi.EndPage != 0 {
		// an article, that starts at page 19 and ends at page 19 has one page
		count := pi.EndPage - pi.StartPage + 1
		if count > 0 {
			return count
		}
	}
	return 0
}

func (doc *Document) Authors() (authors []finc.Author) {
	for _, ra := range doc.Author {
		authors = append(authors, finc.Author{
			FirstName: AuthorReplacer.Replace(span.UnescapeTrim(ra.Given)),
			LastName:  AuthorReplacer.Replace(span.UnescapeTrim(ra.Family)),
		})
	}
	return authors
}

// ID is of the form <kind>-<source-id>-<id-base64-unpadded>
// We simple map any primary key of the source (preferably a URL)
// to a safer alphabet. Since the base64 part is not meant to be decoded
// we drop the padding. It is simple enough to recover the original value.
func (doc *Document) ID() string {
	return fmt.Sprintf("ai-%s-%s", SourceID, base64.RawURLEncoding.EncodeToString([]byte(doc.URL)))
}

// PageInfo parses a page specfication in a best effort manner into a PageInfo struct.
func (doc *Document) PageInfo() PageInfo {
	pi := PageInfo{RawMessage: doc.Page}
	parts := strings.Split(doc.Page, "-")
	if len(parts) != 2 {
		return pi
	}
	spage, err := strconv.Atoi(parts[0])
	if err != nil {
		return pi
	}
	pi.StartPage = spage

	epage, err := strconv.Atoi(parts[1])
	if err != nil {
		return pi
	}
	pi.EndPage = epage
	return pi
}

// Date returns a time.Date in a best effort manner. Date parts seem to be always
// present in the source document, while timestamp is only present if
// dateparts consist of all three: year, month and day.
// It is an error, if no valid date can be extracted.
func (d *DateField) Date() (t time.Time, err error) {
	if len(d.DateParts) == 0 {
		return t, errNoDate
	}
	parts := d.DateParts[0]
	var ds string
	switch len(parts) {
	case 1:
		ds = fmt.Sprintf("%04d-01-01", parts[0])
	case 2:
		ds = fmt.Sprintf("%04d-%02d-01", parts[0], parts[1])
	case 3:
		ds = fmt.Sprintf("%04d-%02d-%02d", parts[0], parts[1], parts[2])
	default:
		return t, nil
	}
	return time.Parse("2006-01-02", ds)
}

// CombinedTitle returns a longish title.
func (doc *Document) CombinedTitle() string {
	if len(doc.Title) > 0 {
		if len(doc.Subtitle) > 0 {
			return span.UnescapeTrim(fmt.Sprintf("%s : %s", strings.Join(doc.Title, " "), strings.Join(doc.Subtitle, " ")))
		}
		return span.UnescapeTrim(strings.Join(doc.Title, " "))
	}
	if len(doc.Subtitle) > 0 {
		return span.UnescapeTrim(strings.Join(doc.Subtitle, " "))
	}
	return ""
}

// ShortTitle returns the first main title only.
func (doc *Document) ShortTitle() (s string) {
	if len(doc.Title) > 0 {
		s = span.UnescapeTrim(doc.Title[0])
	}
	return
}

// ToIntermediateSchema converts a crossref document into IS.
func (doc *Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error
	output := finc.NewIntermediateSchema()

	output.Date, err = doc.PublishedPrint.Date()
	if err != nil {
		// Fallback to previous behaviour, refs #12321.
		output.Date, err = doc.Issued.Date()
	}

	if err != nil {
		return output, err
	}

	output.RawDate = output.Date.Format("2006-01-02")

	if doc.URL == "" {
		return output, errNoURL
	}

	output.ID = doc.ID()
	if len(output.ID) > span.KeyLengthLimit {
		return output, span.Skip{Reason: fmt.Sprintf("ID_TOO_LONG %s", output.ID)}
	}

	if output.Date.After(Future) {
		return output, span.Skip{Reason: fmt.Sprintf("TOO_FUTURISTIC %s", output.ID)}
	}

	if doc.Type == "journal-issue" {
		return output, span.Skip{Reason: fmt.Sprintf("JOURNAL_ISSUE %s", output.ID)}
	}

	output.ArticleTitle = doc.CombinedTitle()
	if len(output.ArticleTitle) == 0 {
		return output, span.Skip{Reason: fmt.Sprintf("NO_ATITLE %s", output.ID)}
	}

	for _, title := range ArticleTitleBlocker {
		if output.ArticleTitle == title {
			return output, span.Skip{Reason: fmt.Sprintf("BLOCKED_ATITLE %s", output.ID)}
		}
	}

	for _, p := range ArticleTitleCleanerPatterns {
		output.ArticleTitle = p.ReplaceAllString(output.ArticleTitle, "")
	}

	// refs. #8428
	if len(output.ArticleTitle) > 32000 {
		return output, span.Skip{Reason: fmt.Sprintf("TOO_LONG_TITLE %s", output.ID)}
	}

	output.DOI = doc.DOI // refs #6312 and #10923, most // URL seem valid
	output.Format = Formats.LookupDefault(doc.Type, DefaultFormat)
	output.Genre = Genres.LookupDefault(doc.Type, "unknown")
	output.ISSN = doc.ISSN
	output.Issue = strings.TrimLeft(doc.Issue, "0")
	output.Languages = []string{"eng"}
	output.Publishers = append(output.Publishers, doc.Publisher)
	output.RefType = RefTypes.LookupDefault(doc.Type, "GEN")
	output.SourceID = SourceID
	output.Subjects = doc.Subjects
	output.Type = doc.Type
	output.URL = append(output.URL, doc.URL)
	output.Volume = strings.TrimLeft(doc.Volume, "0")

	if len(doc.ContainerTitle) > 0 {
		output.JournalTitle = span.UnescapeTrim(doc.ContainerTitle[0])
	} else {
		return output, span.Skip{Reason: fmt.Sprintf("NO_JTITLE %s", output.ID)}
	}

	// refs #10864
	if strings.HasPrefix(doc.Type, "book-") {
		output.ArticleTitle = fmt.Sprintf("%s: %s", output.JournalTitle, output.ArticleTitle)
	}

	if len(doc.Subtitle) > 0 {
		output.ArticleSubtitle = span.UnescapeTrim(doc.Subtitle[0])
	}

	output.Authors = doc.Authors()

	// TODO(miku): do we need a config for these things?
	// Maybe a generic filter (in js?) that will gather exclusion rules?
	// if len(output.Authors) == 0 {
	// 	return output, span.Skip{Reason: fmt.Sprintf("NO_AUTHORS %s", output.ID)}
	// }

	pi := doc.PageInfo()
	output.StartPage = fmt.Sprintf("%d", pi.StartPage)
	output.EndPage = fmt.Sprintf("%d", pi.EndPage)
	output.Pages = pi.RawMessage
	output.PageCount = fmt.Sprintf("%d", pi.PageCount())

	// TODO: use a file for this
	publisherBlacklist := []string{
		"Crossref Testing",
		"test",
		"crossref-test",
	}

	for _, s := range publisherBlacklist {
		if doc.Publisher == s {
			return output, span.Skip{Reason: fmt.Sprintf("BLACKLISTED_COLLECTION %s", output.ID)}
		}
	}

	if doc.Publisher == "" {
		output.MegaCollections = []string{fmt.Sprintf("X-U (CrossRef)")}
	} else {
		publisher := span.UnescapeTrim(strings.Replace(doc.Publisher, "\n", " ", -1))
		output.MegaCollections = []string{fmt.Sprintf("%s (CrossRef)", publisher)}
	}

	return output, nil
}
