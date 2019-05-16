package span

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func filenames(fis []os.FileInfo) (result []string) {
	for _, fi := range fis {
		result = append(result, fi.Name())
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
	var payload = make(map[string]interface{})
	if err := dec.Decode(&payload); err != nil {
		t.Errorf("could not decode JSON: %v", err)
	}
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Errorf("could not read dir: %v", err)
	}
	want := []string{"blob", "files", "mapping.json"}
	if !reflect.DeepEqual(filenames(fis), want) {
		t.Errorf("want %v, got %v", want, filenames(fis))
	}
}
