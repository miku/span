package crossref

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/finc"
	"github.com/miku/span/sets"
)

const (
	// SourceID for internal bookkeeping.
	SourceID = 49
	// BatchSize for batched reading.
	BatchSize = 25000
)

// acceptedLanguages restricts the possible languages for detection.
var acceptedLanguages = sets.NewStringSet("de", "en", "fr", "it", "es")

// Crossref source.
type Crossref struct{}

// Iterate returns a channel which carries batches. The processor function
// is just plain JSON deserialization. It is ok to halt the world, if
// there some error during reading.
func (c Crossref) Iterate(r io.Reader) (chan interface{}, error) {
	batch := span.Batcher{
		Apply: func(s string) (span.Importer, error) {
			doc := new(Document)
			err := json.Unmarshal([]byte(s), doc)
			if err != nil {
				return doc, err
			}
			return doc, nil
		}}

	ch := make(chan interface{})
	reader := bufio.NewReader(r)
	i := 1
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			batch.Items = append(batch.Items, line)
			if i == BatchSize {
				ch <- batch
				batch.Items = batch.Items[:0]
				i = 0
			}
			i++
		}
		ch <- batch
		close(ch)
	}()
	return ch, nil
}

// Author is given by family and given name.
type Author struct {
	Family string `json:"family"`
	Given  string `json:"given"`
}

// String pretty prints the author.
func (author *Author) String() string {
	if author.Given != "" {
		if author.Family != "" {
			return fmt.Sprintf("%s, %s", author.Family, author.Given)
		}
		return author.Given
	}
	return author.Family
}

// DatePart consists of up to three int, representing year, month, day.
type DatePart []int

// DateField contains two representations of one value.
type DateField struct {
	DateParts []DatePart `json:"date-parts"`
	Timestamp int64      `json:"timestamp"`
}

// Document is a example 'works' API response.
type Document struct {
	Authors        []Author  `json:"author"`
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
		count := pi.EndPage - pi.StartPage
		if count > 0 {
			return count
		}
	}
	return 0
}

func (doc *Document) RecordID() string {
	return fmt.Sprintf("ai-%d-%s", SourceID, base64.StdEncoding.EncodeToString([]byte(doc.URL)))
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

// Year returns the first year found inside a DateField.
func (d *DateField) Year() int {
	if len(d.DateParts) >= 1 && len(d.DateParts[0]) > 0 {
		return d.DateParts[0][0]
	}
	return 0
}

// Date returns a time.Date in a best effort manner. Date parts seem to be always
// present in the source document, while timestamp is only present if
// dateparts consist of all three: year, month and day.
func (d *DateField) Date() (t time.Time) {
	if len(d.DateParts) == 0 {
		t, _ = time.Parse("2006-01-02", "0000-00-00")
		return t
	}
	p := d.DateParts[0]
	switch len(p) {
	case 1:
		t, _ = time.Parse("2006-01-02", fmt.Sprintf("%04d-01-01", p[0]))
	case 2:
		t, _ = time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-01", p[0], p[1]))
	case 3:
		t, _ = time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-%02d", p[0], p[1], p[2]))
	default:
		t, _ = time.Parse("2006-01-02", "1970-01-01")
	}
	return t
}

// CombinedTitle returns a longish title.
func (doc *Document) CombinedTitle() string {
	if len(doc.Title) > 0 {
		if len(doc.Subtitle) > 0 {
			return fmt.Sprintf("%s : %s", doc.Title[0], doc.Subtitle[0])
		}
		return doc.Title[0]
	}
	if len(doc.Subtitle) > 0 {
		return doc.Subtitle[0]
	}
	return ""
}

// FullTitle returns everything title.
func (doc *Document) FullTitle() string {
	return strings.Join(append(doc.Title, doc.Subtitle...), " ")
}

// ShortTitle returns the first main title only.
func (doc *Document) ShortTitle() string {
	if len(doc.Title) > 0 {
		return doc.Title[0]
	}
	return ""
}

// MemberName resolves the primary name of the member.
func (doc *Document) MemberName() (name string, err error) {
	id, err := doc.ParseMemberID()
	if err != nil {
		return
	}
	name, err = LookupMemberName(id)
	return
}

// ParseMemberID extracts the numeric member id.
func (doc *Document) ParseMemberID() (int, error) {
	fields := strings.Split(doc.Member, "/")
	if len(fields) > 0 {
		id, err := strconv.Atoi(fields[len(fields)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid member: %s", doc.Member)
		}
		return id, nil
	}
	return 0, fmt.Errorf("invalid member: %s", doc.Member)
}

// Languages returns guessed languages of the document.
// Always includes English as first guess and maybe another
// which is inferred from the short title and subtitle strings
// (high false positive rate).
func (doc *Document) Languages() (langs []string) {
	langs = append(langs, "en")

	// TODO(miku): move lang detect to is-export, address 90% time spent here
	// var buf bytes.Buffer
	// for _, t := range append(doc.Title, doc.Subtitle...) {
	// 	buf.WriteString(t)
	// }

	// lang := franco.DetectOne(buf.String())
	// if lang.Code == "und" {
	// 	return
	// }
	// base, err := language.ParseBase(lang.Code)
	// if err != nil {
	// 	return
	// }
	// if !acceptedLanguages.Contains(base.String()) {
	// 	return
	// }
	// if base.String() != "en" {
	// 	langs = append(langs, base.String())
	// }
	return langs
}

// ToIntermediateSchema converts a crossref document into IS.
func (doc *Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := new(finc.IntermediateSchema)

	if doc.URL == "" {
		return output, errors.New("input document has no URL")
	}

	output.Version = finc.IntermediateSchemaVersion
	output.RecordID = doc.RecordID()
	output.URL = append(output.URL, doc.URL)
	output.DOI = doc.DOI
	output.SourceID = SourceID
	output.Publisher = append(output.Publisher, doc.Publisher)
	output.ArticleTitle = doc.CombinedTitle()
	output.Format = "ElectronicArticle"
	output.Issue = doc.Issue
	output.Volume = doc.Volume
	output.ISSN = doc.ISSN

	if len(doc.ContainerTitle) > 0 {
		output.JournalTitle = doc.ContainerTitle[0]
	}

	for _, author := range doc.Authors {
		output.Authors = append(output.Authors, finc.Author{FirstName: author.Given, LastName: author.Family})
	}

	pi := doc.PageInfo()
	output.StartPage = fmt.Sprintf("%d", pi.StartPage)
	output.EndPage = fmt.Sprintf("%d", pi.EndPage)
	output.Pages = pi.RawMessage
	output.PageCount = fmt.Sprintf("%d", pi.PageCount())

	output.RawDate = doc.Issued.Date().Format("2006-01-02")

	output.Subjects = doc.Subjects

	name, err := doc.MemberName()
	if err == nil {
		output.MegaCollection = fmt.Sprintf("%s (CrossRef)", name)
	}

	output.Languages = doc.Languages()

	return output, nil
}
