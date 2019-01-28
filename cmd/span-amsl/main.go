// Stitch together a discovery like API response.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/sethgrid/pester"
)

var (
	live       = flag.String("live", "", "AMSL live base url")
	staging    = flag.String("staging", "", "AMSL staging base url")
	cached     = flag.Bool("cached", false, "use cached files (metadata_usage.json, holdingsfiles.json, ...)")
	allowEmpty = flag.Bool("allow-empty", false, "allow empty responses from api")
)

// Doc contains keys and values.
type Doc map[string]string

// MatchByKey return true, if both docs contain the same value in the key
// field.
func (doc Doc) MatchByKey(other Doc, key string) bool {
	v, ok := doc[key]
	if !ok {
		return false
	}
	w, ok := other[key]
	if !ok {
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
	return f.Name(), nil
}

// loadDocs loads api response items either from file or link.
func loadDocs(name string) (docs []Doc, err error) {
	var filename = name
	if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https") {
		filename, err = fetchLink(name)
		if err != nil {
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
		"mu":  fmt.Sprintf("%s/outboundservices/list?do=metadata_usage", *live),
		"hfc": fmt.Sprintf("%s/outboundservices/list?do=holdings_file_concat", *staging),
		"hf":  fmt.Sprintf("%s/outboundservices/list?do=holdingsfiles", *live),
		"cf":  fmt.Sprintf("%s/outboundservices/list?do=contentfiles", *live),
	}
	if *cached {
		fetchlist = map[string]string{
			"mu":  "metadata_usage.json",
			"hfc": "holdings_file_concat.json",
			"hf":  "holdingsfiles.json",
			"cf":  "contentfiles.json",
		}
	}

	mu, err := loadDocs(fetchlist["mu"])
	if err != nil {
		log.Fatal(err)
	}
	hfc, err := loadDocs(fetchlist["hfc"])
	if err != nil {
		log.Fatal(err)
	}
	hf, err := loadDocs(fetchlist["hf"])
	if err != nil {
		log.Fatal(err)
	}
	cf, err := loadDocs(fetchlist["cf"])
	if err != nil {
		log.Fatal(err)
	}

	log.Println(len(mu))
	log.Println(len(hfc))
	log.Println(len(hf))
	log.Println(len(cf))

	if len(mu) == 0 || len(hfc) == 0 || len(hf) == 0 || len(cf) == 0 {
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

		// If the item is not shipped in HFC API response, it's a no.
		doc["evaluateHoldingsFileForLibrary"] = "no"

		// Holding files concatenated.
		for _, other := range hfc {
			if !doc.MatchByKey(other, "megaCollection") && !doc.MatchByKey(other, "ISIL") {
				continue
			}
			// Update from matching document.
			for k, v := range other {
				if _, ok := doc[k]; !ok {
					doc[k] = v
				}
			}
			doc["evaluateHoldingsFileForLibrary"] = "yes"
		}

		if v, ok := doc["contentFileURI"]; ok {
			if strings.HasPrefix(v, "http://amsl") {
				doc["linkToContentFile"] = fmt.Sprintf(
					"%s/OntoWiki/files/get?setResource=%s", *live, v)
			}
		}

		if len(hf) == 0 {
			log.Println("no holding files")
		}

		// Holding files. For each holding file, we create a new doc.
		// XXX: No every collection will need a holdings file.
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
