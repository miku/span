// span-oa-filter will set x.oa to true, if the given KBART file validates a record.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span"
	"github.com/miku/span/internal/cmd/oafilter"
	"github.com/miku/span/xflag"
)

var (
	showVersion      = flag.Bool("v", false, "prints current program version")
	kbartFile        = flag.String("f", "", "path to a single KBART file")
	freeContentFile  = flag.String("fc", "", "path to a .../list?do=freeContent AMSL response JSON file")
	batchsize        = flag.Int("b", 5000, "batch size")
	verbose          = flag.Bool("verbose", false, "extended output")
	batchMemoryLimit = flag.Int64("m", 209715200, "memory limit per batch")
	bestEffort       = flag.Bool("B", false, "ignore unmarshaling errors")
)

func main() {
	var (
		excludeSourceIdentifiersFlags    xflag.Array
		openAccessSourceIdentifiersFlags xflag.Array
	)
	flag.Var(&excludeSourceIdentifiersFlags, "xsid",
		"exclude a given SID from checks, x.oa will always be false (repeatable)")
	flag.Var(&openAccessSourceIdentifiersFlags, "oasid",
		"always set x.oa true for a given sid (repeatable)")

	flag.Parse()

	if *showVersion {
		fmt.Println(span.AppVersion)
		os.Exit(0)
	}

	cfg := oafilter.Config{
		KbartFile:        *kbartFile,
		FreeContentFile:  *freeContentFile,
		Verbose:          *verbose,
		BatchSize:        *batchsize,
		BatchMemoryLimit: *batchMemoryLimit,
		BestEffort:       *bestEffort,
		ExcludeSids:      excludeSourceIdentifiersFlags,
		OpenAccessSids:   openAccessSourceIdentifiersFlags,
	}
	if err := oafilter.Run(cfg, os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
