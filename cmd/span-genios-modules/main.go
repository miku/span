// span-genios-modules tries to generate a fresh map from database name to
// modules (like https://git.io/v2ECx).
//
// $ cat filelist | span-genios-modules
//
// Previous list (Aug 22, 2017) contained 659 keys.
//
// $ curl -sL https://git.io/v94aR | jq '.|keys|length'
// 659
//
package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	humanize "github.com/dustin/go-humanize"
	"github.com/miku/parallel"
	"github.com/miku/xmlstream"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html/charset"
)

// Document was generated 2019-12-10 17:00:05 by tir on hayiti.
type Document struct {
	XMLName xml.Name `xml:"Document"`
	ID      string   `xml:"ID,attr"`
	IDNAME  string   `xml:"IDNAME,attr"`
	DB      string   `xml:"DB,attr"`
	Modules struct {
		Text   string   `xml:",chardata"`
		Module []string `xml:"Module"` // journal_wiwi
	} `xml:"Modules"`
}

type Map struct {
	sync.Mutex
	s map[string][]string
}

func stringSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func (m *Map) Add(key, value string) {
	m.Lock()
	defer m.Unlock()
	vs, ok := m.s[key]
	switch {
	case ok:
		if !stringSliceContains(vs, value) {
			vs = append(vs, value)
			m.s[key] = vs
		}
	default:
		m.s[key] = []string{value}
	}
}

func (m *Map) Size() int {
	return len(m.s)
}

func (m *Map) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.s)
}

func main() {

	dbmap := &Map{s: make(map[string][]string)}

	p := parallel.NewProcessor(os.Stdin, ioutil.Discard, func(b []byte) ([]byte, error) {
		filename := string(bytes.TrimSpace(b))
		if !strings.HasSuffix(filename, "zip") {
			return nil, fmt.Errorf("we need zip files for now: %s", filename)
		}

		fi, err := os.Stat(filename)
		if err != nil {
			return nil, err
		}

		r, err := zip.OpenReader(filename)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		for _, f := range r.File {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			scanner := xmlstream.NewScanner(rc, new(Document))
			scanner.Decoder.CharsetReader = charset.NewReaderLabel // https://stackoverflow.com/a/32224438
			for scanner.Scan() {
				tag := scanner.Element()
				switch v := tag.(type) {
				case *Document:
					for _, moduleName := range v.Modules.Module {
						dbmap.Add(v.DB, moduleName)
					}
				}
			}
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			rc.Close()
		}
		log.Printf("%v entries in dbmap [%s %s]",
			dbmap.Size(),
			humanize.Bytes(uint64(fi.Size())),
			filepath.Base(filename),
		)

		return nil, nil
	})

	p.BatchSize = 1
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}

	b, err := json.Marshal(dbmap)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
