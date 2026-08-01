// Freeze file containing urls along with the content of all urls into a zip
// file. Supports two modes:
//
// Legacy mode (default): reads a blob from stdin, extracts URLs, downloads
// them, and packages everything into a zip file.
//
// FOLIO mode (-f): fetches metadata collections from FOLIO API, builds a
// span-tag compatible filter configuration, downloads referenced files, and
// packages everything into a same zip format.
//
// Output zip structure:
//
//	/blob          original input or generated filterconfig JSON
//	/mapping.json  URL to local path mapping
//	/files/<sha1>  downloaded content
//
//	$ curl -s https://queue.acm.org/ | span-freeze -b -o acm.zip
//	$ OKAPI_TOKEN=xxx span-freeze -f -okapi-url https://... -o folio.zip
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/dchest/safefile"
	"github.com/miku/span"
	freezecmd "github.com/miku/span/internal/cmd/freeze"
)

var (
	output      = flag.String("o", "", "output file")
	bestEffort  = flag.Bool("b", false, "report errors but do not stop")
	showVersion = flag.Bool("v", false, "prints current program version")
	useFolio    = flag.Bool("f", false, "use FOLIO API instead of stdin")
	noProxy     = flag.Bool("no-proxy", false, "ignore system proxy settings")
	okapiURL    = flag.String("okapi-url", os.Getenv("OKAPI_URL"), "OKAPI base URL (env: OKAPI_URL)")
	tenant      = flag.String("tenant", "de15", "FOLIO tenant")
	limit       = flag.Int("limit", 100000, "API pagination limit")
	expand      = flag.String("expand", "", "JSON or file mapping meta-ISILs to lists of ISILs to expand into")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-freeze [options] -o output.zip\n\n")
		fmt.Fprintf(os.Stderr, "Freezes a filter configuration and every file it references into a single zip,\n")
		fmt.Fprintf(os.Stderr, "so a tagging run can be reproduced offline. Reads a URL-bearing blob from stdin\n")
		fmt.Fprintf(os.Stderr, "(legacy) or builds the config from the FOLIO API (-f).\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}
	if *output == "" {
		log.Fatal("output file required")
	}
	if *useFolio {
		cfg := freezecmd.FolioConfig{
			OkapiURL: *okapiURL,
			Tenant:   *tenant,
			Token:    os.Getenv("OKAPI_TOKEN"),
			Limit:    *limit,
			Expand:   *expand,
			NoProxy:  *noProxy,
		}
		if err := freezecmd.RunFolio(cfg, *output); err != nil {
			log.Fatal(err)
		}
		return
	}

	var httpClient *http.Client
	if *noProxy {
		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy:                 nil,
				DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
			},
		}
	} else {
		httpClient = http.DefaultClient
	}

	file, err := safefile.Create(*output, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	cfg := freezecmd.LegacyConfig{
		BestEffort: *bestEffort,
		Expand:     *expand,
	}
	if err := freezecmd.RunLegacy(cfg, httpClient, os.Stdin, file); err != nil {
		log.Fatal(err)
	}
	if err := file.Commit(); err != nil {
		log.Fatal(err)
	}
}
