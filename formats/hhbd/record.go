package hhbd

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/miku/span"

	"github.com/miku/span/formats/finc"
)

var datePattern = regexp.MustCompile(`[012][0-9][0-9][0-9]`)

// Record was generated 2017-12-22 16:25:34 by tir on apollo.
type Record struct {
	XMLName xml.Name `xml:"Record"`
	Text    string   `xml:",chardata"`
	Header  struct {
		Text       string `xml:",chardata"`
		Status     string `xml:"status,attr"`
		Identifier struct {
			// oai:digi.ub.uni-heidelberg.de:2579, oai:digi.ub.uni-heidelberg.de:2583, oai:digi.ub.uni-heidelberg.d...
			Text string `xml:",chardata"`
		} `xml:"identifier"`
		Datestamp struct {
			// 2015-01-21T12:50:36Z, 2015-01-21T12:51:57Z, 2015-01-21T13:00:19Z, 2015-01-21T13:05:20Z, 2015-01-21T1...
			Text string `xml:",chardata"`
		} `xml:"datestamp"`
	} `xml:"header"`
	Metadata struct {
		Text string `xml:",chardata"`
		Dc   struct {
			Text      string `xml:",chardata"`
			Dcterms   string `xml:"dcterms,attr"`
			OaiDc     string `xml:"oai_dc,attr"`
			Europeana string `xml:"europeana,attr"`
			Dc        string `xml:"dc,attr"`
			Creator   struct {
				// Johann Wilhelm (Pfalz, Kurfürst), Wolfgang Lazius, Friedrich Riedrer, Johan Weier, Eugen Reiz Kunst...
				Text string `xml:",chardata"`
			} `xml:"creator"`
			Title struct {
				// Ordnung Des Hoch-Fürstlichen Gülich- und Bergischen Hoffgerichts zu Düsseldorff: Sambt denen an g...
				Text string `xml:",chardata"`
			} `xml:"title"`
			Date struct {
				// [um 1695] [VD17 1:018019V], 1619 [VD17-23:233236L], 1535, 1575, 1905 (Nr. 43-52), 1979, 1925, 1674 [...
				Text string `xml:",chardata"`
			} `xml:"date"`
			Issued struct {
				// [um 1695] [VD17 1:018019V], 1619 [VD17-23:233236L], 1535, 1575, 1905 (Nr. 43-52), 1979, 1925, 1674 [...
				Text string `xml:",chardata"`
			} `xml:"issued"`
			Identifier []struct {
				// http://digi.ub.uni-heidelberg.de/diglit/drwjuelichZO1697a, urn:nbn:de:bsz:16-diglit-25799, http://di...
				Text string `xml:",chardata"`
			} `xml:"identifier"`
			Language []struct {
				// de, de, x-unknown, de, x-unknown, pl, de, la, en, fr
				Text string `xml:",chardata"`
			} `xml:"language"`
			Type struct {
				// Monograph, Monograph, Monograph, Monograph, Volume, Volume, Monograph, Monograph, Volume, Volume
				Text string `xml:",chardata"`
			} `xml:"type"`
			Ispartof []struct {
				// Druckschriften, Heidelberger historische Bestände — digital: Rechtsquellen der frühen Neuzeit, D...
				Text string `xml:",chardata"`
			} `xml:"ispartof"`
			Subject []struct {
				// Auktionskataloge; Deutschland; Berlin; Eugen Reiz Kunst-Auktions-Haus <Berlin>, Auction Catalogs; Ge...
				Text string `xml:",chardata"`
			} `xml:"subject"`
			Spatial []struct {
				// München <1917>, München <1927>, Wien <1917>, Wien <1928>, Heidelberg, Vatikanstadt, Heidelberg, He...
				Text string `xml:",chardata"`
			} `xml:"spatial"`
			Temporal []struct {
				// Geschichte 1830-1917, Geschichte 1850-1920, Geschichte 1700-1900, Geschichte 1623, Geschichte 1300-1...
				Text string `xml:",chardata"`
			} `xml:"temporal"`
			Alternative struct {
				// Zweihundert Jahre Darmstädter Kunst, Zweihundert Jahre Darmstädter Kunst, Willehalm von Orlens, Hu...
				Text string `xml:",chardata"`
			} `xml:"alternative"`
		} `xml:"dc"`
	} `xml:"metadata"`
	About struct {
		Text string `xml:",chardata"`
	} `xml:"about"`
}

func (record Record) date() (time.Time, error) {
	results := datePattern.FindAllString(record.Metadata.Dc.Date.Text, -1)
	if len(results) == 0 {
		return time.Time{}, nil
	}
	return time.Parse("2006", results[0])
}

// ToIntermediateSchema converts document.
func (record Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error
	output := finc.NewIntermediateSchema()
	output.RecordID = record.Header.Identifier.Text
	output.SourceID = "107"
	output.ID = fmt.Sprintf("ai-%s-%s", output.SourceID, base64.RawURLEncoding.EncodeToString([]byte(output.RecordID)))
	output.ArticleTitle = record.Metadata.Dc.Title.Text
	output.MegaCollections = []string{"Heidelberger Historische Bestände (Digital)"}

	// Date.
	date, err := record.date()
	if err != nil {
		return nil, span.Skip{Reason: fmt.Sprintf("Cannot parse date: %s", record.Metadata.Dc.Date.Text)}
	}
	output.Date = date
	output.RawDate = date.Format("2006-01-02")
	if output.Date.IsZero() {
		return nil, span.Skip{Reason: fmt.Sprintf("Zero date: %s", record.Metadata.Dc.Date.Text)}
	}

	// Authors.
	for _, creator := range strings.Split(record.Metadata.Dc.Creator.Text, ";") {
		creator = strings.TrimSpace(creator)
		if len(creator) == 0 {
			continue
		}
		output.Authors = append(output.Authors, finc.Author{Name: creator})
	}

	// Subjects.
	for _, s := range record.Metadata.Dc.Subject {
		text := strings.TrimSpace(s.Text)
		if text == "" {
			continue
		}
		output.Subjects = append(output.Subjects, text)
	}

	// Ignore other potential container titles.
	if len(record.Metadata.Dc.Ispartof) > 0 {
		output.JournalTitle = record.Metadata.Dc.Ispartof[0].Text
	}

	// URLs.
	for _, id := range record.Metadata.Dc.Identifier {
		if strings.HasPrefix(id.Text, "http") {
			// XXX: Target contains IIIF manifest.
			output.URL = append(output.URL, id.Text)
		}
		if strings.HasPrefix(id.Text, "urn:nbn:") {
			output.URL = append(output.URL, fmt.Sprintf("http://nbn-resolving.de/%s", id.Text))
		}
	}

	// Languages.
	for _, l := range record.Metadata.Dc.Language {
		tlc := strings.TrimSpace(span.LanguageIdentifier(l.Text))
		if tlc == "" {
			continue
		}
		output.Languages = append(output.Languages, tlc)
	}

	return output, err
}
