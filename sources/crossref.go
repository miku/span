package sources

import (
	"encoding/json"
	"io"

	"github.com/miku/span"
	"github.com/miku/span/crossref"
)

type Crossref struct{}

func (c Crossref) Iterate(r io.Reader) (<-chan []span.Importer, error) {
	return span.FromJSON(r, func(s string) (span.Importer, error) {
		doc := new(crossref.Document)
		err := json.Unmarshal([]byte(s), doc)
		return doc, err
	})
}
