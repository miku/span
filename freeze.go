package span

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// UnfreezeFilterConfig takes the name of a zipfile (from span-freeze) and
// returns the path to the filterconfig. All URLs in the filterconfig have then
// been replaced by absolute path on the file system. Assumes default values
// for mapping.json and content file (blob). Cleanup of temporary directory is
// responsibility of caller.
func UnfreezeFilterConfig(frozenfile string) (string, string, error) {
	dir, err := ioutil.TempDir("", "span-tag-unfreeze-")
	if err != nil {
		return dir, "", err
	}
	r, err := zip.OpenReader(frozenfile)
	if err != nil {
		return dir, "", err
	}
	defer r.Close()

	var mappings = make(map[string]string)
	if err := os.MkdirAll(filepath.Join(dir, "files"), 0777); err != nil {
		return dir, "", err
	}

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return dir, "", err
		}
		ff, err := os.Create(filepath.Join(dir, f.Name))
		if err != nil {
			return dir, "", err
		}
		var buf bytes.Buffer
		tr := io.TeeReader(rc, &buf)
		if _, err := io.Copy(ff, tr); err != nil {
			return dir, "", err
		}
		rc.Close()
		ff.Close()

		if f.Name == "mapping.json" {
			if err := json.NewDecoder(&buf).Decode(&mappings); err != nil {
				return dir, "", err
			}
		}
	}

	// Assume dir/blob contains filterconfig with URLs.
	blobfile := filepath.Join(dir, "blob")
	b, err := ioutil.ReadFile(blobfile)
	if err != nil {
		return dir, "", err
	}

	for url, file := range mappings {
		// Too ugly?
		x := []byte(fmt.Sprintf(`"%s"`, url))
		y := []byte(fmt.Sprintf(`"%s"`, filepath.Join(dir, file)))
		b = bytes.Replace(b, x, y, -1)
		x = []byte(`{"holdings": {"urls":`)
		y = []byte(`{"holdings": {"files":`)
		b = bytes.Replace(b, x, y, -1)
	}
	if err := ioutil.WriteFile(blobfile, b, 0777); err != nil {
		return dir, "", err
	}
	return dir, blobfile, nil
}
