// Directory of open access journals.
package doaj

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/assetutil"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
)

const (
	// Internal bookkeeping.
	SourceID = "28"
	// BatchSize for grouped channel transport.
	BatchSize = 25000
	// Collection name
	Collection = "DOAJ Directory of Open Access Journals"
	// Format for all records
	Format = "ElectronicArticle"
)

var errDateMissing = errors.New("date is missing")

var (
	LCCPatterns = assetutil.MustLoadRegexpMap("assets/finc/lcc.json")
	LanguageMap = assetutil.MustLoadStringMap("assets/doaj/language-iso-639-3.json")
)

type Response struct {
	ID     string   `json:"_id"`
	Index  string   `json:"_index"`
	Source Document `json:"_source"`
	Type   string   `json:"_type"`
}

type Document struct {
	BibJson BibJson `json:"bibjson"`
	Created string  `json:"created_date"`
	ID      string  `json:"id"`
	Index   Index   `json:"index"`
	Updated string  `json:"last_updated"`
	// make Response.Type available here
	Type string
}

type Index struct {
	Classification []string `json:"classification"`
	Country        string   `json:"country"`
	Date           string   `json:"date"`
	ISSN           []string `json:"issn"`
	Language       []string `json:"language"`
	License        []string `json:"license"`
	Publishers     []string `json:"publisher"`
	SchemaCode     []string `json:"schema_code"`
	SchemaSubjects []string `json:"schema_subjects"`
	Subjects       []string `json:"subject"`
}

type License struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}

type Journal struct {
	Country   string    `json:"country"`
	Language  []string  `json:"language"`
	License   []License `json:"license"`
	Number    string    `json:"number"`
	Publisher string    `json:"publisher"`
	Title     string    `json:"title"`
	Volume    string    `json:"volume"`
}

type Author struct {
	Name string `json:"name"`
}

type Subject struct {
	Code   string `json:"code"`
	Scheme string `json:"scheme"`
	Term   string `json:"term"`
}

type Link struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Identifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type BibJson struct {
	Abstract   string       `json:"abstract"`
	Author     []Author     `json:"author"`
	EndPage    string       `json:"end_page"`
	Identifier []Identifier `json:"identifier"`
	Journal    Journal      `json:"journal"`
	Link       []Link       `json:"link"`
	Month      string       `json:"month"`
	StartPage  string       `json:"start_page"`
	Subject    []Subject    `json:"subject"`
	Title      string       `json:"title"`
	Year       string       `json:"year"`
}

type DOAJ struct{}

// NewBatch wraps up a new batch for channel com.
func NewBatch(lines []string) span.Batcher {
	batch := span.Batcher{
		Apply: func(s interface{}) (span.Importer, error) {
			resp := new(Response)
			err := json.Unmarshal([]byte(s.(string)), resp)
			if err != nil {
				return resp.Source, err
			}
			resp.Source.Type = resp.Type
			return resp.Source, nil
		}, Items: make([]interface{}, len(lines))}
	for i, line := range lines {
		batch.Items[i] = line
	}
	return batch
}

func (s DOAJ) Iterate(r io.Reader) (<-chan interface{}, error) {
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

// Date return the document date. Journals entries usually have no date, so
// they will err.
func (doc Document) Date() (time.Time, error) {
	if doc.Index.Date != "" {
		return time.Parse("2006-01-02T15:04:05Z", doc.Index.Date)
	}
	var s string
	if y, err := strconv.Atoi(doc.BibJson.Year); err == nil {
		s = fmt.Sprintf("%04d-01-01", y)
		if m, err := strconv.Atoi(doc.BibJson.Month); err == nil {
			if m > 0 && m < 13 {
				s = fmt.Sprintf("%04d-%02d-01", y, m)
			}
		}
	}
	return time.Parse("2006-01-02", s)
}

func (doc Document) DOI() string {
	for _, identifier := range doc.BibJson.Identifier {
		if identifier.Type == "doi" {
			return identifier.ID
		}
	}
	return ""
}

// ToIntermediateSchema converts a doaj document to intermediate schema. For
// now any record, that has no usable date will be skipped.
func (doc Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error

	output := finc.NewIntermediateSchema()
	output.Date, err = doc.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}

	output.DOI = doc.DOI()
	output.Format = Format
	output.MegaCollection = Collection
	output.RecordID = fmt.Sprintf("ai-%s-%s", SourceID, doc.ID)
	output.SourceID = SourceID

	output.ISSN = doc.Index.ISSN
	output.ArticleTitle = doc.BibJson.Title
	output.JournalTitle = doc.BibJson.Journal.Title
	output.Volume = doc.BibJson.Journal.Volume
	output.Publishers = append(output.Publishers, doc.BibJson.Journal.Publisher)

	for _, link := range doc.BibJson.Link {
		output.URL = append(output.URL, link.URL)
	}

	output.StartPage = doc.BibJson.StartPage
	output.EndPage = doc.BibJson.EndPage

	if sp, err := strconv.Atoi(doc.BibJson.StartPage); err == nil {
		if ep, err := strconv.Atoi(doc.BibJson.EndPage); err == nil {
			output.PageCount = fmt.Sprintf("%d", ep-sp)
			output.Pages = fmt.Sprintf("%d-%d", sp, ep)
		}
	}

	subjects := container.NewStringSet()
	for _, s := range doc.Index.SchemaCode {
		class := LCCPatterns.LookupDefault(strings.Replace(s, "LCC:", "", -1), finc.NOT_ASSIGNED)
		if class != finc.NOT_ASSIGNED {
			subjects.Add(class)
		}
	}
	if subjects.Size() == 0 {
		output.Subjects = []string{finc.NOT_ASSIGNED}
	} else {
		output.Subjects = subjects.SortedValues()
	}

	languages := container.NewStringSet()
	for _, l := range doc.Index.Language {
		languages.Add(LanguageMap.LookupDefault(l, "und"))
	}
	output.Languages = languages.Values()

	for _, author := range doc.BibJson.Author {
		output.Authors = append(output.Authors, finc.Author{Name: author.Name})
	}

	return output, nil
}
