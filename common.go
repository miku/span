package span

import (
	"fmt"
	"strings"

	"github.com/miku/span/finc"
	"github.com/miku/span/holdings"
)

// Version of span
const Version = "0.1.7"

// SolrSchemaConverter should be implemented by
// formats that can convert themselves to
// finc solr schema and which can determine coverage
// information, given holdings information.
type SolrSchemaConverter interface {
	ToSolrSchema() (*finc.SolrSchema, error)
	Institutions(holdings.IsilIssnHolding) []string
}

// InternalSchemaConverter can output finc internal schema documents.
type InternalSchemaConverter interface {
	ToInternalSchema(*finc.InternalSchema, error)
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
