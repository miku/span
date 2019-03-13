package doaj

import (
	"encoding/xml"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
)

// Record was generated 2019-03-07 22:40:57 by tir on hayiti.
type Record struct {
	XMLName xml.Name `xml:"Record"`
	Text    string   `xml:",chardata"`
	Header  struct {
		Text       string   `xml:",chardata"`
		Status     string   `xml:"status,attr"`
		Identifier string   `xml:"identifier"` // oai:doaj.org/article:72a8...
		Datestamp  string   `xml:"datestamp"`  // 2017-12-31T23:00:18Z, 201...
		SetSpec    []string `xml:"setSpec"`    // TENDOk1lZGljYWwgcGh5c2ljc...
	} `xml:"header"`
	Metadata struct {
		Text string `xml:",chardata"`
		Dc   struct {
			Text           string   `xml:",chardata"`
			SchemaLocation string   `xml:"schemaLocation,attr"`
			Title          string   `xml:"title"`       // Diffusion Tensor Metric M...
			Identifier     []string `xml:"identifier"`  // 1178-623X, 10.4137/MRI.S1...
			Date           string   `xml:"date"`        // 2012-01-01T00:00:00Z, 201...
			Relation       []string `xml:"relation"`    // https://doi.org/10.4137/M...
			Description    string   `xml:"description"` // MRI and Monte Carlo simul...
			Creator        []string `xml:"creator"`     // Jonathan D. Thiessen, Tre...
			Publisher      string   `xml:"publisher"`   // SAGE Publishing, Tripoli ...
			Type           string   `xml:"type"`        // article
			Subject        []struct {
				Text string `xml:",chardata"` // Medical physics. Medical ...
				Type string `xml:"type,attr"`
			} `xml:"subject"`
			Language   []string `xml:"language"`   // EN
			Provenance string   `xml:"provenance"` // Journal Licence: CC BY-NC...
			Source     string   `xml:"source"`     // Magnetic Resonance Insigh...
		} `xml:"dc"`
	} `xml:"metadata"`
	About string `xml:"about"`
}

// Date tries to parse the date.
func (record Record) Date() (time.Time, error) {
	// <dc:date>2012-01-01T00:00:00Z</dc:date>
	return time.Parse("2006-01-02T15:04:05Z", record.Metadata.Dc.Date)
}

// DOI returns DOI or empty string.
func (record Record) DOI() string {
	for _, id := range record.Metadata.Dc.Identifier {
		if strings.HasPrefix(id, "10.") {
			return id
		}
	}
	return ""
}

// Authors returns authors.
func (record Record) Authors() (authors []finc.Author) {
	for _, creator := range record.Metadata.Dc.Creator {
		authors = append(authors, finc.Author{Name: html.UnescapeString(creator)})
	}
	return authors
}

// Identifier returns the DOAJ identifier, e.g. ce5cbc9701d14155b0b9a45373027d67.
func (record Record) Identifier() string {
	// https://doaj.org/article/ce5cbc9701d14155b0b9a45373027d67
	prefix := "https://doaj.org/article/"
	for _, id := range record.Metadata.Dc.Identifier {
		if strings.HasPrefix(id, prefix) {
			return strings.Replace(id, prefix, "", -1)
		}
	}
	return ""
}

// Links returns any URL associated with this record. There are more links in
// the dataset, just use the DOI for now.
func (record Record) Links() (links []string) {
	for _, v := range record.Metadata.Dc.Identifier {
		if strings.HasPrefix(v, "10.") {
			links = append(links, fmt.Sprintf("https://doi.org/%s", v))
		}
	}
	return links
}

// Volume parses volume from dc:source, https://git.io/fjekV.
func (record Record) Volume() string {
	// <dc:source>Case Reports in Oncology, Vol 10, Iss 3, Pp 1085-1091 (2017)</dc:source>
	re := regexp.MustCompile(`(?i)vol [0-9]*`)
	return strings.TrimSpace(strings.Replace(re.FindString(record.Metadata.Dc.Source), "Vol", "", -1))
}

func (record Record) Issue() string {
	// <dc:source>Case Reports in Oncology, Vol 10, Iss 3, Pp 1085-1091 (2017)</dc:source>
	re := regexp.MustCompile(`(?i)iss [0-9]*`)
	return strings.TrimSpace(strings.Replace(re.FindString(record.Metadata.Dc.Source), "Iss", "", -1))
}

// JournalTitle returns journal title.
func (record Record) JournalTitle() string {
	// <dc:source>Case Reports in Oncology, Vol 10, Iss 3, Pp 1085-1091 (2017)</dc:source>
	parts := strings.Split(record.Metadata.Dc.Source, ",")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return ""
}

// ISSN returns ISSN, if available.
func (record Record) ISSN() (issn []string) {
	re := regexp.MustCompile(`[0-9]{4,4}-[0-9xX]{4,4}`)
	for _, v := range record.Metadata.Dc.Identifier {
		if re.MatchString(v) {
			issn = append(issn, v)
		}
	}
	return issn
}

// StartPage returns start page.
func (record Record) StartPage() string {
	re := regexp.MustCompile(`[0-9]{1,5}-[0-9]{1,5}`)
	s := re.FindString(record.Metadata.Dc.Source)
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// EndPage returns end page or empty string.
func (record Record) EndPage() string {
	re := regexp.MustCompile(`[0-9]{1,5}-[0-9]{1,5}`)
	s := re.FindString(record.Metadata.Dc.Source)
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// Pages returns number of pages, or empty string.
func (record Record) Pages() string {
	spage, err := strconv.Atoi(record.StartPage())
	if err != nil {
		return ""
	}
	epage, err := strconv.Atoi(record.EndPage())
	if err != nil {
		return ""
	}
	pages := epage - spage
	if pages > 0 {
		return fmt.Sprintf("%d", pages)
	}
	return ""
}

// Subjects returns a mapped (https://git.io/fjesG) list of LCSH subjects.
func (record Record) Subjects() []string {
	subjects := container.NewStringSet()
	for _, v := range record.Metadata.Dc.Subject {
		if v.Type != "dcterms:LCSH" {
			continue
		}
		// RS1-441, Q, ...
		if len(v.Text) > 1 && !strings.Contains(v.Text, "-") {
			continue
		}
		class := LCCPatterns.LookupDefault(v.Text, finc.NotAssigned)
		if class != finc.NotAssigned {
			subjects.Add(class)
		}
	}
	return subjects.SortedValues()
}

// ToIntermediateSchema converts OAI record to intermediate schema.
func (record Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error

	output := finc.NewIntermediateSchema()
	date, err := record.Date()
	if err != nil {
		return output, span.Skip{Reason: "missing date"}
	}
	output.ArticleTitle = record.Metadata.Dc.Title
	output.Date = date
	output.RawDate = date.Format("2006-01-02")
	output.Authors = record.Authors()
	output.DOI = record.DOI()
	output.RecordID = record.Identifier()
	if output.RecordID == "" {
		return output, fmt.Errorf("missing record id")
	}
	output.SourceID = "28"
	output.ID = fmt.Sprintf("ai-28-%s", output.RecordID)

	output.URL = record.Links()
	output.JournalTitle = record.JournalTitle()
	output.Volume = record.Volume()
	output.Issue = record.Issue()
	output.ISSN = record.ISSN()
	output.StartPage = record.StartPage()
	output.EndPage = record.EndPage()
	output.PageCount = record.Pages()
	output.Pages = fmt.Sprintf("%s-%s", output.StartPage, output.EndPage)
	if output.Pages == "-" {
		output.Pages = ""
	}

	languages := container.NewStringSet()
	for _, l := range record.Metadata.Dc.Language {
		languages.Add(LanguageMap.LookupDefault(l, "und"))
	}
	output.Languages = languages.Values()
	output.Format = "ElectronicArticle"
	output.Genre = "article"
	output.RefType = "EJOUR"
	output.MegaCollections = []string{"DOAJ Directory of Open Access Journals"}

	// Subjects, if LCSH can be resolved.
	output.Subjects = record.Subjects()

	return output, nil
}
