package span

import (
	"os"
	"reflect"
	"testing"

	"github.com/segmentio/encoding/json"
)

func filenames(des []os.DirEntry) (result []string) {
	for _, de := range des {
		result = append(result, de.Name())
	}
	return
}

func TestUnfreezeFilterConfig(t *testing.T) {
	dir, blob, err := UnfreezeFilterConfig("fixtures/frozen.zip")
	if err != nil {
		t.Errorf("expected err nil, got %v", err)
	}
	f, err := os.Open(blob)
	if err != nil {
		t.Errorf("could not open blob file %v", err)
	}
	defer f.Close()
	defer os.RemoveAll(dir)

	dec := json.NewDecoder(f)
	var payload = make(map[string]any)
	if err := dec.Decode(&payload); err != nil {
		t.Errorf("could not decode JSON: %v", err)
	}
	fis, err := os.ReadDir(dir)
	if err != nil {
		t.Errorf("could not read dir: %v", err)
	}
	want := []string{"blob", "files", "mapping.json"}
	if !reflect.DeepEqual(filenames(fis), want) {
		t.Errorf("want %v, got %v", want, filenames(fis))
	}
}
