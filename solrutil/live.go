package solrutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

// findSolrServer looks in the config for the whatislive.url key, takes the URL
// and parses the nginx snippet for the given kind, live or nonlive. Poor man's
// service discovery (PMSD).
func findSolrServer(spanConfigFile, kind string) (string, error) {
	if _, err := os.Stat(spanConfigFile); os.IsNotExist(err) {
		return "", fmt.Errorf("missing config file: %s", spanConfigFile)
	}
	log.Printf("using span config at %s", spanConfigFile)
	f, err := os.Open(spanConfigFile)
	if err != nil {
		return "", err
	}
	var conf struct {
		WhatIsLiveURL string `json:"whatislive.url"`
	}
	if err := json.NewDecoder(f).Decode(&conf); err != nil {
		return "", err
	}
	log.Printf("querying [%s] for solr location", conf.WhatIsLiveURL)
	resp, err := http.Get(conf.WhatIsLiveURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// Find hostport in nginx snippet: upstream solr_nonlive { server 10.1.1.10:8080; }.
	p := fmt.Sprintf(`(?m)upstream[\s]*solr_%s[\s]*{[\s]*server[\s]*([0-9:.]*)[\s]*;[\s]*}`, kind)
	matches := regexp.MustCompile(p).FindSubmatch(b)
	if matches == nil || len(matches) != 2 {
		return "", fmt.Errorf("cannot find %s solr server URL in nginx snippet: %s",
			kind, string(b))
	}
	solrServer := fmt.Sprintf("%s/solr/biblio", string(matches[1]))
	return solrServer, nil
}

// FindNonliveSolrServer tries to find the URL of the current testing solr.
// There might be different way to retrieve a useable URL (configuration,
// probes). For now we use a separate configuration file, that contains the URL
// to the nginx snippet.
func FindNonliveSolrServer(spanConfigFile string) (string, error) {
	return findSolrServer(spanConfigFile, "nonlive")
}

// FindLiveSolrServer returns the current live SOLR URL.
func FindLiveSolrServer(spanConfigFile string) (string, error) {
	return findSolrServer(spanConfigFile, "live")
}
