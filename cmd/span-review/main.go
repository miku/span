// span-review runs plausibility queries against a SOLR server, mostly facet
// queries, refs #12756.
//
// TODO:
//
// * configurable
// * fetch from gitlab
// * run
// * post to gitlab or redmine
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/miku/span/solrutil"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// defaultConfig as baseline and documentation.
var defaultConfig = `
# Allowed keys: [Query, Facet-Field, Value, ...] checks if all values of field
# contain only given values.
allowed-keys:
    - ["source_id:30", "format", "eBook", "ElectronicArticle"]
    - ["source_id:30", "format_de15", "Book, eBook", "Article, E-Article"]
    - ["source_id:48", "language", "German", "English"]
    - ["source_id:49", "facet_avail", "Online", "Free"]
    - ["source_id:55", "facet_avail", "Online", "Free"]

# All records: [Query, Facet-Field, Value, ...] checks if all record contain
# only the given values.
all-records:
    - ["source_id:28", "format", "ElectronicArticle"]
    - ["source_id:28", "format_de15", "Article, E-Article"]
    - ["source_id:28", "facet_avail", "Online", "Free"]
    - ["source_id:28", "access_facet", "Electronic Resources"]
    - ["source_id:28", "mega_collection", "DOAJ Directory of Open Access Journals"]
    - ["source_id:28", "finc_class_facet", "not assigned"]
    - ["source_id:30", "facet_avail", "Online", "Free"]
    - ["source_id:30", "access_facet", "Electronic Resources"]
    - ["source_id:30", "mega_collection", "SSOAR Social Science Open Access Repository"]
    - ["source_id:34", "format", "ElectronicThesis"]
    - ["source_id:34", "format_de15", "Thesis"]
    - ["source_id:34", "facet_avail", "Online", "Free"]
    - ["source_id:34", "access_facet", "Electronic Resources"]
    - ["source_id:34", "mega_collection", "PQDT Open"]
    - ["source_id:48", "format", "ElectronicArticle"]
    - ["source_id:48", "format_de15", "Article, E-Article"]
    - ["source_id:48", "facet_avail", "Online"]
    - ["source_id:48", "access_facet", "Electronic Resources"]
    - ["source_id:49", "facet_avail", "Online"]
    - ["source_id:49", "access_facet", "Electronic Resources"]
    - ["source_id:49", "language", "English"]
    - ["source_id:50", "format", "ElectronicArticle"]
    - ["source_id:50", "format_de15", "Article, E-Article"]
    - ["source_id:50", "facet_avail", "Online"]
    - ["source_id:50", "access_facet", "Electronic Resources"]
    - ["source_id:50", "mega_collection", "DeGruyter SSH"]
    - ["source_id:53", "format", "ElectronicArticle"]
    - ["source_id:53", "format_de15", "Article, E-Article"]
    - ["source_id:53", "facet_avail", "Online"]
    - ["source_id:53", "access_facet", "Electronic Resources"]
    - ["source_id:53", "mega_collection", "CEEOL Central and Eastern European Online Library"]
    - ["source_id:55", "format", "ElectronicArticle"]
    - ["source_id:55", "format_de15", "Article, E-Article"]
    - ["source_id:55", "facet_avail", "Online"]
    - ["source_id:55", "access_facet", "Electronic Resources"]
    - ["source_id:60", "format", "ElectronicArticle"]
    - ["source_id:60", "format_de15", "Article, E-Article"]
    - ["source_id:60", "facet_avail", "Online"]
    - ["source_id:60", "access_facet", "Electronic Resources"]
    - ["source_id:60", "mega_collection", "Thieme E-Journals"]
    - ["source_id:60", "facet_avail", "Online"]
    - ["source_id:85", "format", "ElectronicArticle"]
    - ["source_id:85", "format_de15", "Article, E-Article"]
    - ["source_id:85", "facet_avail", "Online"]
    - ["source_id:85", "access_facet", "Electronic Resources"]
    - ["source_id:85", "language", "English"]
    - ["source_id:85", "mega_collection", "Elsevier Journals"]
    - ["source_id:87", "format", "ElectronicArticle"]
    - ["source_id:87", "format_de15", "Article, E-Article"]
    - ["source_id:87", "facet_avail", "Online", "Free"]
    - ["source_id:87", "access_facet", "Electronic Resources"]
    - ["source_id:87", "language", "English"]
    - ["source_id:87", "mega_collection", "International Journal of Communication"]
    - ["source_id:89", "format", "ElectronicArticle"]
    - ["source_id:89", "format_de15", "Article, E-Article"]
    - ["source_id:89", "facet_avail", "Online"]
    - ["source_id:89", "access_facet", "Electronic Resources"]
    - ["source_id:89", "language", "English"]
    - ["source_id:89", "mega_collection", "IEEE Xplore Library"]
    - ["source_id:101", "format", "ElectronicArticle"]
    - ["source_id:101", "format_de15", "Article, E-Article"]
    - ["source_id:101", "facet_avail", "Online"]
    - ["source_id:101", "access_facet", "Electronic Resources"]
    - ["source_id:101", "mega_collection", "Kieler Beiträge zur Filmmusikforschung"]
    - ["source_id:101", "finc_class_facet", "not assigned"]
    - ["source_id:105", "format", "ElectronicArticle"]
    - ["source_id:105", "format_de15", "Article, E-Article"]
    - ["source_id:105", "facet_avail", "Online"]
    - ["source_id:105", "access_facet", "Electronic Resources"]
    - ["source_id:105", "mega_collection", "Springer Journals"]
    - ["source_id:105", "finc_class_facet", "not assigned"]

# MinRatio: Query, Facet-Field, Value, Ratio (Percent), checks if the given
# value appears in a given percentage of documents.
min-ratio:
    - ["source_id:49", "facet_avail", "Free", 0.8]
    - ["source_id:55", "facet_avail", "Free", 2.2]
    - ["source_id:105", "facet_avail", "Free", 0.5]

# MinCount: Query, Facet-Field, Value, Min Count. Checks, if the given value
# appears at least a fixed number of times.
min-count:
    - ["source_id:89", "facet_avail", "Free", 50]
`

const (
	Check = "\u2713"
	Cross = "\u274C"
)

var (
	server     = flag.String("server", "http://localhost:8983/solr/biblio", "location of SOLR server")
	textile    = flag.Bool("t", false, "emit a textile table")
	ascii      = flag.Bool("a", false, "emit ascii table")
	configFile = flag.String("c", "", "path to review.yaml config file")
)

// prependHTTP prepends http, if necessary.
func prependHTTP(s string) string {
	if !strings.HasPrefix(s, "http") {
		return fmt.Sprintf("http://%s", s)
	}
	return s
}

// ReviewConfig contains various index review cases.
type ReviewConfig struct {
	AllowedKeys [][]string `yaml:"allowed-keys"`
	AllRecords  [][]string `yaml:"all-records"`
	MinRatio    [][]string `yaml:"min-ratio"`
	MinCount    [][]string `yaml:"min-count"`
}

// Result represents a single result row.
type Result struct {
	SourceIdentifier string
	Link             string
	SolrField        string
	FixedResult      bool
	Passed           bool
	Comment          string
}

// TextileResultWriter converts Results to Textile markup.
type TextileResultWriter struct {
	w io.Writer
}

// NewTextileTableWriter creates a new markup writer.
func NewTextileTableWriter(w io.Writer) *TextileResultWriter {
	return &TextileResultWriter{w: w}
}

// WriteHeader writes a header.
func (w *TextileResultWriter) WriteHeader() (int, error) {
	return io.WriteString(w.w,
		"| *Source ID, Field* | *Fixed* | *Passed* | *Comment* |\n")
}

// WriteResult writes a single Result.
func (w *TextileResultWriter) WriteResult(r Result) (int, error) {
	f, p := Check, Check
	if !r.FixedResult {
		f = Cross
	}
	if !r.Passed {
		p = Cross
	}
	return fmt.Fprintf(w.w, "| \"%s %s\":%s | %v | %v | %v |\n",
		r.SourceIdentifier, r.SolrField, r.Link, f, p, r.Comment)
}

// WriteResults writes a batch of results.
func (w *TextileResultWriter) WriteResults(rs []Result) (int, error) {
	bw := 0
	if n, err := w.WriteHeader(); err != nil {
		return 0, err
	} else {
		bw += n
	}
	for _, r := range rs {
		if n, err := w.WriteResult(r); err != nil {
			return 0, err
		} else {
			bw += n
		}
	}
	return bw, nil
}

// Given a query string, parse out the source identifier, panics currently, if
// query is not of the form source_id:23.
func MustParseSourceIdentifier(s string) string {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		panic(fmt.Sprintf("failed to parse source id query: %s", s))
	}
	return parts[1]
}

// ErrorOrComment returns error or message if error is nil.
func ErrorOrComment(err error, message string) string {
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return message
}

func main() {
	flag.Parse()

	index := solrutil.Index{Server: prependHTTP(*server)}

	var results []Result
	var err error

	var configReader io.Reader
	if *configFile == "" {
		log.Println("using default config, similar to https://git.io/fNfSk")
		configReader = strings.NewReader(defaultConfig)
	} else {
		f, err := os.Open(*configFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		configReader = f
	}

	// Grab config.
	var config ReviewConfig
	if yaml.NewDecoder(configReader).Decode(&config) != nil {
		log.Fatal(err)
	}

	// Cases like "access_facet:"Electronic Resources" für alle Records".
	// Multiple values are alternatives.
	for _, c := range config.AllowedKeys {
		if len(c) < 3 {
			log.Fatal("invalid test case, too few fields: %s", c)
		}
		query, field, values := c[0], c[1], c[2:]
		if err = index.AllowedKeys(query, field, values...); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(query),
			Link:             index.FacetLink(query, field),
			SolrField:        field,
			FixedResult:      true,
			Passed:           err == nil,
			Comment:          ErrorOrComment(err, fmt.Sprintf("%s %s %s", query, field, values)),
		})
	}

	// Cases like "facet_avail:Online UND facet_avail:Free für alle Records".
	// All records must have one or more facet values.
	for _, c := range config.AllRecords {
		if len(c) < 3 {
			log.Fatal("invalid test case, too few fields: %s", c)
		}
		query, field, values := c[0], c[1], c[2:]
		if err = index.EqualSizeTotal(query, field, values...); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(query),
			Link:             index.FacetLink(query, field),
			SolrField:        field,
			FixedResult:      true,
			Passed:           err == nil,
			Comment:          ErrorOrComment(err, fmt.Sprintf("%s %s %s", query, field, values)),
		})
	}

	// Cases like "facet_avail:Free für mindestens 0,5% aller Records".
	for _, c := range config.MinRatio {
		if len(c) != 4 {
			log.Fatal("invalid test case, expected four fields: %s", c)
		}
		query, field, value := c[0], c[1], c[2]
		minRatioPct, err := strconv.ParseFloat(c[3], 64)
		if err != nil {
			log.Fatal("minRatio is not a float: %s", err)
		}
		if err = index.MinRatioPct(query, field, value, minRatioPct); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(query),
			Link:             index.FacetLink(query, field),
			SolrField:        field,
			FixedResult:      true,
			Passed:           err == nil,
			Comment: ErrorOrComment(err, fmt.Sprintf("%s %s %s %0.4f",
				query, field, value, minRatioPct)),
		})
	}

	// Cases like "facet_avail:Free für mindestens 50 Records".
	for _, c := range config.MinCount {
		if len(c) != 4 {
			log.Fatal("invalid test case, expected four fields: %s", c)
		}
		query, field, value := c[0], c[1], c[2]
		minCount, err := strconv.Atoi(c[3])
		if err != nil {
			log.Fatal("minCount is not an int: %s", err)
		}
		if err = index.MinCount(query, field, value, minCount); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(query),
			Link:             index.FacetLink(query, field),
			SolrField:        field,
			FixedResult:      true,
			Passed:           err == nil,
			Comment: ErrorOrComment(err, fmt.Sprintf("%s %s %s %d",
				query, field, value, minCount)),
		})
	}

	if *textile {
		tw := NewTextileTableWriter(os.Stdout)
		if _, err := tw.WriteResults(results); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if *ascii {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		red, green := color.New(color.FgRed), color.New(color.FgGreen)
		for i, r := range results {
			passed := green.Sprintf("ok")
			if !r.Passed {
				passed = red.Sprintf("X")
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%v\t%s\t\n", i,
				r.SourceIdentifier, r.SolrField, passed, r.Comment)
		}
		w.Flush()
		os.Exit(0)
	}
}
