package thieme

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/finc"
)

const (
	SourceID = "60"
	// Format   = "ElectronicArticle"
	Collection = "Thieme"
	Genre      = "article"
	batchSize  = 2000
)

type Thieme struct{}

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

func (s Thieme) Iterate(r io.Reader) (<-chan interface{}, error) {
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
				if se.Name.Local == "record" {
					doc := new(Document)
					err := decoder.DecodeElement(&doc, &se)
					if err != nil {
						log.Fatal(err)
					}
					i++
					docs = append(docs, doc)
					if i == batchSize {
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

type Document struct {
	Identifier struct {
		Value string `xml:",chardata"`
	} `xml:"header>identifier"`
	Metadata struct {
		ArticleSet struct {
			Article struct {
				Journal struct {
					Publisher struct {
						Name string `xml:",chardata"`
					} `xml:"publishername"`
					Title struct {
						Value string `xml:",chardata"`
					} `xml:"journaltitle"`
					ISSN struct {
						Value string `xml:",chardata"`
					} `xml:"issn"`
					EISSN struct {
						Value string `xml:",chardata"`
					} `xml:"e-issn"`
					Volume struct {
						Value string `xml:",chardata"`
					} `xml:"volume"`
					Issue struct {
						Value string `xml:",chardata"`
					} `xml:"issue"`
					PubDate struct {
						Year struct {
							Value string `xml:",chardata"`
						} `xml:"year"`
						Month struct {
							Value string `xml:",chardata"`
						} `xml:"month"`
						Day struct {
							Value string `xml:",chardata"`
						} `xml:"day"`
					} `xml:"pubdate"`
				} `xml:"journal"`
			} `xml:"article"`
		} `xml:"articleset"`
	} `xml:"metadata"`
}

func (doc Document) Date() (time.Time, error) {
	pd := doc.Metadata.ArticleSet.Article.Journal.PubDate
	if pd.Year.Value != "" && pd.Month.Value != "" && pd.Day.Value != "" {
		s := fmt.Sprintf("%s-%s-%s", pd.Year.Value, pd.Month.Value, pd.Day.Value)
		return time.Parse("2006-01-02", s)
	}
	if pd.Year.Value != "" && pd.Month.Value != "" {
		s := fmt.Sprintf("%s-%s-01", pd.Year.Value, pd.Month.Value)
		return time.Parse("2006-01-02", s)
	}
	if pd.Year.Value != "" {
		s := fmt.Sprintf("%s-01-01", pd.Year.Value)
		return time.Parse("2006-01-02", s)
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func (doc Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	output.RecordID = doc.Identifier.Value
	output.SourceID = SourceID
	output.MegaCollection = Collection
	output.Genre = Genre

	date, err := doc.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	output.Date = date

	journal := doc.Metadata.ArticleSet.Article.Journal
	if journal.Publisher.Name != "" {
		output.Publishers = append(output.Publishers, journal.Publisher.Name)
	}

	if journal.Title.Value == "" {
		return output, span.Skip{Reason: fmt.Sprintf("SKIP NO_JTITLE %s", output.RecordID)}
	}

	output.JournalTitle = journal.Title.Value
	output.ISSN = append(output.ISSN, journal.ISSN.Value)
	output.EISSN = append(output.EISSN, journal.ISSN.Value)
	output.Volume = journal.Volume.Value
	output.Issue = journal.Issue.Value

	return output, nil
}
