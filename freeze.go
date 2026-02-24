package span

import (
	"archive/zip"
	"bytes"
	"io"
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
	if dir, err = os.MkdirTemp("", "span-tag-unfreeze-"); err != nil {
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
	if b, err = os.ReadFile(blob); err != nil {
		return
	}
	// Parse the blob as JSON, replace URLs in the tree, and re-serialize.
	// This is more robust than byte-level replacement, which can fail when
	// the JSON encoder uses different escaping (e.g. \u0026 for &) than
	// Go's %q format.
	var parsed any
	if err = json.Unmarshal(b, &parsed); err != nil {
		return
	}
	parsed = replaceURLs(parsed, mappings, dir)
	if b, err = json.MarshalIndent(parsed, "", "  "); err != nil {
		return
	}
	if err = os.WriteFile(blob, b, 0777); err != nil {
		return
	}
	return dir, blob, nil
}

// replaceURLs recursively walks a parsed JSON value and replaces any string
// that matches a URL in the mappings with its corresponding file:// path.
func replaceURLs(v any, mappings map[string]string, dir string) any {
	switch val := v.(type) {
	case string:
		if file, ok := mappings[val]; ok {
			return "file://" + filepath.Join(dir, file)
		}
		return val
	case map[string]any:
		result := make(map[string]any, len(val))
		for k, child := range val {
			result[k] = replaceURLs(child, mappings, dir)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, child := range val {
			result[i] = replaceURLs(child, mappings, dir)
		}
		return result
	default:
		return val
	}
}
