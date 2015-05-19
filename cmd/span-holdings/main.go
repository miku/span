package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/miku/span/holdings"
)

const LOW = "000000000000000000000000000000"
const HIGH = "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"

func Singular(year, volume, issue string, empty string) string {
	if year == "" && volume == "" && issue == "" {
		return empty
	}
	return fmt.Sprintf("%010s%010s%010s", year, volume, issue)
}

type ISSN string

type ISSNLicense map[ISSN]Licenses
type Licenses []License
type License struct {
	From   string `json:"from"`
	To     string `json:"to"`
	URL    string `json:"url"`
	Status string `json:"status"`
}

func main() {

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("input file required")
	}
	filename := flag.Arg(0)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	decoder := xml.NewDecoder(bufio.NewReader(file))
	var tag string

	licenceMap := make(ISSNLicense)

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			tag = se.Name.Local
			if tag == "holding" {
				var item holdings.Holding
				decoder.DecodeElement(&item, &se)
				var lics Licenses
				for _, e := range item.Entitlements {
					from := Singular(e.FromYear, e.FromVolume, e.FromIssue, LOW)
					to := Singular(e.ToYear, e.ToVolume, e.ToIssue, HIGH)
					x, _ := url.QueryUnescape(e.URL)
					lics = append(lics, License{From: from, To: to, URL: x, Status: e.Status})
				}
				for _, issn := range append(item.EISSN, item.PISSN...) {
					for _, l := range lics {
						licenceMap[ISSN(issn)] = append(licenceMap[ISSN(issn)], l)
					}
				}
			}
		}
	}
	b, err := json.Marshal(licenceMap)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
