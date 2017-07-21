// Given a set of API responses from crossref, generate a file that contains only
// the latest version of a record, determined by DOI and deposit date.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"io/ioutil"

	"github.com/miku/clam"
	"github.com/miku/parallel"
	"github.com/miku/span/formats/crossref"
	"github.com/reusee/mmh3"
)

const (
	tmpPrefix = "span-crossref-snapshot-"
)

var (
	batchSize = flag.Int("b", 20, "batch size")
	cacheBase = filepath.Join(UserHomeDir(), ".cache/span-crossref-snapshot")
)

// UserHomeDir returns the home of the user.
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// HashReader returns the sha256 hex digest of the given reader.
func HashReader(r io.ReadSeeker) (hexdigest string, err error) {
	hasher := mmh3.New32()
	if _, err = io.Copy(hasher, r); err != nil {
		return
	}
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return
	}
	return fmt.Sprintf("mmh3-%x", hasher.Sum(nil)), nil
}

// WriteFields writes a variable number of fields as tab separated values into a writer.
func WriteFields(w io.Writer, s ...string) (int, error) {
	return io.WriteString(w, fmt.Sprintf("%s\n", strings.Join(s, "\t")))
}

// SetupProcessor creates a new processor. Here, we extract the interesting bits
// from the given file and write them as tabular values into the given writer.
func SetupProcessor(f *os.File, w io.Writer) *parallel.Processor {
	// reduceDocs is our transformation function.
	reduceDocs := func(b []byte) ([]byte, error) {
		var resp crossref.BulkResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		for i, doc := range resp.Message.Items {
			date, err := doc.Deposited.Date()
			if err != nil {
				return nil, err
			}
			isodate := date.Format("2006-01-02")
			if _, err := WriteFields(&buf, f.Name(), strconv.Itoa(i), isodate, doc.DOI); err != nil {
				return nil, err
			}
		}
		return buf.Bytes(), nil
	}
	p := parallel.NewProcessor(bufio.NewReader(f), w, reduceDocs)
	p.BatchSize = *batchSize
	return p
}

func main() {
	// First the filename, DOI and deposited date are extracted into a temporary file.
	// This file is sorted by DOI and deposited date, only the latest date is kept.
	// Then, for each file extract only the newest records (must keep a list of DOI in
	// memory or maybe in an embedded key value store, say bolt).
	// ...
	// For each file (sha), keep the extracted list compressed and cached at
	// ~/.cache/span-crossref-snapshot/. Also, keep a result cache for a set of files.

	flag.Parse()

	// Setup cache directory.
	cacheDir := filepath.Join(cacheBase, "tables")
	if _, err := os.Stat(cacheDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cacheDir, 0755); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}
	log.Println(cacheDir)

	// Collect filenames of extracted files here.
	var files []string

	for _, filename := range flag.Args() {

		log.Printf(filename)

		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		hexdigest, err := HashReader(f)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(hexdigest)

		cacheFile := filepath.Join(cacheDir, hexdigest)

		if _, err := os.Stat(cacheFile); err != nil {
			if os.IsNotExist(err) {
				tf, err := ioutil.TempFile("", tmpPrefix)
				if err != nil {
					log.Fatal(err)
				}
				bw := bufio.NewWriter(tf)
				if err := SetupProcessor(f, bw).Run(); err != nil {
					log.Fatal(err)
				}
				if err := bw.Flush(); err != nil {
					log.Fatal(err)
				}
				if err := os.Rename(tf.Name(), cacheFile); err != nil {
					log.Fatal(err)
				}
			}
		}
		files = append(files, cacheFile)
	}

	// Concatenate files. Sort by DOI (3), then date reversed (2); then unique by DOI
	// (3). Should keep the entry of the last update (filename, document date, DOI).
	t := "cat {{ files }} | LC_ALL=C sort -S25% -k3,3 -rk2,2 | LC_ALL=C sort -S25% -k3,3 -u > {{ output }}"

	output, err := clam.RunOutput(t, clam.Map{"files": strings.Join(files, " ")})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(output)
}
