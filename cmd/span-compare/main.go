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
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/miku/span/solrutil"
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
	liveServer = flag.String("a", "http://localhost:8983/solr/biblio",
		"live server location")
	nonliveServer = flag.String("b", "http://localhost:8983/solr/biblio",
		"non-live server location")
)

// prependHTTP prepends http, if necessary.
func prependHTTP(s string) string {
	if !strings.HasPrefix(s, "http") {
		return fmt.Sprintf("http://%s", s)
	}
	return s
}

func main() {
	flag.Parse()

	live := solrutil.Index{Server: prependHTTP(*liveServer)}
	nonlive := solrutil.Index{Server: prependHTTP(*nonliveServer)}

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
			numLive, err := live.NumFound(query)
			if err != nil {
				log.Fatal(err)
			}
			numNonlive, err := nonlive.NumFound(query)
			if err != nil {
				log.Fatal(err)
			}
			if numLive == 0 && numNonlive == 0 {
				continue
			}
			name, ok := SourceNames[sid]
			if !ok {
				name = "XXX: missing source name"
			}
			fmt.Printf("%s\t%s\t%s\t%d\t%d\t%d\n",
				institution, sid, name, numLive, numNonlive, numNonlive-numLive)
		}
	}
}
