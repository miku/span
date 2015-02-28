package span

import (
	"fmt"
	"strings"

	"github.com/miku/span/holdings"
)

// Version of span
const Version = "0.1.7"

// Tagger can set tags, e.g. institutions (list of ISILs)
type Tagger interface {
	SetTags([]string)
}

// ConverterCoverer should be implemented by
// formats that can convert themselves to
// other format and which can determine coverage
// information, given holdings information.
type ConverterCoverer interface {
	ConvertFormat(string) Tagger
	Institutions(holdings.IsilIssnHolding) []string
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
