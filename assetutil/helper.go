package assetutil

import (
	"encoding/json"
	"log"
	"regexp"

	"github.com/miku/span/container"
)

// RegexpMapEntry maps a regex pattern to a string value.
type RegexpMapEntry struct {
	Pattern *regexp.Regexp
	Value   string
}

// RegexpMap holds a list of entries, which contain pattern string pairs.
type RegexpMap struct {
	Entries []RegexpMapEntry
}

// Lookup tries to match a given string against patterns. If none of the
// matterns match, return a given default.
func (r RegexpMap) Lookup(s, def string) string {
	for _, entry := range r.Entries {
		if entry.Pattern.MatchString(s) {
			return entry.Value
		}
	}
	return def
}

// LoadRegexpMap loads the content of a given asset path into a RegexpMap. It
// will bail out, if the asset path is not found and will panic if the
// patterns cannot be compiled.
func LoadRegexpMap(ap string) RegexpMap {
	b, err := Asset(ap)
	if err != nil {
		log.Fatal(err)
	}
	d := make(map[string]string)
	err = json.Unmarshal(b, &d)
	if err != nil {
		log.Fatal(err)
	}
	remap := RegexpMap{}
	for k, v := range d {
		remap.Entries = append(remap.Entries, RegexpMapEntry{Pattern: regexp.MustCompile(k), Value: v})
	}
	return remap
}

// LoadStringMap loads a JSON file from an asset path and parses it into a
// container.StringMap. This function will halt the world, if it is called
// with an invalid argument.
func LoadStringMap(ap string) container.StringMap {
	b, err := Asset(ap)
	if err != nil {
		log.Fatal(err)
	}
	d := make(map[string]string)
	err = json.Unmarshal(b, &d)
	if err != nil {
		log.Fatal(err)
	}
	return container.StringMap(d)
}

// LoadStringSliceMap loads a JSON file from an asset path and parses it into
// a container.StringSliceMap. This function will halt the world, if it is
// called with an invalid argument.
func LoadStringSliceMap(ap string) container.StringSliceMap {
	b, err := Asset(ap)
	if err != nil {
		log.Fatal(err)
	}
	d := make(map[string][]string)
	err = json.Unmarshal(b, &d)
	if err != nil {
		log.Fatal(err)
	}
	return container.StringSliceMap(d)
}
