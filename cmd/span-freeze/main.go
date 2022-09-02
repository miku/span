// Freeze file containing urls along with the content of all urls. Frozen file
// will be a zip file, containing something like:
//
//     /blob
//     /mapping.json
//     /files/<sha1 of url>
//     /files/<sha1 of url>
//     /files/...
//
//     $ curl -s https://queue.acm.org/ | span-freeze -b -o acm.zip
package main

import (
	"archive/zip"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/segmentio/encoding/json"

	"github.com/dchest/safefile"
	log "github.com/sirupsen/logrus"

	"github.com/miku/span"
	"mvdan.cc/xurls"
)

const (
	NameBlob    = "blob"
	NameMapping = "mapping.json"
	NameDir     = "files"
)

var (
	output      = flag.String("o", "", "output file")
	bestEffort  = flag.Bool("b", false, "report errors but do not stop")
	showVersion = flag.Bool("v", false, "prints current program version")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	if *output == "" {
		log.Fatal("output file required")
	}

	file, err := safefile.Create(*output, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	w := zip.NewWriter(file)

	comment := fmt.Sprintf(`Freeze-Date: %s`, time.Now().Format(time.RFC3339))
	if err := w.SetComment(comment); err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	f, err := w.Create(NameBlob)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(b); err != nil {
		log.Fatal(err)
	}

	urls := xurls.Strict.FindAllString(string(b), -1)

	seen := make(map[string]bool)
	var unique []string

	for _, u := range urls {
		u = strings.TrimSpace(u)
		if _, ok := seen[u]; !ok {
			unique = append(unique, u)
			seen[u] = true
		}
	}

	// Not necessary, but keep an additional mapping to simplify reading later.
	mapping := make(map[string]string)

	for i, u := range unique {
		if !strings.HasPrefix(u, "http") {
			log.Printf("skip: %s", u)
			continue
		}
		h := sha1.New()
		h.Write([]byte(u))
		name := fmt.Sprintf("%s/%x", NameDir, h.Sum(nil))

		mapping[u] = name

		resp, err := http.Get(u)
		if err != nil || resp.StatusCode >= 400 {
			if *bestEffort {
				log.Printf("[%04d %s] %v", i, u, err)
				continue
			} else {
				log.Fatalf("failed to fetch resource (%d): err", resp.StatusCode)
			}
		}
		defer resp.Body.Close()

		f, err := w.Create(name)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(f, resp.Body); err != nil {
			log.Fatal(err)
		}
		log.Printf("[%04d %s] %s", i, name, u)
	}

	f, err = w.Create(NameMapping)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(f).Encode(mapping); err != nil {
		log.Fatal(err)
	}
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}
	if err := file.Commit(); err != nil {
		log.Fatal(err)
	}
}
