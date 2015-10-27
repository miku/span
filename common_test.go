package span

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/miku/span/finc"
)

// mockSchema is a toy schema.
type mockSchema struct {
	Name string `json:"name"`
}

// ToIntermediateSchema is a toy converter.
func (s mockSchema) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	is := finc.NewIntermediateSchema()
	is.ArticleTitle = s.Name
	return is, nil
}

// mockDecoder is a toy decoder.
func mockDecoder(s string) (Importer, error) {
	var target mockSchema
	err := json.Unmarshal([]byte(s), &target)
	return target, err
}

// unroll turns a chan of importer batches into a slice of importers.
func unroll(ch chan []Importer) []Importer {
	var docs []Importer
	for batch := range ch {
		for _, doc := range batch {
			docs = append(docs, doc)
		}
	}
	return docs
}

// mockReader is a sample JSON input file.
var mockReader = strings.NewReader(`
			{"name": "Anna"}

			{"name": "Beryl"}
			{"name": "Coran"}
`)

// TestFromJSONSize reading a JSON file into span.Importer values.
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
			// if size is 1, the goroutines start to randomize the output order
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

type mockSchemaReader struct {
	current int
	size    int
}

var sampleLine = []byte(`{"name": "TESTVALUE"}\n`)

func (r *mockSchemaReader) Read(p []byte) (n int, err error) {
	if r.current == r.size {
		return 0, io.EOF
	}
	n = len(p)
	if n == 0 {
		return 0, nil
	}
	n = copy(p, sampleLine)
	r.current += 1
	return n, nil
}

func BenchmarkMockSchemaReader10k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rdr := &mockSchemaReader{size: 10000}
		for {
			p := make([]byte, 24)
			_, err := rdr.Read(p)
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func BenchmarkFromJSONSize10kDocs10Batch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := FromJSONSize(&mockSchemaReader{size: 10000}, mockDecoder, 10)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkFromJSONSize10kDocs100Batch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := FromJSONSize(&mockSchemaReader{size: 10000}, mockDecoder, 100)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkFromJSONSize10kDocs1kBatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := FromJSONSize(&mockSchemaReader{size: 10000}, mockDecoder, 1000)
		if err != nil {
			b.Error(err)
		}
	}
}
