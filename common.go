package span

import (
	"fmt"
	"strings"

	"github.com/miku/span/finc"
)

// Version of span
const Version = "0.1.7"

// FincConverter for finc.Schema output format.
type FincConverter interface {
	ToSchema() (finc.Schema, error)
}

// FincSolrConverter for finc.SolrSchema output format.
type FincSolrConverter interface {
	ToSolrSchema() (finc.SolrSchema, error)
}

// ParseHoldingSpec parses a holdings flag value into a map.
func ParseHoldingSpec(s string) (map[string]string, error) {
	fields := strings.Split(s, ",")
	pathmap := make(map[string]string)
	for _, f := range fields {
		parts := strings.Split(f, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid spec: %s", f)
		}
		pathmap[parts[0]] = parts[1]

	}
	return pathmap, nil
}
