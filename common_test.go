package span

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/miku/span/finc"
)

type mockSchema struct {
	Name string `json:"name"`
}

func (s mockSchema) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	is := finc.NewIntermediateSchema()
	is.ArticleTitle = s.Name
	return is, nil
}

func mockDecoder(s string) (Importer, error) {
	var target mockSchema
	err := json.Unmarshal([]byte(s), &target)
	return target, err
}

func unroll(ch chan []Importer) []Importer {
	var docs []Importer
	for batch := range ch {
		for _, doc := range batch {
			docs = append(docs, doc)
		}
	}
	return docs
}

var mockReader = strings.NewReader(`{"name": "Anna"}
			{"name": "Beryl"}
			{"name": "Coran"}
`)

// func FromJSONSize(r io.Reader, decoder JSONDecoderFunc, size int) (chan []Importer, error) {
func TestFromJSONSize(t *testing.T) {
	var cases = []struct {
		r       io.Reader
		decoder JSONDecoderFunc
		size    int
		result  []Importer
		err     error
	}{
		{
			mockReader,
			mockDecoder,
			3,
			[]Importer{mockSchema{Name: "Anna"}, mockSchema{Name: "Beryl"}, mockSchema{Name: "Coran"}},
			nil,
		},
	}

	for _, c := range cases {
		result, err := FromJSONSize(c.r, c.decoder, c.size)
		docs := unroll(result)
		if err != c.err {
			t.Errorf("got %v, want %v", err, c.err)
		}
		if !reflect.DeepEqual(docs, c.result) {
			t.Errorf("got %v, want %v", docs, c.result)
		}
	}
}
