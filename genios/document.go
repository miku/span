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

	"github.com/kapsteur/franco"
	"github.com/miku/span"
	"github.com/miku/span/assetutil"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
)

const (
	SourceID   = "48"
	BatchSize  = 2000
	Format     = "ElectronicArticle"
	Collection = "Genios"
	Genre      = "article"
	// If no abstract is found accept this number of chars from doc.Text as Abstract.
	TextAsAbstractCutoff = 2000
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
	Text             string   `xml:"Text"`
}

var (
	RawDateReplacer = strings.NewReplacer(`"`, "", "\n", "", "\t", "")
	collections     = assetutil.MustLoadStringMap("assets/genios/collections.json")
	// Restricts the possible languages for detection.
	acceptedLanguages = container.NewStringSet("deu", "eng")
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

func (doc Document) Headings() []string {
	var headings []string
	fields := strings.FieldsFunc(doc.Descriptors, func(r rune) bool {
		return r == ';' || r == '/'
	})
	for _, f := range fields {
		headings = append(headings, strings.TrimSpace(f))
	}
	return headings
}

func (doc Document) Date() (time.Time, error) {
	raw := strings.TrimSpace(RawDateReplacer.Replace(doc.RawDate))
	if len(raw) > 8 {
		raw = raw[:8]
	}
	return time.Parse("20060102", raw)
}

func (doc Document) SourceAndID() string {
	return fmt.Sprintf("%s__%s", strings.TrimSpace(doc.Source), strings.TrimSpace(doc.ID))
}

func (doc Document) URL() string {
	return fmt.Sprintf("https://www.wiso-net.de/document/%s", doc.SourceAndID())
}

func IsNN(s string) bool {
	return strings.ToLower(strings.TrimSpace(s)) == "n.n."
}

func (doc Document) Authors() []string {
	var authors []string
	for _, s := range doc.RawAuthors {
		fields := strings.FieldsFunc(s, func(r rune) bool {
			return r == ';' || r == '/'
		})
		for _, f := range fields {
			if !IsNN(f) {
				authors = append(authors, strings.TrimSpace(f))
			}
		}
	}
	return authors
}

func (doc Document) RecordID() string {
	enc := fmt.Sprintf("ai-%s-%s", SourceID, base64.StdEncoding.EncodeToString([]byte(doc.SourceAndID())))
	return strings.TrimRight(enc, "=")
}

// Languages returns the given and guessed languages
// found in abstract and fulltext. Note: This is slow.
// Skip detection on too short strings.
func (doc Document) Languages() []string {
	set := container.NewStringSet()

	vals := []string{doc.Title, doc.Text}

	for _, s := range vals {
		if len(s) < 20 {
			continue
		}
		lang := franco.DetectOne(s)
		if !acceptedLanguages.Contains(lang.Code) {
			continue
		}
		if lang.Code == "und" {
			continue
		}
		set.Add(lang.Code)
	}

	return set.Values()
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
	} else {
		cutoff := len(doc.Text)
		if cutoff > TextAsAbstractCutoff {
			cutoff = TextAsAbstractCutoff
		}
		output.Abstract = strings.TrimSpace(doc.Text[:cutoff])
	}

	output.ArticleTitle = strings.TrimSpace(doc.Title)
	output.JournalTitle = strings.TrimSpace(doc.PublicationTitle)

	if !IsNN(doc.ISSN) {
		output.ISSN = append(output.ISSN, strings.TrimSpace(doc.ISSN))
	}

	if !IsNN(doc.Issue) {
		output.Issue = strings.TrimSpace(doc.Issue)
	}

	if !IsNN(doc.Volume) {
		output.Volume = strings.TrimSpace(doc.Volume)
	}

	output.Format = Format
	output.Genre = Genre
	output.Languages = doc.Languages()
	output.MegaCollection = fmt.Sprintf("Genios (%s)", collections[doc.Group])
	output.RecordID = doc.RecordID()
	output.SourceID = SourceID
	output.Subjects = doc.Headings()

	output.Date, err = doc.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	return output, nil
}
