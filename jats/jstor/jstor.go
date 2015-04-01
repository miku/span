package jstor

import (
	"bufio"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/miku/span"
	"github.com/miku/span/finc"
	"github.com/miku/span/jats"
)

const (
	// SourceID for internal bookkeeping.
	SourceID = "52"
	// SourceName for finc.mega_collection.
	SourceName = "JSTOR"
	// Format for intermediate schema.
	Format = "ElectronicArticle"
	// Batchsize number of documents per batch.
	BatchSize = 2000
)

// Jstor source.
type Jstor struct{}

// Article with extras for this source.
type Article struct {
	jats.Article
}

// NewBatch wraps up a new batch for channel com.
func NewBatch(docs []*Article) span.Batcher {
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
func (s Jstor) Iterate(r io.Reader) (<-chan interface{}, error) {
	ch := make(chan interface{})
	i := 0
	var docs []*Article
	go func() {
		decoder := xml.NewDecoder(bufio.NewReader(r))
		for {
			t, _ := decoder.Token()
			if t == nil {
				break
			}
			switch se := t.(type) {
			case xml.StartElement:
				if se.Name.Local == "article" {
					doc := new(Article)
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

// Identifiers returns the doi and the dependent url and recordID in a struct.
// Records from this source do not need a DOI necessarily.
func (article *Article) Identifiers() (jats.Identifiers, error) {
	doi, _ := article.DOI()
	locator := article.Front.Article.SelfURI.Value
	enc := fmt.Sprintf("ai-%s-%s", SourceID, base64.URLEncoding.EncodeToString([]byte(locator)))
	recordID := strings.TrimRight(enc, "=")
	return jats.Identifiers{DOI: doi, URL: locator, RecordID: recordID}, nil
}

// Authors returns the authors as slice.
func (article *Article) Authors() []finc.Author {
	var authors []finc.Author
	group := article.Front.Article.ContribGroup
	for _, contrib := range group.Contrib {
		if contrib.Type != "author" {
			continue
		}
		authors = append(authors, finc.Author{
			LastName:  contrib.StringName.Surname.Value,
			FirstName: contrib.StringName.GivenNames.Value})
	}
	return authors
}

// ToInternalSchema converts an article into an internal schema.
func (article *Article) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output, err := article.Article.ToIntermediateSchema()
	if err != nil {
		return output, err
	}

	ids, err := article.Identifiers()
	if err != nil {
		return output, err
	}
	output.DOI = ids.DOI
	output.RecordID = ids.RecordID
	output.URL = append(output.URL, ids.URL)

	output.Authors = article.Authors()
	output.Format = Format
	output.MegaCollection = SourceName
	output.SourceID = SourceID

	var normalized []string
	for _, issn := range output.ISSN {
		if len(issn) == 8 && !strings.Contains(issn, "-") {
			normalized = append(normalized, fmt.Sprintf("%s-%s", issn[:4], issn[4:]))
		}
	}
	output.ISSN = normalized

	return output, nil
}
