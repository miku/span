// Directory of open access journals.
package doaj

import (
	"bufio"
	"encoding/json"
	"io"
	"log"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

const (
	// Internal bookkeeping.
	SourceID = "28"
	// BatchSize for grouped channel transport.
	BatchSize = 25000
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
			doc := new(Document)
			err := json.Unmarshal([]byte(s.(string)), doc)
			if err != nil {
				return doc, err
			}
			return doc, nil
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

func (doc *Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	output.ISSN = doc.Index.ISSN
	// TODO(miku): language code lookups
	output.Languages = doc.Index.Language
	output.ArticleTitle = doc.BibJson.Title

	return output, nil
}
