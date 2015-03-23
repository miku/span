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
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
)

const (
	// Internal bookkeeping.
	SourceID = "49"
	// BatchSize for grouped channel transport.
	BatchSize = 25000
)

var (
	errNoDate = errors.New("date is missing")
	errNoURL  = errors.New("URL is missing")
)

var (
	DefaultFormat = "ElectronicArticle"
	Formats       = container.StringMap{
		"book":                "eBook",
		"book-chapter":        "ElectronicBookPart",
		"book-part":           "ElectronicBookPart",
		"book-section":        "ElectronicBookPart",
		"book-series":         "ElectronicSerial",
		"component":           "Unknown",
		"dataset":             "ElectronicResourceRemoteAccess",
		"dissertation":        "ElectronicThesis",
		"journal":             "ElectronicJournal",
		"journal-article":     "ElectronicArticle",
		"journal-issue":       "ElectronicJournal",
		"monograph":           "eBook",
		"proceedings":         "ElectronicProceeding",
		"proceedings-article": "ElectronicProceeding",
		"reference-book":      "eBook",
		"reference-entry":     "ElectronicResourceRemoteAccess",
		"report":              "ElectronicArticle",
		"report-series":       "ElectronicSerial",
	}
	Genres = container.StringMap{
		"book":                "book",
		"book-chapter":        "bookitem",
		"book-part":           "bookitem",
		"book-section":        "bookitem",
		"book-series":         "bookitem",
		"component":           "document",
		"dataset":             "document",
		"dissertation":        "book",
		"journal":             "unknown",
		"journal-article":     "article",
		"journal-issue":       "issue",
		"monograph":           "book",
		"proceedings":         "proceeding",
		"proceedings-article": "proceeding",
		"reference-book":      "book",
		"reference-entry":     "document",
		"report":              "report",
		"report-series":       "report",
	}
	RefTypes = container.StringMap{
		"book":                "EBOOK",
		"book-chapter":        "ECHAP",
		"book-part":           "ECHAP",
		"book-section":        "ECHAP",
		"book-series":         "SER",
		"component":           "GEN",
		"dataset":             "DATA",
		"dissertation":        "THES",
		"journal":             "EJOUR",
		"journal-article":     "EJOUR",
		"journal-issue":       "EJOUR",
		"monograph":           "EBOOK",
		"proceedings":         "CONF",
		"proceedings-article": "CONF",
		"reference-book":      "EBOOK",
		"reference-entry":     "GEN",
		"report":              "RPRT",
		"report-series":       "SER",
	}
)

// Crossref source.
type Crossref struct{}

// NewBatch wraps up a new batch for channel com.
func NewBatch(lines []string) span.Batcher {
	batch := span.Batcher{
		Apply: func(s string) (span.Importer, error) {
			doc := new(Document)
			err := json.Unmarshal([]byte(s), doc)
			if err != nil {
				return doc, err
			}
			return doc, nil
		}, Items: make([]string, len(lines))}
	copy(batch.Items, lines)
	return batch
}

// Iterate returns a channel which carries batches. The processor function
// is just plain JSON deserialization. It is ok to halt the world,
// if there some error during reading.
func (c Crossref) Iterate(r io.Reader) (<-chan interface{}, error) {
	ch := make(chan interface{})
	reader := bufio.NewReader(r)
	i := 0
	var lines []string
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			i++
			lines = append(lines, line)
			if i == BatchSize {
				ch <- NewBatch(lines)
				lines = lines[:0]
				i = 0
			}
		}
		ch <- NewBatch(lines)
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

// RecordID is of the form <kind>-<source-id>-<id-base64-unpadded>
// We simple map any primary key of the source (preferably a URL)
// to a safer alphabet. Since the base64 part is not meant to be decoded
// we drop the padding. It is simple enough to recover the original value.
func (doc *Document) RecordID() string {
	enc := fmt.Sprintf("ai-%s-%s", SourceID, base64.StdEncoding.EncodeToString([]byte(doc.URL)))
	return strings.TrimRight(enc, "=")
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
	switch len(parts) {
	case 1:
		t, err = time.Parse("2006-01-02", fmt.Sprintf("%04d-01-01", parts[0]))
		if err != nil {
			return t, err
		}
	case 2:
		t, err = time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-01", parts[0], parts[1]))
		if err != nil {
			return t, err
		}
	case 3:
		t, err = time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-%02d", parts[0], parts[1], parts[2]))
		if err != nil {
			return t, err
		}
	}
	return t, err
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
func (doc *Document) ShortTitle() (s string) {
	if len(doc.Title) > 0 {
		s = doc.Title[0]
	}
	return
}

// MemberName resolves the primary name of the member.
func (doc *Document) MemberName() (name string, err error) {
	id, err := doc.parseMemberID()
	if err != nil {
		return
	}
	name, err = LookupMemberName(id)
	return
}

// ParseMemberID extracts the numeric member id.
func (doc *Document) parseMemberID() (id int, err error) {
	fields := strings.Split(doc.Member, "/")
	if len(fields) > 0 {
		id, err = strconv.Atoi(fields[len(fields)-1])
		if err != nil {
			return id, fmt.Errorf("invalid member: %s", doc.Member)
		}
		return id, nil
	}
	return id, fmt.Errorf("invalid member: %s", doc.Member)
}

// ToIntermediateSchema converts a crossref document into IS.
func (doc *Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	date, err := doc.Issued.Date()
	output.RawDate = date.Format("2006-01-02")
	if err != nil {
		return output, err
	}

	if doc.URL == "" {
		return output, errNoURL
	}

	output.ArticleTitle = doc.CombinedTitle()
	output.DOI = doc.DOI
	output.Format = Formats.Lookup(doc.Type, DefaultFormat)
	output.Genre = Genres.Lookup(doc.Type, "unknown")
	output.ISSN = doc.ISSN
	output.Issue = doc.Issue
	output.Languages = []string{"eng"}
	output.Publishers = append(output.Publishers, doc.Publisher)
	output.RecordID = doc.RecordID()
	output.RefType = RefTypes.Lookup(doc.Type, "GEN")
	output.SourceID = SourceID
	output.Subjects = doc.Subjects
	output.Type = doc.Type
	output.URL = append(output.URL, doc.URL)
	output.Volume = doc.Volume

	if len(doc.ContainerTitle) > 0 {
		output.JournalTitle = doc.ContainerTitle[0]
	}

	if len(doc.Subtitle) > 0 {
		output.ArticleSubtitle = doc.Subtitle[0]
	}

	for _, author := range doc.Authors {
		output.Authors = append(output.Authors, finc.Author{
			FirstName: author.Given, LastName: author.Family})
	}

	pi := doc.PageInfo()
	output.StartPage = fmt.Sprintf("%d", pi.StartPage)
	output.EndPage = fmt.Sprintf("%d", pi.EndPage)
	output.Pages = pi.RawMessage
	output.PageCount = fmt.Sprintf("%d", pi.PageCount())

	name, err := doc.MemberName()
	if err == nil {
		output.MegaCollection = fmt.Sprintf("%s (CrossRef)", name)
	} else {
		output.MegaCollection = fmt.Sprintf("X-U (CrossRef)")
	}

	return output, nil
}
