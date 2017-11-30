// freeze file containing urls along with the content of all urls.
//
//     /blob
//     /files/<sha1 of url>
//
package main

import (
	"archive/zip"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"mvdan.cc/xurls"
)

func main() {
	output := flag.String("o", "", "output file")
	flag.Parse()

	if *output == "" {
		log.Fatal("output file required")
	}

	// Output will be a zipfile.
	file, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}

	w := zip.NewWriter(file)

	seen := make(map[string]bool)
	var uniq []string

	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	f, err := w.Create("blob")
	if err != nil {
		log.Fatal(err)
	}
	f.Write(b)

	urls := xurls.Strict().FindAllString(string(b), -1)

	for _, u := range urls {
		if _, ok := seen[u]; !ok {
			uniq = append(uniq, strings.TrimSpace(u))
			seen[u] = true
		}
	}

	for _, u := range uniq {
		// Create a unique name.
		h := sha1.New()
		h.Write([]byte(u))
		name := fmt.Sprintf("files/%x", h.Sum(nil))

		resp, err := http.Get(u)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		// Create zip entry.
		f, err := w.Create(name)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(f, resp.Body); err != nil {
			log.Fatal(err)
		}
		log.Printf("%s -> %s", u, name)
	}

	if err := w.Close(); err != nil {
		log.Fatal(err)
	}
}
