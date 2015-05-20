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

// ISILAttacher maps an ISIL to a number of attachers.
// If any of these attachers return true, the ISIL should be attached.
type ISILAttacher map[string][]Attacher

// Attach will populate the institutions fields accordingly.
func (a ISILAttacher) Attach(is finc.IntermediateSchema) []string {
	isils := container.NewStringSet()
	for isil, attachers := range a {
		for _, attacher := range attachers {
			if attacher.ShouldAttach(is) {
				isils.Add(isil)
			}
		}
	}
	return isils.Values()
}

type Attacher interface {
	ShouldAttach(is finc.IntermediateSchema) bool
}

type AttachByHolding struct {
	Table Licenses
}

func NewAttachByHolding(io.Reader) AttachByHolding {
	return AttachByHolding{Table: make(Licenses)}
}

// AttachByHolding compares the (year, volume, issue) of the
// record with license information, including possible moving walls.
func (a AttachByHolding) ShouldAttach(is finc.IntermediateSchema) bool {
	date, _ := is.Date()
	signature := holdings.CombineDatum(fmt.Sprintf(date.Year(), is.Volume, is.Issue, ""))
	now := time.Now()
	for _, issn := range append(is.ISSN, is.EISSN...) {
		licenses, ok := a.Table[issn]
		if !ok {
			continue
		}
		for _, l := range licenses {
			if l.Covers(signature) {
				if now.After(l.Boundary()) {
					return true
				}
			}
		}
	}
	return false
}

// AttachByList will include records, whose ISSN is contained in a given set.
type AttachByList struct {
	Set container.StringSet
}

func NewAttachByList(r io.Reader) AttachByList {
	br := bufio.NewReader(r)
	attacher := AttachByList{Set: container.NewStringSet()}
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		attacher.Add(strings.TrimSpace(line))
	}
	return attacher
}

func (a AttachByList) ShouldAttach(is finc.IntermediateSchema) bool {
	for _, issn := range ss.ISSN {
		if a.Set.Contains(issn) {
			return true
		}
	}
	return false
}

type AttachAll struct{}

func (a AttachAll) ShouldAttach(is finc.IntermediateSchema) bool {
	return true
}

type AttachNone struct{}

func (a AttachNone) ShouldAttach(is finc.IntermediateSchema) bool {
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
