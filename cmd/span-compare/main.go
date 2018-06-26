// WIP: move siskin:bin/indexcompare into a tool, factor out solr stuff into
// solrutil.go.
//
// $ span-compare -a 10.1.1.7:8085/solr/biblio  -b 10.1.1.15:8085/solr/biblio
// DE-1156    13   Diss Online                         0         219277    219277
// DE-D117    13   Diss Online                         0         219277    219277
// DE-L152    13   Diss Online                         0         219277    219277
// DE-14      49   Crossref                            35417291  36671073  1253782
// DE-14      48   GBI Genios Wiso                     12712115  174470    -12537645
// DE-14      85   Elsevier Journals (Sachsen Profil)  1359818   1397962   38144
// DE-14      13   Diss Online                         0         219277    219277
// DE-D275    49   Crossref                            0         32177218  32177218
// DE-D275    48   GBI Genios Wiso                     10501192  2050930   -8450262
// DE-Gla1    49   Crossref                            31640516  31467567  -172949
// ...
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// SourceNames generated from wiki via https://git.io/f4Qyi.
var SourceNames = map[string]string{
	"0":   "BSZ (SWB)",
	"1":   "Project Gutenberg",
	"2":   "BSZ Löschdaten",
	"3":   "Nielsen Book Data (UBL Profil)",
	"4":   "Schweitzer EBL (UBL Profil)",
	"5":   "Naxos Music Library",
	"6":   "\" \"\"Music Online Reference\"",
	"7":   "\" \"\"Periodicals Archive Online\"",
	"8":   "\" \"\"Lizenzfreie elektronische Ressourcen\"",
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
	"19":  "Kubon & Sagner Digital",
	"20":  "Bibliothèque nationale de France (BnF) : Gallica Musik",
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
	"56":  "\" \"\"Folkwang Universität der Künste Essen\"",
	"57":  "\" \"\"Robert Schumann Hochschule RSH Düsseldorf\"",
	"58":  "\" \"\"Hochschule für Musik und Tanz HfMT Köln\"",
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
	"72":  "Morgan & Claypool eBooks",
	"73":  "MedienwRezensionen",
	"74":  "Taylor & Francis eBooks EBA (UBL Profil)",
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
	"88":  "\" \"\"Zeitschrift \\\"\"Rundfunk und Geschichte\\\"\"\"\"\"",
	"89":  "IEEE Xplore Library",
	"92":  "DBIS",
	"93":  "ZVDD",
	"94":  "Nationallizenzen Zeitschriften (aus GBV Central)",
	"95":  "Alexander Street Academic Video Online",
	"96":  "Thieme eBooks Tiermedizin VetCenter",
	"97":  "Schweitzer EBL (SLUB Profil)",
	"98":  "GNM - Germanisches Nationalmuseum Nürnberg",
	"99":  "Media Perspektiven",
	"100": "\" \"\"Medienwissenschaft \\\"\"Berichte und Papiere\\\"\"\"\"\"",
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
	"111": "\" \"\"Volltexte \\\"\"Illustrierte Presse\\\"\"\"\"\"",
	"112": "Deutsche Nationalbibliothek Leipzig",
	"113": "Loeb Classical Library",
	"114": "\" \"\"Ebook Central - SUB Collection \\\"\"The Arts\\\"\"\"\"\"",
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
	"144": "\" \"\"Taylor and Francis EBS (UBC\"",
	"145": "TIB AV-Portal",
	"146": "Produktionsarchivdaten Fernsehen von ARD und ZDF",
	"147": "Nationallizenz Palgrave Economics & Finance 2000-2013",
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
	"200": "finc TEST",
}

var (
	liveServer    = flag.String("a", "http://localhost:8983/solr/biblio", "live server location")
	nonliveServer = flag.String("b", "http://localhost:8983/solr/biblio", "non-live server location")
)

// prependHTTP prepends http, if necessary.
func prependHTTP(s string) string {
	if !strings.HasPrefix(s, "http") {
		return fmt.Sprintf("http://%s", s)
	}
	return s
}

// Index index type implementing various query types and operations.
type Index struct {
	Server     string
	FacetLimit int
}

// SelectLink constructs a link to a JSON response. XXX: Allow client to
// override options.
func (ix Index) SelectLink(query string) string {
	vals := url.Values{}
	if query == "" {
		query = "*:*"
	}
	vals.Add("q", query)
	vals.Add("wt", "json")

	return fmt.Sprintf("%s/select?%s", ix.Server, vals.Encode())
}

// FacetLink constructs a link to a JSON response.
func (ix Index) FacetLink(query, facetField string) string {
	vals := url.Values{}
	if query == "" {
		query = "*:*"
	}
	if ix.FacetLimit == 0 {
		ix.FacetLimit = 100000
	}
	vals.Add("q", query)
	vals.Add("facet", "true")
	vals.Add("facet.field", facetField)
	vals.Add("facet.limit", fmt.Sprintf("%d", ix.FacetLimit))
	vals.Add("rows", "0")
	vals.Add("wt", "json")

	return fmt.Sprintf("%s/select?%s", ix.Server, vals.Encode())
}

// Facet runs a facet query.
func (ix Index) Facet(query, facetField string, r *FacetResponse) error {
	return decodeLink(ix.FacetLink(query, facetField), r)
}

// Select runs a select query.
func (ix Index) Select(query string, r *SelectResponse) error {
	return decodeLink(ix.SelectLink(query), r)
}

// decodeLink fetches a link and unmarshals the JSON response into a given value.
func decodeLink(link string, val interface{}) error {
	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("select failed with HTTP %d at %s", resp.StatusCode, link)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(val)
}

// FacetKeysFunc returns all facet keys, that pass a filter, given as function
// of facet value and frequency.
func (ix Index) FacetKeysFunc(query, field string, f func(string, int) bool) (result []string, err error) {
	var r FacetResponse
	if err := ix.Facet(query, field, &r); err != nil {
		return result, err
	}
	fmap, err := r.Facets()
	if err != nil {
		return result, err
	}
	for k, v := range fmap {
		if f(k, v) {
			result = append(result, k)
		}
	}
	return result, nil
}

// FacetKeys returns the values of a facet as a string slice.
func (ix Index) FacetKeys(query, field string) (result []string, err error) {
	var r FacetResponse
	if err := ix.Facet(query, field, &r); err != nil {
		return result, err
	}
	fmap, err := r.Facets()
	if err != nil {
		return result, err
	}
	for k := range fmap {
		result = append(result, k)
	}
	return result, nil
}

// Institutions returns a list of International Standard Identifier for
// Libraries and Related Organisations (ISIL), ISO 15511 identifiers.
func (ix Index) Institutions() (result []string, err error) {
	return ix.FacetKeys("*:*", "institution")
}

// SourceIdentifiers returns a list of source identifiers.
func (ix Index) SourceIdentifiers() (result []string, err error) {
	return ix.FacetKeys("*:*", "source_id")
}

// SourceCollections returns the collections for a given source identifier.
func (ix Index) SourceCollections(sid string) (result []string, err error) {
	return ix.FacetKeysFunc("source_id:"+sid, "mega_collection",
		func(_ string, v int) bool { return v > 0 })
}

// NumFound returns the size of the result set for a query.
func (ix Index) NumFound(query string) (int64, error) {
	var r SelectResponse
	if err := ix.Select(query, &r); err != nil {
		return 0, err
	}
	return r.Response.NumFound, nil
}

// SelectResponse wraps a select response, adjusted from JSONGen.
type SelectResponse struct {
	Response struct {
		Docs     []interface{} `json:"docs"`
		NumFound int64         `json:"numFound"`
		Start    int64         `json:"start"`
	} `json:"response"`
	ResponseHeader struct {
		Params struct {
			Indent string `json:"indent"`
			Q      string `json:"q"`
			Wt     string `json:"wt"`
		} `json:"params"`
		QTime  int64
		Status int64 `json:"status"`
	} `json:"responseHeader"`
}

// FacetMap maps a facet value to frequency. Solr uses pairs put into a list,
// which is a bit awkward to work with.
type FacetMap map[string]int

// FacetResponse wraps a facet response, adjusted from JSONGen output.
type FacetResponse struct {
	FacetCounts struct {
		FacetDates struct {
		} `json:"facet_dates"`
		FacetFields   json.RawMessage `json:"facet_fields"`
		FacetHeatmaps struct {
		} `json:"facet_heatmaps"`
		FacetIntervals struct {
		} `json:"facet_intervals"`
		FacetQueries struct {
		} `json:"facet_queries"`
		FacetRanges struct {
		} `json:"facet_ranges"`
	} `json:"facet_counts"`
	Response struct {
		Docs     []interface{} `json:"docs"`
		NumFound int64         `json:"numFound"`
		Start    int64         `json:"start"`
	} `json:"response"`
	ResponseHeader struct {
		Params struct {
			Facet      string `json:"facet"`
			Facetfield string `json:"facet.field"`
			Indent     string `json:"indent"`
			Q          string `json:"q"`
			Rows       string `json:"rows"`
			Wt         string `json:"wt"`
		} `json:"params"`
		QTime  int64
		Status int64 `json:"status"`
	} `json:"responseHeader"`
}

// Facets unwraps the facet_fields list into a FacetMap. The facet_fields list
// contains an even number of elements, with items at even positions being the
// facet value, at odd positions the cardinality as int.
func (fr FacetResponse) Facets() (FacetMap, error) {
	unwrap := make(map[string]interface{})
	if err := json.Unmarshal(fr.FacetCounts.FacetFields, &unwrap); err != nil {
		return nil, err
	}
	if len(unwrap) == 0 {
		return nil, fmt.Errorf("invalid response")
	}
	if len(unwrap) > 1 {
		return nil, fmt.Errorf("multiple facet fields not implemented")
	}

	result := make(FacetMap)
	var name string
	var freq float64

	for _, v := range unwrap {
		flist, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("facet frequencies is not a list")
		}
		for i, item := range flist {
			if i%2 == 0 {
				if name, ok = item.(string); !ok {
					return nil, fmt.Errorf("expected string at even position, got %T", item)
				}
			} else {
				if freq, ok = item.(float64); !ok {
					return nil, fmt.Errorf("expected int at odd position, got %T", item)
				}
				result[name] = int(freq)
			}
		}
	}
	return result, nil
}

func main() {
	flag.Parse()

	live := Index{Server: prependHTTP(*liveServer)}
	nonlive := Index{Server: prependHTTP(*nonliveServer)}

	sids, err := nonlive.SourceIdentifiers()
	if err != nil {
		log.Fatal(err)
	}
	institutions, err := nonlive.Institutions()
	if err != nil {
		log.Fatal(err)
	}
	for _, institution := range institutions {
		for _, sid := range sids {
			query := fmt.Sprintf(`source_id:"%s" AND institution:"%s"`, sid, institution)
			liveCount, err := live.NumFound(query)
			if err != nil {
				log.Fatal(err)
			}
			nonliveCount, err := nonlive.NumFound(query)
			if err != nil {
				log.Fatal(err)
			}
			if liveCount == 0 && nonliveCount == 0 {
				continue
			}
			name, ok := SourceNames[sid]
			if !ok {
				name = ""
			}
			fmt.Printf("%s\t%s\t%s\t%d\t%d\t%d\n",
				institution, sid, name, liveCount, nonliveCount, nonliveCount-liveCount)
		}
	}
}
