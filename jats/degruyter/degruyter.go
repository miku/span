package degruyter

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
	SourceID = "50"
	// SourceName for finc.mega_collection.
	SourceName = "DeGruyter SSH"
	// Format for intermediate schema.
	Format = "ElectronicArticle"
	// Batchsize number of documents per batch.
	BatchSize = 2000
)

// DeGruyter source.
type DeGruyter struct{}

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
func (s DeGruyter) Iterate(r io.Reader) (<-chan interface{}, error) {
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
// It is an error, if there is no DOI.
func (article *Article) Identifiers() (jats.Identifiers, error) {
	var ids jats.Identifiers
	doi, err := article.DOI()
	if err != nil {
		return ids, err
	}
	locator := fmt.Sprintf("http://dx.doi.org/%s", doi)
	enc := fmt.Sprintf("ai-%s-%s", SourceID, base64.URLEncoding.EncodeToString([]byte(locator)))
	recordID := strings.TrimRight(enc, "=")
	return jats.Identifiers{DOI: doi, URL: locator, RecordID: recordID}, nil
}

// ToInternalSchema converts a jats article into an internal schema.
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

	output.Format = Format
	output.MegaCollection = SourceName
	output.SourceID = SourceID

	return output, nil
}
