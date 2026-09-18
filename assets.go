package span

import (
	"embed"
	"encoding/json"
)

//go:embed assets
var Static embed.FS

// MustLoadAsset decodes the JSON file at path in Static into a value of type
// T. Assets are compiled into the binary, so any error is a programming error
// and panics.
func MustLoadAsset[T any](path string) T {
	b, err := Static.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		panic(err)
	}
	return v
}
