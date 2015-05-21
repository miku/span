// Add holdings information to intermediate schema records.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miku/span/holdings"
)

// // Attacher can make a decision based on the record an an ISIL,
// // whether this ISIL should be attached or not.
// type Attacher interface {
// 	ShouldAttach(finc.IntermediateSchema, string) (bool, error)
// }

// // HoldingsAttacher make a decision based on Holdings information.
// type HoldingsAttacher struct {
// 	Table map[string]Licenses
// }

// func NewHoldingsAttacher() HoldingsAttacher {
// 	return HoldingsAttacher{Table: make(map[string]Licenses)}
// }

// func (ha HoldingsAttacher) ShouldAttach(is finc.IntermediaSchema, isil string) (bool, error) {
// 	date, err := is.Date()
// 	if err != nil {

// 	}
// 	holdings.CombineDatum(fmt.Sdate.Year, is.)
// }

func main() {
	skip := flag.Bool("skip", false, "skip errorneous entries")

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("input file required")
	}

	filename := flag.Arg(0)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	lmap, errs := holdings.ParseHoldings(file)
	if len(errs) > 0 && !*skip {
		for _, e := range errs {
			log.Println(e)
		}
		log.Fatal("errors during processing")
	}

	b, err := json.Marshal(lmap)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
