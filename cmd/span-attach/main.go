// span-attach -c config.tsv < in > out
//
// Does what span-tag did, but try to boil it down to a two step process: find
// relevant lines in a tabular file for attachment (generated from AMSL or
// FOLIO MOD-FINC-CONFIG), then - if required - use relevant entry from holding
// file in a second step and attach ISIL to input file.
//
// Also, try to get rid of span-freeze by using some locally cached files (with
// expiry date).
package main

import (
	"flag"
	"log"
)

var (
	force   = flag.Bool("f", false, "force all external referenced links to be downloaded")
	tabConf = flag.String("c", "", "path to tabular config file (format, see: TBA)")
)

func main() {
	flag.Parse()
	log.Println("attach is the new tag")
}
