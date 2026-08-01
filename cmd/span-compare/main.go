// span-compare renders a table with ISIL/SID counts of two indices side by
// side. It can parse the hidden whatislive endpoint to find live, non-live
// pairs.
//
//	$ span-compare -a 10.1.1.7:8085/solr/biblio -b 10.1.1.15:8085/solr/biblio
//
// Use hidden whatislive endpoint and render textile (for redmine):
//
//	$ span-compare -e -t
package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/miku/span/internal/cmd/compare"
	"github.com/miku/span/xio"
)

// TODO: move to XDG
var defaultConfigPath = path.Join(xio.UserHomeDir(), ".config/span/span.json")

var (
	amslLiveServer   = flag.String("amsl", "", "url to live amsl api for ad-hoc source names, e.g. https://example.technology")
	liveServer       = flag.String("a", "http://localhost:8983/solr/biblio", "live server location")
	nonliveServer    = flag.String("b", "http://localhost:8983/solr/biblio", "non-live server location")
	whatIsLive       = flag.Bool("e", false, "use whatislive.url to determine live and non live servers")
	liveLinkTemplate = flag.String("tl", compare.DefaultLiveLinkTemplate,
		"live link template for source (for focus institution)")
	spanConfigFile   = flag.String("span-config", defaultConfigPath, "for whatislive.url")
	textile          = flag.Bool("t", false, "emit textile")
	focusInstitution = flag.String("emph", "DE-15", "emphasize institution in textile output")
)

func main() {
	flag.Parse()

	cfg := compare.Config{
		AMSLLiveServer:   *amslLiveServer,
		LiveServer:       *liveServer,
		NonliveServer:    *nonliveServer,
		WhatIsLive:       *whatIsLive,
		LiveLinkTemplate: *liveLinkTemplate,
		SpanConfigFile:   *spanConfigFile,
		Textile:          *textile,
		FocusInstitution: *focusInstitution,
	}
	if err := compare.Run(cfg, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
