package assetutil

import (
	"encoding/json"
	"log"

	"github.com/miku/span/container"
)

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
