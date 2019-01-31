// Stitch together a discovery like API response.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sethgrid/pester"
	log "github.com/sirupsen/logrus"
)

var (
	live       = flag.String("live", "", "AMSL live base url")
	staging    = flag.String("staging", "", "AMSL staging base url")
	cached     = flag.Bool("cached", false, "use cached files (metadata_usage.json, holdingsfiles.json, ...)")
	allowEmpty = flag.Bool("allow-empty", false, "allow empty responses from api")
	verbose    = flag.Bool("verbose", false, "be verbose")
)

// Doc contains keys and values.
type Doc map[string]string

// MatchByKey return true, if both docs contain the same value at key.
func (doc Doc) MatchByKey(other Doc, key string) bool {
	var (
		v, w string
		ok   bool
	)
	if v, ok = doc[key]; !ok {
		return false
	}
	if w, ok = other[key]; !ok {
		return false
	}
	return v == w
}

// fetchLink fetches a link and stores the result in a temporary file, which
// the caller should clean up. Returns the name of the temporary file.
func fetchLink(link string) (filename string, err error) {
	log.Println(link)
	resp, err := pester.Get(link)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	f, err := ioutil.TempFile("", "span-")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}
	log.Printf("%s saved at %s", link, f.Name())
	return f.Name(), nil
}

// loadDocs loads api response items either from filename or link.
func loadDocs(location string) (docs []Doc, err error) {
	var filename = location
	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https") {
		if filename, err = fetchLink(location); err != nil {
			return
		}
		defer os.Remove(filename)
	}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	docs = make([]Doc, 0)
	if err := json.Unmarshal(b, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func main() {
	flag.Parse()

	fetchlist := map[string]string{
		"mu": fmt.Sprintf("%s/outboundservices/list?do=metadata_usage", *live),
		"hc": fmt.Sprintf("%s/outboundservices/list?do=holdings_file_concat", *staging),
		"hf": fmt.Sprintf("%s/outboundservices/list?do=holdingsfiles", *live),
		"cf": fmt.Sprintf("%s/outboundservices/list?do=contentfiles", *live),
	}
	if *cached {
		fetchlist = map[string]string{
			"mu": "metadata_usage.json",
			"hc": "holdings_file_concat.json",
			"hf": "holdingsfiles.json",
			"cf": "contentfiles.json",
		}
	}

	var (
		mu, hc, hf, cf []Doc
		err            error
	)

	if mu, err = loadDocs(fetchlist["mu"]); err != nil {
		log.Fatal(err)
	}
	if hc, err = loadDocs(fetchlist["hc"]); err != nil {
		log.Fatal(err)
	}
	if hf, err = loadDocs(fetchlist["hf"]); err != nil {
		log.Fatal(err)
	}
	if cf, err = loadDocs(fetchlist["cf"]); err != nil {
		log.Fatal(err)
	}

	log.Println(len(mu))
	log.Println(len(hc))
	log.Println(len(hf))
	log.Println(len(cf))

	if len(mu) == 0 || len(hc) == 0 || len(hf) == 0 || len(cf) == 0 {
		if !*allowEmpty {
			log.Fatalf("at least one empty response in %s", fetchlist)
		}
		log.Println("at least one API endpoint returned an empty response")
	}

	updates := make([]Doc, 0)

	// Iterate over items in metadata usage (seems to be the most complete).
	for i, doc := range mu {
		if i%10000 == 0 {
			log.Printf("%d/%d", i, len(mu))
		}
		if doc["megaCollection"] == "" {
			log.Printf("skipping empty megaCollection in %d", i)
			continue
		}
		// Content files.
		for _, other := range cf {
			if !doc.MatchByKey(other, "megaCollection") {
				continue
			}
			// Update from matching document.
			for k, v := range other {
				if _, ok := doc[k]; !ok {
					doc[k] = v
				}
			}
		}

		// If the item is not shipped in HC (HoldingsFile Concat) API response,
		// it's a no.
		doc["evaluateHoldingsFileForLibrary"] = "no"

		// Holding files concatenated, if there is a matching entry here, we
		// need to evaluate the file.
		for _, other := range hc {
			if !doc.MatchByKey(other, "megaCollection") {
				continue
			}
			if !doc.MatchByKey(other, "ISIL") {
				continue
			}
			// Update from matching document, if collection and ISIL are matching.
			for k, v := range other {
				if _, ok := doc[k]; !ok {
					doc[k] = v
				}
			}
			if doc["ISIL"] == "DE-14" {
				log.Printf("DE-14 match: %s, other: %s", doc, other)
			}
			doc["evaluateHoldingsFileForLibrary"] = "yes"
		}

		if v, ok := doc["contentFileURI"]; ok {
			if strings.HasPrefix(v, "http://amsl") {
				doc["linkToContentFile"] = fmt.Sprintf(
					"%s/OntoWiki/files/get?setResource=%s", *live, v)
			}
		}

		// If we do not need holdings file, just emit the doc as is and
		// continue.
		if doc["evaluateHoldingsFileForLibrary"] == "no" {
			updates = append(updates, doc)
			if *verbose {
				log.Printf("claim without holding file for %s and %s", doc["ISIL"], doc["sourceID"])
			}
			continue
		}

		// Holding files. For each holding file, we create a new doc.
		for _, other := range hf {
			if !doc.MatchByKey(other, "ISIL") {
				continue
			}
			// Copy existing fields.
			ndoc := make(Doc)
			for k, v := range doc {
				ndoc[k] = v
			}
			// Update from matching document.
			for k, v := range other {
				if k == "LinkToFile" {
					k = "linkToHoldingsFile"
				}
				if k == "DokumentURI" {
					k = "linkToHoldingsFile"
					v = fmt.Sprintf("%s/OntoWiki/files/get?setResource=%s", *live, v)
				}
				ndoc[k] = v
			}
			// Pad empty fields.
			ensure := []string{
				"holdingsFileLabel",
				"holdingsFileURI",
				"linkToHoldingsFile",
				"contentFileLabel",
				"contentFileURI",
				"linkToContentFile",
				"externalLinkToContentFile",
			}
			for _, key := range ensure {
				if _, ok := doc[key]; !ok {
					doc[key] = ""
				}
			}
			updates = append(updates, ndoc)
		}
	}

	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()
	if err := json.NewEncoder(bw).Encode(updates); err != nil {
		log.Fatal(err)
	}
}
