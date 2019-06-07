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

// SourceNames generated from wiki via https://git.io/f4Qyi.
// curl -v "https://projekte.ub.uni-leipzig.de/projects/metadaten-quellen/wiki/SIDs.xml?key=s0mek3y" | xmlcutty -path /wiki_page/text | cut -f2,4 -d '|' | awk -F'|' '{print "\"" $1 "\": \"" $2 "\","}'
var SourceNames = map[string]string{
	"0":   "BSZ (SWB)",
	"1":   "Project Gutenberg",
	"2":   "BSZ Löschdaten",
	"3":   "Nielsen Book Data (UBL Profil)",
	"4":   "Schweitzer EBL (UBL Profil)",
	"5":   "Naxos Music Library",
	"6":   "Music Online Reference, MOR (Nationallizenz)",
	"7":   "Periodicals Archive Online, PAO (Nationallizenz) [obsolet!]",
	"8":   "Lizenzfreie elektronische Ressourcen, LFER (BSZ/SWB)",
	"9":   "Early Music Online",
	"10":  "Music Treasures Consortium",
	"11":  "Bibliographie des Musikschrifttums online (BMS online)",
	"12":  "B3Kat",
	"13":  "Diss Online",
	"14":  "Répertoire International des Sources Musicale (RISM)",
	"15":  "imslp.org (Petrucci)",
	"16":  "Elsevier E-Book Metadaten (UBL Profil)",
	"17":  "Nationallizenzen (eBooks / Monographien)",
	"18":  "Oxford Scholarship Online",
	"19":  "Kubon &amp; Sagner Digital",
	"20":  "Bibliothèque nationale de France (BnF): Gallica (Musik)",
	"21":  "GBVcentral Dumps Musik",
	"22":  "Qucosa",
	"23":  "Hochschulschriften HSZIGR Hochschule Zittau/Görlitz",
	"24":  "Ebrary (TUF Profil)",
	"25":  "SWB-Verbunddatenbank als Linked Open Data",
	"26":  "DOAB Directory of Open Access Books",
	"27":  "Munzinger",
	"28":  "DOAJ Directory of Open Access Journals",
	"29":  "Handwörterbuch der musikalischen Terminologie (vifa)",
	"30":  "SSOAR Social Science Open Access Repository",
	"31":  "Opera in Video (OPIV)",
	"32":  "Test Pilsen",
	"33":  "Test Liberec",
	"34":  "PQDT Open",
	"35":  "Hathi Trust",
	"36":  "HistBest System am URZ Leipzig",
	"37":  "Solr-Fusion statisches Mapping source_id für GBV",
	"38":  "Solr-Fusion statisches Mapping source_id für DBOD",
	"39":  "Persee.fr",
	"40":  "Dance in Video (DAIV)",
	"41":  "Classical Music in Video (CMIV)",
	"42":  "Classical Scores Library II (CSL2)",
	"43":  "TUF Digitale Bibliothek (Anreicherung SWB-Daten)",
	"44":  "Deutsches Textarchiv (DTA)",
	"45":  "Solr-Fusion statisches Mapping DBOD No. 2",
	"46":  "De Gruyter PDA eBooks (HTW Dresden Profil)",
	"47":  "Vahlen 2014",
	"48":  "GBI Genios Wiso",
	"49":  "Crossref",
	"50":  "DeGruyter SSH",
	"51":  "VUB PDA print dt (UBL Profil)",
	"52":  "OECD iLibrary",
	"53":  "CEEOL Central and Eastern European Online Library",
	"54":  "World Bank eLibrary",
	"55":  "JSTOR",
	"56":  "Folkwang Universität der Künste Essen, Lokalsystem",
	"57":  "Robert Schumann Hochschule RSH Düsseldorf, Lokalsystem",
	"58":  "Hochschule für Musik und Tanz HfMT Köln, Lokalsystem",
	"59":  "MyiLibrary PDA (UBL Profil)",
	"60":  "Thieme",
	"61":  "IMF International Monetary Fund",
	"62":  "Juris",
	"63":  "De Gruyter eBooks EBS (Sachsen Profil)",
	"64":  "Databases on Demand (DBoD)",
	"65":  "GBV Central (sonstige Inhalte)",
	"66":  "HeidICON",
	"67":  "SLUB/Deutsche Fotothek",
	"68":  "OLC Online Contents (aus GBV Central)",
	"69":  "Wiley ebooks (SLUB Profil)",
	"70":  "Institut Ägyptologie (UBL)",
	"71":  "OstDok (Virtuelle Fachbibliothek Osteuropa)",
	"72":  "Morgan &amp; Claypool eBooks",
	"73":  "MedienwRezensionen",
	"74":  "Taylor &amp; Francis eBooks EBA (UBL Profil)",
	"75":  "ProQuest Film-Datenbanken",
	"76":  "PDA eBooks EBL Schweitzer (FID Profil)",
	"77":  "Portraitstichsammlung UB Leipzig",
	"78":  "IZI Datenbank",
	"79":  "EBL PDA (Berufsakademien Sachsen Profil)",
	"80":  "Datenbank Internetquellen",
	"81":  "Digitale Sammlungen der ULB Düsseldorf",
	"82":  "Jahresbibliografie Massenkommunikation",
	"83":  "SLUB/Mediathek",
	"84":  "OCLC WorldShare Collection Manager",
	"85":  "Elsevier Journals (Sachsen Profil)",
	"86":  "Cambridge University Press eBooks (UBL Profil)",
	"87":  "International Journal of Communication",
	"88":  "Zeitschrift \"Rundfunk und Geschichte\"",
	"89":  "IEEE Xplore Library",
	"92":  "DBIS",
	"93":  "ZVDD",
	"94":  "Nationallizenzen Zeitschriften (aus GBV Central)",
	"95":  "Alexander Street Academic Video Online",
	"96":  "Thieme eBooks Tiermedizin VetCenter",
	"97":  "Schweitzer EBL (SLUB Profil)",
	"98":  "GNM - Germanisches Nationalmuseum Nürnberg",
	"99":  "Media Perspektiven",
	"100": "Medienwissenschaft \"Berichte und Papiere\"",
	"101": "Kieler Beiträge zur Filmmusikforschung",
	"102": "Daphne Objektdatenbank SKD",
	"103": "Margaret Herrick Library",
	"104": "WTI",
	"105": "Springer Journals",
	"106": "Primary Sources for Slavic Studies",
	"107": "Heidelberger historische Bestände digital",
	"108": "De Gruyter eBooks Open Access",
	"109": "Kunsthochschule für Medien Köln (VK Film)",
	"110": "Sowiport (GESIS)",
	"111": "Volltexte \"Illustrierte Presse\"",
	"112": "Deutsche Nationalbibliothek Leipzig",
	"113": "Loeb Classical Library",
	"114": "Ebook Central - SUB Collection \"The Arts\"",
	"115": "Heidelberger Bibliographien",
	"116": "Deutsches Filminstitut Frankfurt (VK Film)",
	"117": "Universität der Künste Berlin (VK Film)",
	"118": "Zentrum für Kunst und Medien Karlsruhe (VK Film)",
	"119": "Universitätsbibliothek Frankfurt am Main (VK Film)",
	"120": "Cambridge University Press eBooks (Sachsen)",
	"121": "Arxiv",
	"122": "Springer eBooks",
	"123": "OpenAIRE",
	"124": "DawsonEra PDA (Profil RWTH Aachen)",
	"125": "Universitätsbibliothek Siegen (VK Film)",
	"126": "BASE Bielefeld Academic Search Engine",
	"127": "Filmuniversität Konrad Wolf Potsdam (VK Film)",
	"128": "McGraw-Hill Access Engineering (RWTH Aachen Profil)",
	"129": "GEOSCAN Database",
	"130": "Stahlinstitut VDEh - Fachbibliothek",
	"131": "GDMB",
	"132": "Römpp (Thieme)",
	"133": "Cambridge University Press Journals",
	"134": "Feministische Bibliothek MONAliesA",
	"135": "Elsevier EBS (Profil RWTH Aachen)",
	"136": "Leipziger Städtische Bibliotheken",
	"137": "Bundeskunsthalle Bonn",
	"138": "Europeana",
	"139": "Deutsche Digitale Bibliothek (DDB)",
	"140": "Kalliope-Verbund",
	"141": "Lynda.com (video2brain)",
	"142": "Gesamtkatalog der Düsseldorfer Kulturinstitute (VK Film)",
	"143": "JOVE Journal of Visualized Experiments",
	"144": "Taylor and Francis EBS (UBC, HSZG)",
	"145": "TIB AV-Portal",
	"146": "Produktionsarchivdaten Fernsehen von ARD und ZDF",
	"147": "Nationallizenz Palgrave Economics &amp; Finance 2000-2013",
	"148": "Bundesarchiv (Filmarchiv)",
	"149": "Wiley Journals (Sachsen)",
	"150": "MOnAMi Hochschulschriftenserver Mittweida",
	"151": "Filmakademie Baden-Württemberg",
	"152": "United Nations iLibrary",
	"153": "Internet Archive (archive.org)",
	"154": "CAMECO Catholic Media Council",
	"155": "Deutsche Kinemathek (VK Film)",
	"156": "Umweltbibliothek Leipzig",
	"157": "Hanser EBS eBooks (Berufsakademien Sachsen Profil)",
	"158": "Olms Online",
	"159": "Buchhandschriften der UB Leipzig",
	"160": "Diplomarbeiten der Sportwissenschaftlichen Fakultät der Universität Leipzig",
	"161": "Cambridge eBooks Open Access",
	"162": "Gender Open Repositorium",
	"163": "Nachlässe der UB Leipzig",
	"164": "Japanische Videospielsammlung (UBL)",
	"165": "Tectum eBooks (HfMT Köln)",
	"166": "Kunsthistorisches Institut Florenz",
	"167": "Bildarchiv Foto Marburg",
	"168": "Bibliotheca Hertziana – Max–Planck–Institut für Kunstgeschichte",
	"169": "MediathekView",
	"170": "media/rep/ - Repositorium für die Medienwissenschaft",
	"171": "PressReader (via Digento)",
	"172": "OER aus OPAL (UB Freiberg)",
	"173": "Wolfenbütteler Bibliographie zur Geschichte des Buchwesens",
	"174": "Zenodo",
	"175": "Libris-Katalog der schwedischen Nationalbibliothek",
	"176": "Fennica-Katalog der finnischen Nationalbibliothek",
	"200": "finc TEST",
	"201": "Perinorm",
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
