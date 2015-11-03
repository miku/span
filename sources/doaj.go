package sources

import (
	"encoding/json"
	"io"

	"github.com/miku/span"
	"github.com/miku/span/doaj"
)

type DOAJ struct{}

func (s DOAJ) Iterate(r io.Reader) (<-chan []span.Importer, error) {
	return span.FromJSON(r, func(s string) (span.Importer, error) {
		resp := new(doaj.Response)
		err := json.Unmarshal([]byte(s), resp)
		if err != nil {
			return resp.Source, err
		}
		resp.Source.Type = resp.Type
		return resp.Source, nil
	})
}
