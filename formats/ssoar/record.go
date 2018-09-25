package ssoar

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goodsign/monday"
	log "github.com/sirupsen/logrus"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/formats/marc"
)

type Record struct {
	marc.Record
}

// Title returns the a record title.
func (r Record) Title() string {
	result := r.MustGetFirstDataField("245.a")
	subtitle := strings.TrimSpace(r.MustGetFirstDataField("245.b"))
	if subtitle != "" {
		result = result + ": " + subtitle
	}
	return result
}

// JournalTitle tries to parse out a journal title.
func (r Record) JournalTitle() string {
	// In: Journal of Social Work Practice ; 19 (2005) 1 ; 87-101
	// In: Balzer, Wolfgang (Hg.), Pearce, David A. (Hg.), Schmidt,
	// Heinz-Jürgen (Hg.): Reduction in science : structure, examples,
	// philosophical problems. 1984. S. 331-357. ISBN 90-277-1811-3
	for _, s := range r.MustGetDataFields("500.a") {
		if !strings.HasPrefix(s, "In:") {
			continue
		}
		parts := strings.Split(s, ";")
		switch len(parts) {
		case 1:
			pp := strings.Split(s[4:], ":")
			switch len(pp) {
			case 1:
				ppp := strings.Split(s[4:], ".")
				return ppp[0]
			case 2:
				ppp := strings.Split(pp[1], ".")
				return ppp[0]
			case 3:
				a := pp[1]
				b := strings.Split(pp[2], ".")
				return fmt.Sprintf("%s: %s", a, b)
			default:
				return s
			}
		case 2:
			return strings.TrimSpace(parts[0][4:])
		case 3:
			return strings.TrimSpace(parts[0][4:])
		case 4:
			return strings.TrimSpace(parts[0][4:])
		default:
			log.Printf("unparsed parent title: %s", s)
		}
	}
	return ""
}

// ID returns an identifier.
func (r Record) ID() (string, error) {
	parts := strings.Split(r.Header.Identifier.Text, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected identifier: %s", r.Header.Identifier.Text)
	} else {
		return strings.TrimSpace(parts[1]), nil
	}
}

// FindYear tries to find a year.
func (r Record) FindYear() string {
	date := strings.TrimSpace(r.MustGetFirstDataField("264.c"))
	if len(date) == 4 {
		// XXX: o.J.
		if date == "o.J." {
			return fmt.Sprintf("2010-01-01")
		}
		return fmt.Sprintf("%s-01-01", date)
	}
	yp := regexp.MustCompile(`[12][6789012]\d\d`)
	for _, f := range r.MustGetDataFields("500.a") {
		matches := yp.FindAllString(f, -1)
		if len(matches) > 0 {
			return fmt.Sprintf("%s-01-01", matches[0])
		}
	}
	return "1970-01-01"
}

// FindFormat tries to parse a format.
func (r Record) FindFormat() string {
	leader := string(r.Metadata.Record.Leader.Text)
	llen := len(string(leader))
	switch {
	case llen > 7 && string(leader[7]) == "m":
		return "eBook"
	case llen > 19 && (string(leader[19]) == " " || string(leader[19]) == "b"):
		return "eBook"
	default:
		return "ElectronicArticle"
	}
}

// HasEmbargo looks for a fixed string in the MARC record and tries to find
// out, whether the embargo holds. Some example text (mixed date formats): "Der
// Volltext unterliegt einer Embargofrist bis zum 18. Okt. 2018."
func (r Record) HasEmbargo() (time.Time, bool) {
	loc, _ := time.LoadLocation("Europe/Berlin")

	locs := []monday.Locale{monday.LocaleDeDE, monday.LocaleEnUS}
	templates := []string{
		`Der Volltext unterliegt einer Embargofrist bis zum 02. Jan. 2006.`,
		`Der Volltext unterliegt einer Embargofrist bis zum 02. Jan 2006.`,
	}

	for _, f := range r.MustGetDataFields("500.a") {
		for _, l := range locs {
			for _, tmpl := range templates {
				t, err := monday.ParseInLocation(tmpl, f, loc, l)
				if err == nil {
					if t.After(time.Now()) {
						return t, true
					}
				}
			}
		}
	}
	return time.Time{}, false
}

// stringDifference returns the numeric value of a-b of strings a and b as
// string, an empty string if something goes wrong.
func stringDifference(a, b string) string {
	i, err := strconv.Atoi(a)
	if err != nil {
		return ""
	}
	j, err := strconv.Atoi(b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", i-j)
}

// FindPages returns start, end and page count as string.
func (r Record) FindPages() (string, string, string) {
	v := r.MustGetFirstDataField("300.a")
	p := regexp.MustCompile(`([1-9][0-9]*)`)
	matches := p.FindAllString(v, -1)
	switch {
	case len(matches) == 1:
		return matches[0], matches[0], "1"
	case len(matches) == 2:
		return matches[0], matches[1], stringDifference(matches[1], matches[0])
	default:
		// Try to find pages in 500.a
		q := regexp.MustCompile(`([1-9][0-9]{0,3})-([1-9][0-9]{0,3})`)
		for _, w := range r.MustGetDataFields("500.a") {
			sms := q.FindAllStringSubmatch(w, -1)
			if sms == nil {
				return "", "", ""
			}
			if len(sms[0]) >= 3 {
				pagecount := stringDifference(sms[0][2], sms[0][1])
				return sms[0][0], sms[0][1], pagecount
			}
		}
	}
	return "", "", ""
}

// ToIntermediateSchema converts a MarcXML-ish record.
func (r Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()
	id, err := r.ID()
	if err != nil {
		return output, err
	}

	if t, ok := r.HasEmbargo(); ok {
		msg := fmt.Sprintf("embargo restriction for %s", id)
		log.Printf("embargo for %s expires on %s", id, t.Format("2006-01-02"))
		return output, span.Skip{Reason: msg}
	}

	output.RecordID = id
	output.SourceID = "30"
	output.ID = fmt.Sprintf("ai-%s-%s", output.SourceID, output.RecordID)
	output.Format = r.FindFormat()
	output.MegaCollections = []string{"SSOAR Social Science Open Access Repository"}

	switch output.Format {
	case "eBook":
		output.BookTitle = r.Title()
		output.Genre = "book"
	default:
		output.ArticleTitle = r.Title()
		output.Genre = "article"
	}
	output.JournalTitle = r.JournalTitle()
	output.URL = r.MustGetDataFields("856.u")
	for _, author := range r.MustGetDataFields("100.a") {
		output.Authors = append(output.Authors, finc.Author{Name: author})
	}
	for _, author := range r.MustGetDataFields("700.a") {
		output.Authors = append(output.Authors, finc.Author{Name: author})
	}
	output.Abstract = r.MustGetFirstDataField("520.a")
	output.Subjects = r.MustGetDataFields("650.a")

	output.RawDate = r.FindYear()
	date, err := time.Parse("2006-01-02", output.RawDate)
	if err != nil {
		log.Fatal(err)
	}
	output.Date = date
	output.Languages = r.MustGetDataFields("041.a")
	output.StartPage, output.EndPage, output.PageCount = r.FindPages()
	output.Series = r.MustGetFirstDataField("490.a")
	output.ISBN = r.MustGetDataFields("020.a")

	for _, place := range r.MustGetDataFields("264.a") {
		if place != "" {
			output.Places = append(output.Places, place)
		}
	}
	if pub := r.MustGetFirstDataField("264.b"); pub != "" {
		output.Publishers = append(output.Publishers, pub)
	}
	return output, nil
}
