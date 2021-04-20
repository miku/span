// Package genderopen, refs #13024.
package genderopen

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
)

// bookTitlePattern for extracting book title from dc.source.
var bookTitlePattern = regexp.MustCompile(`([^:]*):([^\(]*)`)

// Record was generated 2019-06-06 16:38:16 by tir on sol.
type Record struct {
	XMLName xml.Name `xml:"record"`
	Text    string   `xml:",chardata"`
	Header  struct {
		Text       string   `xml:",chardata"`
		Status     string   `xml:"status,attr"`
		Identifier string   `xml:"identifier"` // oai:www.genderopen.de:255...
		Datestamp  string   `xml:"datestamp"`  // 2017-11-30T13:54:17Z, 201...
		SetSpec    []string `xml:"setSpec"`    // com_25595_1, col_25595_3,...
	} `xml:"header"`
	Metadata struct {
		Text string `xml:",chardata"`
		Dc   struct {
			Text           string   `xml:",chardata"`
			OaiDc          string   `xml:"oai_dc,attr"`
			Doc            string   `xml:"doc,attr"`
			Xsi            string   `xml:"xsi,attr"`
			Dc             string   `xml:"dc,attr"`
			SchemaLocation string   `xml:"schemaLocation,attr"`
			Title          string   `xml:"title"`       // Ausweitung der Geschlecht...
			Creator        []string `xml:"creator"`     // Brunner, Claudia, Döllin...
			Contributor    []string `xml:"contributor"` // Lakitsch, Maximilian, Ste...
			Subject        []string `xml:"subject"`     // Geschlecht, Krieg, Gerech...
			Date           string   `xml:"date"`        // 2015, 2004, 2016, 2007, 2...
			Type           []string `xml:"type"`        // info:eu-repo/semantics/bo...
			Identifier     []string `xml:"identifier"`  // urn:ISBN:978-3-643-50677-...
			Language       string   `xml:"language"`    // ger
			Rights         []string `xml:"rights"`      // https://creativecommons.o...
			Format         string   `xml:"format"`      // application/pdf
			Publisher      []string `xml:"publisher"`   // LIT, Wien, VSA-Verlag, Ha...
			Source         string   `xml:"source"`      // Lakitsch, Maximilian; Ste...
			Description    []string `xml:"description"` // Nachdem kosmetische Genit...
		} `xml:"dc"`
	} `xml:"metadata"`
	About string `xml:"about"`
}

// BookTitle parses book title out of a citation string. Input may be "Knapp,
// Gudrun-Axeli; Wetterer, Angelika\n (Hrsg.): Achsen der Differenz.
// Gesellschaftstheorie und feministische Kritik II (Münster: Westfälisches
// Dampfboot, 2003), 73-100", https://play.golang.org/p/LApV7V_Ogz5. Fallback
// to original string, refs #13024.
func (r *Record) BookTitle() string {
	s := strings.Replace(r.Metadata.Dc.Source, "\n", " ", -2)
	matches := bookTitlePattern.FindStringSubmatch(s)
	if len(matches) == 3 {
		return strings.TrimSpace(matches[2])
	}
	return s
}

func parsePages(s string) (start, end, total string) {
	p := regexp.MustCompile(`([1-9][0-9]*)-([1-9][0-9]*)`)
	match := p.FindStringSubmatch(s)
	if len(match) < 3 {
		return "", "", ""
	}
	ss, es := match[1], match[2]
	u, _ := strconv.Atoi(ss)
	v, _ := strconv.Atoi(es)
	return ss, es, fmt.Sprintf("%d", v-u)
}

// stringsContainsAny returns true, if vals contains v, comparisons are case
// insensitive.
func stringsContainsAny(v string, vals []string) bool {
	for _, vv := range vals {
		if strings.ToLower(v) == strings.ToLower(vv) {
			return true
		}
	}
	return false
}

func (record Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	output.SourceID = "162"
	encodedRecordID := base64.RawURLEncoding.EncodeToString([]byte(record.Header.Identifier))
	output.RecordID = encodedRecordID
	output.ID = fmt.Sprintf("ai-%s-%s", output.SourceID, output.RecordID)
	output.MegaCollections = append(output.MegaCollections, "Gender Open")
	output.Genre = "article"
	output.RefType = "EJOUR"
	output.Format = "ElectronicArticle"
	output.Languages = []string{record.Metadata.Dc.Language}

	output.ArticleTitle = record.Metadata.Dc.Title

	for _, v := range record.Metadata.Dc.Creator {
		output.Authors = append(output.Authors, finc.Author{Name: v})
	}
	for _, v := range record.Metadata.Dc.Identifier {
		if strings.HasPrefix(v, "http") {
			output.URL = append(output.URL, v)
		}
		if strings.HasPrefix(v, "urn:ISSN:") {
			output.ISSN = append(output.ISSN, strings.Replace(v, "urn:ISSN:", "", 1))
		}
		if strings.HasPrefix(v, "http://dx.doi.org/") {
			output.DOI = strings.Replace(v, "http://dx.doi.org/", "", -1)
		}
	}

	// Article from books, articles from journals.
	if stringsContainsAny(output.ArticleTitle, []string{"zeitschrift", "journal"}) || len(output.ISSN) > 0 {
		output.JournalTitle = record.Metadata.Dc.Source
	} else {
		output.BookTitle = record.BookTitle()
	}
	for _, p := range record.Metadata.Dc.Publisher {
		output.Publishers = append(output.Publishers, p)
	}

	if record.Metadata.Dc.Date == "" {
		return output, span.Skip{Reason: "empty date"}
	}
	if len(record.Metadata.Dc.Date) < 4 {
		return output, span.Skip{Reason: "short date"}
	}
	if record.Metadata.Dc.Date != "" {
		s := record.Metadata.Dc.Date[:4] // XXX: Check.
		date, err := time.Parse("2006", s)
		if err != nil {
			return output, err
		}
		output.Date = date
		output.RawDate = output.Date.Format("2006-01-02")
	}

	for _, s := range record.Metadata.Dc.Subject {
		output.Subjects = append(output.Subjects, s)
	}
	start, end, total := parsePages(record.Metadata.Dc.Source)
	output.StartPage = start
	output.EndPage = end
	output.PageCount = total
	output.OpenAccess = true

	return output, nil
}
