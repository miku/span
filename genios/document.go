package genios

import (
	"bufio"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/assetutil"
	"github.com/miku/span/finc"
)

const (
	SourceID   = "48"
	BatchSize  = 2000
	Format     = "ElectronicArticle"
	Collection = "Genios"
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
	RawAuthors       []string `xml:"Authors>Author"`
	Language         string   `xml:"Language"`
	Abstract         string   `xml:"Abstract"`
	Group            string   `xml:"x-group"`
	Descriptors      string   `xml:"Descriptors>Descriptor"`
}

var (
	RawDateReplacer = strings.NewReplacer(`"`, "", "\n", "", "\t", "")
	collections     = assetutil.MustLoadStringMap("assets/genios/collections.json")
)

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
// TODO(miku): abstract this away (and in the other sources as well)
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

func (doc Document) URL() string {
	return fmt.Sprintf("https://www.genios.de/document/%s__%s/", strings.TrimSpace(doc.Source), strings.TrimSpace(doc.ID))
}

func IsNN(s string) bool {
	return strings.ToLower(strings.TrimSpace(s)) == "n.n."
}

func (doc Document) Authors() []string {
	var authors []string
	for _, v := range doc.RawAuthors {
		fields := strings.Split(v, ";")
		for _, f := range fields {
			if !IsNN(f) {
				authors = append(authors, strings.TrimSpace(f))
			}
		}
	}
	return authors
}

func (doc Document) RecordID() string {
	return fmt.Sprintf("ai-%s-%s", SourceID, base64.StdEncoding.EncodeToString([]byte(doc.ID)))
}

func (doc Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error
	output := finc.NewIntermediateSchema()

	for _, author := range doc.Authors() {
		output.Authors = append(output.Authors, finc.Author{Name: author})
	}

	output.URL = append(output.URL, doc.URL())

	if !IsNN(doc.Abstract) {
		output.Abstract = strings.TrimSpace(doc.Abstract)
	}

	if !IsNN(doc.Title) {
		output.ArticleTitle = strings.TrimSpace(doc.Title)
	}

	if !IsNN(doc.ISSN) {
		output.ISSN = append(output.ISSN, strings.TrimSpace(doc.ISSN))
	}

	if !IsNN(doc.Issue) {
		output.Issue = strings.TrimSpace(doc.Issue)
	}

	if !IsNN(doc.PublicationTitle) {
		output.JournalTitle = strings.TrimSpace(doc.PublicationTitle)
	}

	if !IsNN(doc.Volume) {
		output.Volume = strings.TrimSpace(doc.Volume)
	}

	output.RecordID = doc.RecordID()
	output.SourceID = SourceID
	output.Format = Format
	output.MegaCollection = fmt.Sprintf("Genios (%s)", collections[doc.Group])

	output.Date, err = doc.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	return output, nil
}
