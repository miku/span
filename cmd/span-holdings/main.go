// Add holdings information to intermediate schema records.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/miku/span/container"
	"github.com/miku/span/finc"
	"github.com/miku/span/holdings"
)

// Filter wraps the decision, whether a given record should be attached or not.
type Filter interface {
	Apply(is finc.IntermediateSchema) bool
}

// Any attaches all records.
type Any struct{}

func (f Any) Apply(is finc.IntermediateSchema) bool { return true }

// None declines any record.
type None struct{}

func (f None) Apply(is finc.IntermediateSchema) bool { return false }

// ISILAttacher maps an ISIL to a number of Filters.
// If any of these filters return true, the ISIL should be attached.
type ISILTagger map[string][]Filter

// Tags will return all ISILs that can be attached to this record.
func (t ISILTagger) Tags(is finc.IntermediateSchema) []string {
	isils := container.NewStringSet()
	for isil, filters := range t {
		for _, f := range filters {
			if f.Apply(is) {
				isils.Add(isil)
			}
		}
	}
	return isils.Values()
}

// HoldingFilter decides ISIL-attachment by looking at licensing information from OVID files.
type HoldingFilter struct{ Table holdings.Licenses }

// NewHoldingFilter loads the holdings information for a single institution.
func NewHoldingFilter(r io.Reader) HoldingFilter {
	licenses, errors := holdings.ParseHoldings(r)
	if len(errors) > 0 {
		log.Fatal(errors)
	}
	return HoldingFilter{Table: licenses}
}

// HoldingFilter compares the (year, volume, issue) of the
// record with license information, including possible moving walls.
func (f HoldingFilter) Apply(is finc.IntermediateSchema) bool {
	// TODO(miku): make is.Date() fail earlier.
	date, _ := is.Date()
	signature := holdings.CombineDatum(fmt.Sprintf("%d", date.Year()), is.Volume, is.Issue, "")
	now := time.Now()
	for _, issn := range append(is.ISSN, is.EISSN...) {
		licenses, ok := f.Table[issn]
		if !ok {
			continue
		}
		for _, l := range licenses {
			if !l.Covers(signature) {
				continue
			}
			if now.After(l.Boundary()) {
				return true
			}
		}
	}
	return false
}

// ListFilter will include records, whose ISSN is contained in a given set.
type ListFilter struct {
	Set *container.StringSet
}

// NewAttachByList reads one record per line from reader.
func NewListFilter(r io.Reader) ListFilter {
	br := bufio.NewReader(r)
	f := ListFilter{Set: container.NewStringSet()}
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		f.Set.Add(strings.TrimSpace(line))
	}
	return f
}

func (f ListFilter) Apply(is finc.IntermediateSchema) bool {
	for _, issn := range append(is.ISSN, is.EISSN...) {
		if f.Set.Contains(issn) {
			return true
		}
	}
	return false
}

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
