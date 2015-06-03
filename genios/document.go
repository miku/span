package genios

import (
	"bufio"
	"encoding/xml"
	"io"
	"log"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

const (
	SourceID  = "48"
	BatchSize = 2000
)

type Document struct {
	ID               string   `xml:"ID,attr"`
	ISSN             string   `xml:"ISSN"`
	Source           string   `xml:"Source"`
	PublicationTitle string   `xml:"Publication-Title"`
	Title            string   `xml:"Title"`
	Year             string   `xml:"Year"`
	RawDate          string   `xml:"Date"`
	Volume           string   `xml:"Volume"`
	Issue            string   `xml:"Issue"`
	Authors          []string `xml:"Authors>Author"`
	Language         string   `xml:"Language"`
}

var RawDateReplacer = strings.NewReplacer(`"`, "", "\n", "", "\t", "")

type Genios struct{}

// NewBatch wraps up a new batch for channel com.
func NewBatch(docs []*Document) span.Batcher {
	batch := span.Batcher{
		Apply: func(s interface{}) (span.Importer, error) {
			return s.(span.Importer), nil
		}, Items: make([]interface{}, len(docs))}
	for i, doc := range docs {
		batch.Items[i] = doc
	}
	return batch
}

// Iterate emits Converter elements via XML decoding.
func (s Genios) Iterate(r io.Reader) (<-chan interface{}, error) {
	ch := make(chan interface{})
	i := 0
	var docs []*Document
	go func() {
		decoder := xml.NewDecoder(bufio.NewReader(r))
		for {
			t, _ := decoder.Token()
			if t == nil {
				break
			}
			switch se := t.(type) {
			case xml.StartElement:
				if se.Name.Local == "Document" {
					doc := new(Document)
					err := decoder.DecodeElement(&doc, &se)
					if err != nil {
						log.Fatal(err)
					}
					i++
					docs = append(docs, doc)
					if i == BatchSize {
						ch <- NewBatch(docs)
						docs = docs[:0]
						i = 0
					}
				}
			}
		}
		ch <- NewBatch(docs)
		close(ch)
	}()
	return ch, nil
}

func (doc Document) Date() (time.Time, error) {
	raw := strings.TrimSpace(RawDateReplacer.Replace(doc.RawDate))
	if len(raw) > 8 {
		raw = raw[:8]
	}
	return time.Parse("20060102", raw)
}

func (doc Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error
	log.Println(doc.Authors)
	output := finc.NewIntermediateSchema()
	output.ArticleTitle = strings.TrimSpace(doc.Title)
	output.ISSN = append(output.ISSN, strings.TrimSpace(doc.ISSN))
	output.Issue = strings.TrimSpace(doc.Issue)
	output.JournalTitle = strings.TrimSpace(doc.PublicationTitle)
	output.Volume = strings.TrimSpace(doc.Volume)

	output.Date, err = doc.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	return output, nil
}
