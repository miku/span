// span-compare renders a table with ISIL/SID counts of two indices side by
// side. It can parse the hidden whatislive endpoint to find live, non-live
// pairs.
//
//   $ span-compare -a 10.1.1.7:8085/solr/biblio -b 10.1.1.15:8085/solr/biblio
//   DE-1156    13   Diss Online                         0         219277    219277
//   DE-D117    13   Diss Online                         0         219277    219277
//   DE-L152    13   Diss Online                         0         219277    219277
//   DE-14      49   Crossref                            35417291  36671073  1253782
//   DE-14      48   GBI Genios Wiso                     12712115  174470    -12537645
//   DE-14      85   Elsevier Journals (Sachsen Profil)  1359818   1397962   38144
//   DE-14      13   Diss Online                         0         219277    219277
//   DE-D275    49   Crossref                            0         32177218  32177218
//   DE-D275    48   GBI Genios Wiso                     10501192  2050930   -8450262
//   DE-Gla1    49   Crossref                            31640516  31467567  -172949
//   ...
//
// Use hidden whatislive endpoint and render textile (for redmine):
//
//   $ span-compare -e -t
//   ...
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/miku/span"
	"github.com/miku/span/solrutil"

	log "github.com/sirupsen/logrus"
)

// SourceNames generated from AMSL API.
// $ taskcat AMSLService | jq -rc '.[]| [.sourceID, .megaCollection] | @tsv' |
// sort -un | awk -F '      ' '{printf "\"%s\":\"%s\",\n", $1, $2 }' # CTRL-TAB
var SourceNames = map[string]string{
	"0":   "Digital Concert Hall",
	"3":   "PDA Print Nielsen",
	"4":   "PDA E-Book Central",
	"5":   "Naxos Music Library",
	"9":   "Early Music Online",
	"10":  "Music Treasures Consortium",
	"12":  "B3Kat BBI",
	"13":  "DNB / Diss Online",
	"14":  "RISM Répertoire International des Sources Musicale",
	"15":  "IMSLP (Petrucci Library)",
	"16":  "Elsevier E-Books",
	"17":  "17th and 18th Century Burney Collection Newspapers",
	"18":  "Oxford E-Books / Economics and Finance",
	"20":  "BNF / Gallica / Buch und Bibliothek",
	"21":  "GBV Musikdigitalisate",
	"22":  "Qucosa",
	"24":  "E-Book Central",
	"26":  "DOAB Directory of Open Access Books",
	"27":  "Munzinger / Chronik",
	"28":  "DOAJ Directory of Open Access Journals",
	"29":  "Handwörterbuch der musikalischen Terminologie",
	"30":  "SSOAR Social Science Open Access Repository",
	"31":  "Opera in Video",
	"34":  "ProQuest Open Access Dissertations and Theses (PQDT Open)",
	"35":  "Hathi Trust",
	"39":  "Persée",
	"40":  "Dance in Video",
	"41":  "Classical Music in Video",
	"42":  "Classical Scores Library II",
	"44":  "Deutsches Textarchiv",
	"48":  "Wiso Journals",
	"49":  "CrossRef",
	"50":  "De Gruyter Journals / Social Sciences and Humanities",
	"51":  "PDA Print VUB",
	"52":  "OECD iLibrary",
	"53":  "CEEOL Central and Eastern European Online Library",
	"55":  "JSTOR Arts & Sciences I Archive",
	"56":  "Folkwang Universität der Künste Essen, Lokalsystem",
	"57":  "Robert Schumann Hochschule RSH Düsseldorf",
	"58":  "Hochschule für Musik und Tanz HfMT Köln",
	"60":  "Thieme Journals",
	"61":  "IMF International Monetary Fund",
	"63":  "De Gruyter E-Books",
	"65":  "GVK Verbundkatalog / Nordeuropa",
	"66":  "Heidelberger Bilddatenbank",
	"67":  "Künstler/Conart",
	"68":  "OLC SSG Anglistik",
	"70":  "UBL / Institut Ägyptologie",
	"71":  "OstDok",
	"72":  "Morgan & Claypool E-Books",
	"73":  "MedienwRezensionen",
	"76":  "E-Books adlr",
	"77":  "UBL / Portraitstichsammlung",
	"78":  "IZI-Datenbank",
	"79":  "E-Book Central",
	"80":  "Datenbank Internetquellen",
	"84":  "JSTOR E-Books Open Access",
	"85":  "Elsevier Journals",
	"86":  "Cambridge E-Books",
	"87":  "International Journal of Communication",
	"88":  "Zeitschrift \"Rundfunk und Geschichte\"",
	"89":  "IEEE Xplore Library",
	"94":  "17th and 18th Century Burney Collection Newspapers",
	"96":  "Thieme E-Books / Tiermedizin (VetCenter)",
	"99":  "Media Perspektiven",
	"101": "Kieler Beiträge zur Filmmusikforschung",
	"103": "Margaret Herrick Library",
	"105": "Springer Journals",
	"106": "Primary Sources for Slavic Studies",
	"107": "Heidelberger historische Bestände",
	"109": "Kunsthochschule für Medien Köln (VK Film)",
	"111": "Volltexte aus Illustrierten Magazinen",
	"113": "Loeb Classical Library",
	"114": "Academic E-Books 'The Arts'",
	"115": "Bibliographie Caricature & Comic",
	"117": "Universität der Künste Berlin (VK Film)",
	"119": "Universitätsbibliothek Frankfurt am Main (VK Film)",
	"126": "BASE_meta_collection",
	"127": "Filmuniversität Konrad Wolf Potsdam (VK Film)",
	"129": "GEOSCAN",
	"130": "VDEH",
	"131": "GDMB",
	"133": "Cambridge Journals",
	"134": "MONAliesA",
	"136": "Leipziger Städtische Bibliotheken",
	"137": "Bundeskunsthalle Bonn",
	"140": "Nachlässe SLUB Dresden",
	"142": "Gesamtkatalog der Düsseldorfer Kulturinstitute (VK Film)",
	"143": "JOVE Journal of Visualized Experiments (Biology)",
	"145": "TIB AV-Portal",
	"147": "Palgrave E-Books / Economics & Finance",
	"148": "Bundesarchiv (Filmarchiv)",
	"150": "MOnAMi Hochschulschriftenserver Mittweida",
	"151": "Filmakademie Baden-Württemberg (VK Film)",
	"153": "Internet Archive / Animation Shorts",
	"156": "Umweltbibliothek Leipzig",
	"158": "Olms / Altertumswissenschaft",
	"159": "Digitale Sammlungen UBL / Abendländische mittelalterliche Handschriften",
	"160": "UBL / Diplomarbeiten Sportwissenschaft",
	"162": "Gender Open",
	"163": "Digitale Sammlungen UBL / Nachlass Karl Bücher",
	"165": "Tectum E-Books",
	"166": "Kunsthistorisches Institut Florenz",
	"167": "Bildarchiv Foto Marburg",
	"168": "Bibliotheca Hertziana - Max-Planck-Institut für Kunstgeschichte",
	"169": "MediathekView",
	"170": "media/rep/",
	"171": "PressReader",
	"172": "OPAL",
	"173": "Wolfenbütteler Bibliographie zur Geschichte des Buchwesens",
	"175": "Libris-Katalog der schwedischen Nationalbibliothek",
	"176": "Fennica-Katalog der finnischen Nationalbibliothek",
	"178": "Nationalbibliographie Norwegen",
	"179": "LIS Scholarship Archive",
	"180": "British National Bibliography",
	"181": "British Library Catalogue",
	"182": "Bibliographie zum Archivwesen der Archivschule Marburg",
	"183": "K10plus Verbundkatalog",
	"184": "RILM Music Encyclopedias",
	"185": "DABI Datenbank Deutsches Bibliothekswesen",
	"186": "Buch und Papier",
	"187": "UBL / Digitalisierte Zeitschriften (UrMEL)",
	"188": "GoeScholar - Publikationenserver der Georg-August-Universität Göttingen",
	"190": "Early English Books Online EEBO",
	"191": "MediArxiv",
	"201": "Perinorm (DIN-Normen)",
	"202": "Zentralinstitut für Kunstgeschichte",
}

var defaultConfigPath = path.Join(span.UserHomeDir(), ".config/span/span.json")

var (
	liveServer       = flag.String("a", "http://localhost:8983/solr/biblio", "live server location")
	nonliveServer    = flag.String("b", "http://localhost:8983/solr/biblio", "non-live server location")
	whatIsLive       = flag.Bool("e", false, "use whatislive.url to determine live and non live servers")
	liveLinkTemplate = flag.String("tl", "https://katalog.ub.uni-leipzig.de/Search/Results?lookfor=source_id:{{ .SourceID }}",
		"live link template for source (for focus institution)")
	spanConfigFile   = flag.String("span-config", defaultConfigPath, "for whatislive.url")
	textile          = flag.Bool("t", false, "emit textile")
	focusInstitution = flag.String("emph", "DE-15", "emphasize institution in textile output")
)

// ResultWriter for report generator.
type ResultWriter interface {
	WriteHeader(...string)
	WriteFields(...interface{})
	Err() error
}

// TabWriter is the simplest writer.
type TabWriter struct {
	w   io.Writer
	err error
}

func (w *TabWriter) WriteHeader(header ...string) {}
func (w *TabWriter) WriteFields(fields ...interface{}) {
	var s []string
	for _, f := range fields {
		s = append(s, fmt.Sprintf("%v", f))
	}
	_, w.err = fmt.Fprintf(w.w, "%s\n", strings.Join(s, "\t"))
}
func (w *TabWriter) Err() error {
	return w.err
}

// TextileWriter allows to writer textile tables.
type TextileWriter struct {
	w       io.Writer
	columns int
	err     error
}

// Err returns any error that happened.
func (w *TextileWriter) Err() error {
	return w.err
}

// WriteHeader write a header and fixes the number of columns.
func (w *TextileWriter) WriteHeader(header ...string) {
	if w.columns > 0 || w.err != nil {
		return
	}
	w.columns = len(header)
	var decorated []string
	for _, h := range header {
		decorated = append(decorated, fmt.Sprintf("*%s*", h))
	}
	_, w.err = fmt.Fprintf(w.w, "| %s |\n", strings.Join(decorated, " | "))
}

// WriteFields writes fields.
func (w *TextileWriter) WriteFields(fields ...interface{}) {
	if w.err != nil {
		return
	}
	if len(fields) != w.columns {
		w.err = fmt.Errorf("got %d fields, want %d", len(fields), w.columns)
		return
	}
	var s []string
	for _, f := range fields {
		var v string
		switch t := f.(type) {
		case string:
			v = t
		case fmt.Stringer:
			v = t.String()
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			v = fmt.Sprintf("%d", t)
		case float32, float64:
			v = fmt.Sprintf("%0.4f", t)
		default:
			v = fmt.Sprintf("%s", t)
		}
		s = append(s, v)
	}
	_, w.err = fmt.Fprintf(w.w, fmt.Sprintf("| %s |\n", strings.Join(s, " | ")))
}

// prependHTTP prepends http, if necessary.
func prependHTTP(s string) string {
	if !strings.HasPrefix(s, "http") {
		return fmt.Sprintf("http://%s", s)
	}
	return s
}

// renderSourceLink renders a textile link to a source, with the given link text.
func renderSourceLink(tmpl string, data interface{}, text string) (string, error) {
	t, err := template.New("t").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return fmt.Sprintf(`"%s":%s`, text, buf.String()), nil
}

func main() {
	flag.Parse()

	if *whatIsLive {
		// Fallback configuration.
		if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
			*spanConfigFile = "/etc/span/span.json"
		}
		if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
			log.Fatal(err)
		}

		var err error

		*liveServer, err = solrutil.FindLiveSolrServer(*spanConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		*nonliveServer, err = solrutil.FindNonliveSolrServer(*spanConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("live=%s, nonlive=%s", *liveServer, *nonliveServer)
	}

	live := solrutil.Index{Server: prependHTTP(*liveServer)}
	nonlive := solrutil.Index{Server: prependHTTP(*nonliveServer)}

	sids, err := nonlive.SourceIdentifiers()
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(sids)

	institutions, err := nonlive.Institutions()
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(institutions)

	var rw ResultWriter
	switch {
	case *textile:
		rw = &TextileWriter{w: os.Stdout}
	default:
		rw = &TabWriter{w: os.Stdout}
	}

	rw.WriteHeader("ISIL", "Source", "Name", "Live", "Nonlive", "Diff", "Pct", "Comment")

	// TODO(miku): Parallelize queries.
	for _, institution := range institutions {
		if strings.TrimSpace(institution) == "" {
			continue
		}
		// TODO(miku): This should not be in the index in the first place
		if strings.TrimSpace(institution) == `" "` {
			continue
		}
		for _, sid := range sids {
			query := fmt.Sprintf(`source_id:"%s" AND institution:"%s"`, sid, institution)
			numLive, err := live.NumFound(query)
			if err != nil {
				log.Fatal(err)
			}
			numNonlive, err := nonlive.NumFound(query)
			if err != nil {
				log.Fatal(err)
			}
			// TODO(miku): Might catch too much, e.g. DOAJ, refs #14417.
			if numLive == 0 && numNonlive == 0 {
				continue
			}
			name, ok := SourceNames[sid]
			if !ok {
				name = "XXX: missing source name"
			}

			// Emphasize focussed institution.
			var renderInstitution = institution
			if *textile && institution == *focusInstitution {
				renderInstitution = fmt.Sprintf("*%s*", institution)
			}

			// Fields with links.
			var liveField = fmt.Sprintf("%d", numLive)
			var nonliveField = fmt.Sprintf("%d", numNonlive)

			// Percentage change, refs #12756.
			var pctChange float64
			switch {
			case numLive == 0 && numNonlive > 0:
				pctChange = 100
			default:
				pctChange = (float64(numNonlive-numLive) / (float64(numLive))) * 100
			}

			// Remove -0.00 from rendering.
			if pctChange == 0 {
				pctChange = math.Copysign(pctChange, 1)
			}
			var pctChangeField string

			switch {
			case pctChange > 5.0 || pctChange < -5.0:
				pctChangeField = fmt.Sprintf("*%0.2f*", pctChange)
			default:
				pctChangeField = fmt.Sprintf("%0.2f", pctChange)
			}

			if *textile {
				data := struct {
					SourceID    string
					Institution string
				}{
					sid,
					institution,
				}
				// XXX: Put all live link templates into a configaration file (or scrape from wiki).
				if institution == *focusInstitution {
					liveField, err = renderSourceLink(*liveLinkTemplate, data, fmt.Sprintf("%d", numLive))
					if err != nil {
						log.Fatal(err)
					}
				} else {
					liveField, err = renderSourceLink(*liveLinkTemplate, data, fmt.Sprintf("%d", numLive))
					if err != nil {
						log.Fatal(err)
					}
				}
				nonliveField = fmt.Sprintf("%d", numNonlive)
			}

			rw.WriteFields(renderInstitution, sid, name, liveField, nonliveField, numNonlive-numLive, pctChangeField, "")
			if rw.Err() != nil {
				log.Fatal(rw.Err())
			}
		}
	}
}
