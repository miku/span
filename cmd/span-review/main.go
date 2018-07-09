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
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
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
# Review configuration, refs #12756.
#
# Proposed workflow:
#
# 1. Edit this file via GitLab at
# https://git.sc.uni-leipzig.de/miku/span/blob/master/docs/review.yaml. Add,
# edit or remove rules, update ticket number. If done, commit.
# 2. A trigger will run an index review based on these rules.
# 3. Find the results in your ticket, in case the ticket number was valid.

# The solr server to query, including scheme, port and collection, e.g.
# "http://localhost:8983/solr/biblio". If "auto", then the current testing solr
# server will be figured out automatically.
solr: "auto"

# The ticket number of update. Set this to "NA" or anything non-numeric to
# suppress ticket updates.
ticket: "NA"

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
	server     = flag.String("server", "", "location of SOLR server, overrides review.yaml")
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
	SolrServer  string     `yaml:"solr"`
	Ticket      string     `yaml:"ticket"`
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

// UserHomeDir returns the home directory of the user.
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// findTestingSolrServer tries to find the URL of the current testing solr.
// There might be different way to retrieve a useable URL (configuration,
// probes). For now we use a separate configuration file, that contains the URL
// to the nginx snippet.
func findTestingSolrServer() (string, error) {
	configFile := path.Join(UserHomeDir(), ".config/span/span.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(configFile), 0755); err != nil {
			return "", err
		}
		data := []byte(`{"gitlab.token": "xxx", "whatislive.url": "xxx"}`)
		if err := ioutil.WriteFile(configFile, data, 0600); err != nil {
			return "", err
		}
		return "", fmt.Errorf("created new config file, please adjust: %s", configFile)
	}
	f, err := os.Open(configFile)
	if err != nil {
		return "", err
	}
	var conf struct {
		WhatIsLiveURL string `json:"whatislive.url"`
	}
	if err := json.NewDecoder(f).Decode(&conf); err != nil {
		return "", err
	}
	resp, err := http.Get(conf.WhatIsLiveURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// Find hostport in nginx snippet: upstream solr_nonlive { server 10.1.1.10:8080; }.
	p := `(?m)upstream[\s]*solr_nonlive[\s]*{[\s]*server[\s]*([0-9:.]*)[\s]*;[\s]*}`
	matches := regexp.MustCompile(p).FindSubmatch(b)
	if matches == nil || len(matches) != 2 {
		return "", fmt.Errorf("cannot find solr server URL in nginx snippet: %s", string(b))
	}
	solrServer := fmt.Sprintf("%s/solr/biblio", string(matches[1]))
	return solrServer, nil
}

func main() {
	flag.Parse()

	// Read review configuration.
	var configReader io.Reader
	if *configFile == "" {
		log.Println("using default config")
		configReader = strings.NewReader(defaultConfig)
	} else {
		f, err := os.Open(*configFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		configReader = f
	}
	var config ReviewConfig
	if err := yaml.NewDecoder(configReader).Decode(&config); err != nil {
		log.Fatal(err)
	}

	// Find suitable server from config, "nginx snippet" or flag.
	var solrServer string
	var err error

	if strings.ToLower(config.SolrServer) == "auto" {
		solrServer, err = findTestingSolrServer()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("found test solr at %s", solrServer)
	} else {
		solrServer = config.SolrServer
	}
	if *server != "" {
		solrServer = *server
	}
	index := solrutil.Index{Server: prependHTTP(solrServer)}

	// Collect review results.
	var results []Result

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

	// Serialize.
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
