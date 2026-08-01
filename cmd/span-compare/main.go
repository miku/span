// span-compare renders a table with ISIL/SID counts of two indices side by
// side.
//
//	$ span-compare -a 10.1.1.7:8085/solr/biblio -b 10.1.1.15:8085/solr/biblio
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span/internal/cmd/compare"
)

var (
	amslLiveServer   = flag.String("amsl", "", "url to live amsl api for ad-hoc source names, e.g. https://example.technology")
	liveServer       = flag.String("a", "http://localhost:8983/solr/biblio", "live server location")
	nonliveServer    = flag.String("b", "http://localhost:8983/solr/biblio", "non-live server location")
	liveLinkTemplate = flag.String("tl", compare.DefaultLiveLinkTemplate,
		"live link template for source (for focus institution)")
	textile          = flag.Bool("t", false, "emit textile")
	focusInstitution = flag.String("emph", "DE-15", "emphasize institution in textile output")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: span-compare [options]\n\n")
		fmt.Fprintf(os.Stderr, "Renders a table of ISIL/SID record counts for two Solr indices side by side.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg := compare.Config{
		AMSLLiveServer:   *amslLiveServer,
		LiveServer:       *liveServer,
		NonliveServer:    *nonliveServer,
		LiveLinkTemplate: *liveLinkTemplate,
		Textile:          *textile,
		FocusInstitution: *focusInstitution,
	}
	if err := compare.Run(cfg, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
