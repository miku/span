package assetutil

import (
	"encoding/json"
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

// LookupDefault tries to match a given string against patterns. If none of the
// matterns match, return a given default.
func (r RegexpMap) LookupDefault(s, def string) string {
	for _, entry := range r.Entries {
		if entry.Pattern.MatchString(s) {
			return entry.Value
		}
	}
	return def
}

// MustLoadRegexpMap loads the content of a given asset path into a RegexpMap. It
// will panic, if the asset path is not found and if the patterns found in the
// file cannot be compiled.
func MustLoadRegexpMap(ap string) RegexpMap {
	b, err := Asset(ap)
	if err != nil {
		panic(err)
	}
	d := make(map[string]string)
	err = json.Unmarshal(b, &d)
	if err != nil {
		panic(err)
	}
	remap := RegexpMap{}
	for k, v := range d {
		remap.Entries = append(remap.Entries, RegexpMapEntry{Pattern: regexp.MustCompile(k), Value: v})
	}
	return remap
}

// MustLoadStringMap loads a JSON file from an asset path and parses it into a
// container.StringMap. This function will panic, if the asset cannot be found
// or the JSON is erroneous.
func MustLoadStringMap(ap string) container.StringMap {
	b, err := Asset(ap)
	if err != nil {
		panic(err)
	}
	d := make(map[string]string)
	err = json.Unmarshal(b, &d)
	if err != nil {
		panic(err)
	}
	return container.StringMap(d)
}

// MustLoadStringSliceMap loads a JSON file from an asset path and parses it into
// a container.StringSliceMap. This function will halt the world, if it is
// called with an invalid argument.
func MustLoadStringSliceMap(ap string) container.StringSliceMap {
	b, err := Asset(ap)
	if err != nil {
		panic(err)
	}
	d := make(map[string][]string)
	err = json.Unmarshal(b, &d)
	if err != nil {
		panic(err)
	}
	return container.StringSliceMap(d)
}
