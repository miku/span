package doaj

import (
	"maps"
	"regexp"
	"slices"

	"github.com/miku/span"
)

// regexpMap maps patterns to values. Patterns are tried in lexical order and
// the first match wins.
type regexpMap []regexpEntry

type regexpEntry struct {
	pattern *regexp.Regexp
	value   string
}

// lookup returns the value of the first pattern matching s, or def.
func (r regexpMap) lookup(s, def string) string {
	for _, e := range r {
		if e.pattern.MatchString(s) {
			return e.value
		}
	}
	return def
}

// mustLoadRegexpMap loads a JSON object of pattern-value pairs from an asset
// path. It panics if the asset is missing or a pattern does not compile.
func mustLoadRegexpMap(path string) regexpMap {
	m := span.MustLoadAsset[map[string]string](path)
	r := make(regexpMap, 0, len(m))
	for _, k := range slices.Sorted(maps.Keys(m)) {
		r = append(r, regexpEntry{regexp.MustCompile(k), m[k]})
	}
	return r
}
