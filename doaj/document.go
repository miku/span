// Directory of open access journals.
package doaj

import (
	"bufio"
	"encoding/json"
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
	SourceID = "28"
	// BatchSize for grouped channel transport.
	BatchSize = 25000
	// Collection name
	Collection = "DOAJ Directory of Open Access Journals"
	// Format for all records
	Format = "ElectronicArticle"
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

func (doc Document) Date() (s string) {
	if doc.Index.Date != "" {
		layout := "2006-01-02T15:04:05Z"
		t, err := time.Parse(layout, doc.Index.Date)
		if err != nil {
			return t.Format("2006-01-02")
		}
	}
	if IsYear(doc.BibJson.Year) {
		if IsMonth(doc.BibJson.Month) {
			return fmt.Sprintf("%04s-%02s-01", doc.BibJson.Year, doc.BibJson.Month)
		}
		return fmt.Sprintf("%04s-01-01", doc.BibJson.Year)
	}
	// TODO(miku): resolve missing date records
	// records w/o date seem to represent journal homepages and the like
	// should be skipable in general
	return "1970-01-01"
}

func (doc Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	output.SourceID = SourceID
	output.RecordID = doc.ID
	output.MegaCollection = Collection
	output.Format = Format

	output.ISSN = doc.Index.ISSN
	output.ArticleTitle = doc.BibJson.Title
	output.JournalTitle = doc.BibJson.Journal.Title
	output.Volume = doc.BibJson.Journal.Volume
	output.Publishers = append(output.Publishers, doc.BibJson.Journal.Publisher)
	output.RawDate = doc.Date()

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
		class := finc.RegexpResolve(strings.Replace(s, "LCC:", "", -1), finc.LCCFincMap)
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
		languages.Add(LanguageMap.Lookup(l, "und"))
	}
	output.Languages = languages.Values()

	for _, author := range doc.BibJson.Author {
		output.Authors = append(output.Authors, finc.Author{Name: author.Name})
	}

	return output, nil
}
