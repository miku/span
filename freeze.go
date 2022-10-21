package span

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	json "github.com/segmentio/encoding/json"
)

// UnfreezeFilterConfig takes the name of a zipfile (from span-freeze) and
// returns of the path the thawed filterconfig (along with the temporary
// directory and error). When this function returns, all URLs in the
// filterconfig have then been replaced by absolute path on the file system.
// Cleanup of temporary directory is responsibility of caller.
func UnfreezeFilterConfig(frozenfile string) (dir, blob string, err error) {
	var (
		r        *zip.ReadCloser
		rc       io.ReadCloser
		ff       *os.File
		mappings = make(map[string]string)
		b        []byte
	)
	if dir, err = ioutil.TempDir("", "span-tag-unfreeze-"); err != nil {
		return
	}
	if r, err = zip.OpenReader(frozenfile); err != nil {
		return
	}
	defer r.Close()
	if err = os.MkdirAll(filepath.Join(dir, "files"), 0777); err != nil {
		return
	}
	for _, f := range r.File {
		if rc, err = f.Open(); err != nil {
			return
		}
		if ff, err = os.Create(filepath.Join(dir, f.Name)); err != nil {
			return
		}
		if f.Name == "mapping.json" {
			var (
				buf bytes.Buffer
				tr  = io.TeeReader(rc, &buf)
			)
			if _, err = io.Copy(ff, tr); err != nil {
				return
			}
			if err = json.NewDecoder(&buf).Decode(&mappings); err != nil {
				return
			}
		} else {
			if _, err = io.Copy(ff, rc); err != nil {
				return
			}
		}
		if err = rc.Close(); err != nil {
			return
		}
		if err = ff.Close(); err != nil {
			return
		}
	}
	blob = filepath.Join(dir, "blob")
	if b, err = ioutil.ReadFile(blob); err != nil {
		return
	}
	for url, file := range mappings {
		value := []byte(fmt.Sprintf(`%q`, url))
		replacement := []byte(fmt.Sprintf(`"file://%s"`, filepath.Join(dir, file)))
		b = bytes.Replace(b, value, replacement, -1)
	}
	if err = ioutil.WriteFile(blob, b, 0777); err != nil {
		return
	}
	return dir, blob, nil
}
