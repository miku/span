// span-genios-modules tries to generate a fresh map from database name to
// modules (like https://git.io/v2ECx).
//
// $ cat filelist | span-genios-modules
//
// Where filelist contains a list (one per line) WISO supplied RELOAD zip file,
// e.g. xyz_WWON_reload_201911.zip.
//
// Previous list (Aug 22, 2017) contained 659 keys. Current list (Dec 11, 2019) 743.
//
// $ curl -sL https://git.io/v94aR | jq '.|keys|length'
// 659
//
// Compare old to new list.
//
// $ comm <(curl -sL https://git.io/v94aR | jq -r '.|keys[]') <(curl -sL https://git.io/Je522 | jq -r '.|keys[]')
//
// Discontinued
// ABAL ABKL ABMS ABSC AE ANP AUTO AUW AVFB BB CBS CI DZI EMAR FERT FLUI FOGR
// HMDS HOLZ INST KE KONT LEDI LEMO MF MUM PMGC PMGI PROD SBIL STB STIF WEBR
// WEFO WRP WUV.
//
// New
// ADWM BAVE BDI BIKO BIOS BIZT BUAE BUDR BUVI CBNB CFI CHET CICE CZWD CZWE
// DATV DBBL DEH DFVE DIPL DNV DU DUHU DVZO EAWW EBOK EGGI EIND EJ EMA EMT ERIS
// ESP EWIS FAZB FBPG FBV FEDS FEMI FFR FG FOVO FWP GABA GEND GPGD GRIA GRIN
// HAUF HAUT HBDM HBIF HUSS IEE IJRE IKPR IKZF IKZP IMKP IMKR IMKS IMKW INFE
// INW IRJ JAPS JCE KAS KI KMUV KUEP LF LMW LUCE LZDI MACH MBFR MIWI MTK MU MUL
// NOTA NV PBST PCS PHF PIR PLS PLV POTI PRTR RBIH REDL RUP SOCI SOLA SOZM SSOA
// STUB TOPA TRT TUE VEJA VER VHAU VKU WLBU WPUS WSIP WSIS WSIW WWBW WZM ZAP
// ZERB ZFIP ZFMF ZISU ZPTH ZQF.
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

func stringSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// Map maps a string key to a set of strings, atomically.
type Map struct {
	sync.Mutex
	s map[string][]string
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
