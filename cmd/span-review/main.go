// span-review runs plausibility queries against a SOLR server, mostly facet
// queries, refs #12756.
//
// There is a yaml file for configuring queries. It is possible to send results
// directly to a Redmine ticket. This program can be used standalone, or via
// span-webhookd.
//
// Additional rules: Ansigelung, sigeltest.
//
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/miku/span"
	"github.com/miku/span/reviewutil"
	"github.com/miku/span/solrutil"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// defaultConfig as baseline and documentation. Consider switching to
// https://github.com/elastic/go-ucfg.
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

# If set to "fail" an empty result set will be marked as failure.
# Otherwise a empty result set will - most of the time - not be considered a violoation.
zero-results-policy: "fail"

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
	server         = flag.String("server", "", "location of SOLR server, overrides review.yaml")
	textile        = flag.Bool("t", false, "emit a textile table to stdout")
	ascii          = flag.Bool("a", false, "emit ascii table to stdout")
	reviewFile     = flag.String("c", "", "path to review.yaml file containing test cases, e.g. https://git.io/fh5Zh")
	spanConfigFile = flag.String("span-config", path.Join(span.UserHomeDir(), ".config/span/span.json"), "gitlab, redmine tokens, whatislive location")
	ticket         = flag.String("ticket", "", "post result to redmine, overrides review.yaml, requires redmine.baseurl and redmine.apitoken configured in span-config")
	noCollapse     = flag.Bool("C", false, "do not collapse details")
)

// Result represents a single result row. XXX: Maybe add more fields, e.g.
// number of results in this test case.
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

// WriteResult writes a single Result. XXX: Make this more generic, accepting
// an interface.
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

// WriteResults writes a batch of results. XXX: Make this a bit more generic.
func (w *TextileResultWriter) WriteResults(rs []Result) (written int, err error) {
	if n, err := w.WriteHeader(); err != nil {
		return 0, err
	} else {
		written += n
	}
	for _, r := range rs {
		if n, err := w.WriteResult(r); err != nil {
			return 0, err
		} else {
			written += n
		}
	}
	return written, nil
}

// Given a query string, parse out the source identifier, panics currently, if
// query is not of the form source_id:23. TODO(miku): make this more flexible.
func MustParseSourceIdentifier(s string) string {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		panic(fmt.Sprintf("failed to parse source id query: %s", s))
	}
	return parts[1]
}

// ErrorOrMessage returns error or message if error is nil.
func ErrorOrMessage(err error, message string) string {
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return message
}

func main() {
	flag.Parse()

	// Fallback configuration, since daemon home is /usr/sbin.
	if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
		*spanConfigFile = "/etc/span/span.json"
	}
	// XXX: Use a real framework like go-ucfg or globalconf.
	if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
		log.Fatal("no configuration found, put one into /etc/span/span.json")
	}

	// Read review configuration.
	var configReader io.Reader
	if *reviewFile == "" {
		log.Println("no file given, using default review config (https://git.io/fNUD4)")
		configReader = strings.NewReader(defaultConfig)
	} else {
		f, err := os.Open(*reviewFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		configReader = f
	}

	var config reviewutil.ReviewConfig

	dec := yaml.NewDecoder(configReader)
	dec.SetStrict(true)
	if err := dec.Decode(&config); err != nil {
		log.Fatal(err)
	}

	// Find suitable server from config, "nginx snippet" or flag.
	var solrServer string
	var err error

	if strings.ToLower(config.SolrServer) == "auto" {
		solrServer, err = solrutil.FindNonliveSolrServer(*spanConfigFile)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Be a bit more flexible and handle the default admin interface URL,
		// e.g. http://example.com/solr/#/biblio, as well.
		if strings.Contains(config.SolrServer, "/#/") {
			log.Printf("adjusting SOLR admin URL %s", config.SolrServer)
			config.SolrServer = strings.Replace(config.SolrServer, "/#", "", -1)
		}
		solrServer = config.SolrServer

		// A URL like http://solr.web:8080/solr/biblio//select?q=*:* would 404 (with 5.5.5).
		solrServer = strings.TrimRight(solrServer, "/")
	}
	if *server != "" {
		solrServer = *server
	}
	log.Printf("using solr at %s", solrServer)
	if *ticket != "" {
		log.Printf("will attempt to update ticket %s", *ticket)
	}
	log.Printf("%d/%d/%d/%d", len(config.AllowedKeys), len(config.AllRecords), len(config.MinRatio), len(config.MinCount))
	index := solrutil.Index{Server: solrutil.PrependHTTP(solrServer)}

	// Collect review results.
	var results []Result

	// Cases like "access_facet:"Electronic Resources" für alle Records".
	// Multiple values are alternatives.
	for _, c := range config.AllowedKeys {
		if len(c) < 3 {
			log.Fatalf("invalid test case, too few fields: %s", c)
		}
		query, field, values := c[0], c[1], c[2:]
		if err = index.AllowedKeys(query, field, values...); err != nil {
			log.Println(err)
		}
		passed := (err == nil)

		numFound, err := index.NumFound(query)
		if err != nil {
			log.Fatal(err)
		}
		if numFound == 0 && config.ZeroResultsPolicy == "fail" {
			passed = false
		}

		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(query),
			Link:             index.FacetLink(query, field),
			SolrField:        field,
			FixedResult:      true,
			Passed:           passed,
			Comment:          ErrorOrMessage(err, fmt.Sprintf("%s %s %s", query, field, values)),
		})
	}

	// Cases like "facet_avail:Online UND facet_avail:Free für alle Records".
	// All records must have one or more facet values.
	for _, c := range config.AllRecords {
		if len(c) < 3 {
			log.Fatalf("invalid test case, too few fields: %s", c)
		}
		query, field, values := c[0], c[1], c[2:]
		if err = index.EqualSizeTotal(query, field, values...); err != nil {
			log.Println(err)
		}
		passed := (err == nil)

		numFound, err := index.NumFound(query)
		if err != nil {
			log.Fatal(err)
		}
		if numFound == 0 && config.ZeroResultsPolicy == "fail" {
			passed = false
		}

		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(query),
			Link:             index.FacetLink(query, field),
			SolrField:        field,
			FixedResult:      true,
			Passed:           passed,
			Comment:          ErrorOrMessage(err, fmt.Sprintf("%s %s %s", query, field, values)),
		})
	}

	// Cases like "facet_avail:Free für mindestens 0,5% aller Records".
	for _, c := range config.MinRatio {
		if len(c) != 4 {
			log.Fatalf("invalid test case, expected four fields: %s", c)
		}
		query, field, value := c[0], c[1], c[2]
		minRatioPct, err := strconv.ParseFloat(c[3], 64)
		if err != nil {
			log.Fatalf("minRatio is not a float: %s", err)
		}
		if err = index.MinRatioPct(query, field, value, minRatioPct); err != nil {
			log.Println(err)
		}
		passed := (err == nil)

		numFound, err := index.NumFound(query)
		if err != nil {
			log.Fatal(err)
		}
		if numFound == 0 && config.ZeroResultsPolicy == "fail" {
			passed = false
		}

		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(query),
			Link:             index.FacetLink(query, field),
			SolrField:        field,
			FixedResult:      true,
			Passed:           passed,
			Comment: ErrorOrMessage(err, fmt.Sprintf("%s %s %s %0.4f",
				query, field, value, minRatioPct)),
		})
	}

	// Cases like "facet_avail:Free für mindestens 50 Records".
	for _, c := range config.MinCount {
		if len(c) != 4 {
			log.Fatalf("invalid test case, expected four fields: %s", c)
		}
		query, field, value := c[0], c[1], c[2]
		minCount, err := strconv.Atoi(c[3])
		if err != nil {
			log.Fatalf("minCount is not an int: %s", err)
		}
		if err = index.MinCount(query, field, value, minCount); err != nil {
			log.Println(err)
		}
		passed := (err == nil)

		numFound, err := index.NumFound(query)
		if err != nil {
			log.Fatal(err)
		}
		if numFound == 0 && config.ZeroResultsPolicy == "fail" {
			passed = false
		}

		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(query),
			Link:             index.FacetLink(query, field),
			SolrField:        field,
			FixedResult:      true,
			Passed:           passed,
			Comment: ErrorOrMessage(err, fmt.Sprintf("%s %s %s %d",
				query, field, value, minCount)),
		})
	}

	// Serialization options.
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

	// Ticket handling.
	if *ticket != "" {
		config.Ticket = *ticket
	}
	if config.Ticket != "" {
		if _, err := strconv.Atoi(config.Ticket); err != nil {
			log.Printf("ignoring ticket update for non-numeric ticket id: %s", config.Ticket)
			os.Exit(0)
		}
		// Fallback configuration, since daemon home is /usr/sbin.
		if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
			*spanConfigFile = "/etc/span/span.json"
		}
		if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
			log.Printf("warning: no span config file found, thing might break")
		}
		f, err := os.Open(*spanConfigFile)
		if err != nil {
			log.Printf("failed to open span config: %s", err)
		}
		var conf struct {
			BaseURL string `json:"redmine.baseurl"`
			Token   string `json:"redmine.apitoken"`
		}
		if err := json.NewDecoder(f).Decode(&conf); err != nil {
			log.Fatal(err)
		}
		// Collect message for ticket.
		var buf bytes.Buffer

		hostname, err := os.Hostname()
		if err != nil {
			hostname = "an unidentified host"
		}
		fmt.Fprintf(&buf, "* tested SOLR at %s\n", index.Server)
		fmt.Fprintf(&buf, "* ran span-review %s on %s with review %s\n\n", span.AppVersion, hostname, *reviewFile)

		tw := NewTextileTableWriter(&buf)
		if _, err := tw.WriteResults(results); err != nil {
			log.Fatal(err)
		}

		redmine := &reviewutil.Redmine{BaseURL: conf.BaseURL, Token: conf.Token}
		var message string
		switch {
		case *noCollapse:
			message = fmt.Sprintf("index review results\n\n%s", buf.String())
		default:
			message = fmt.Sprintf("index review results\n\n{{collapse(Details)\n%s\n}}\n", buf.String())
		}
		if err := redmine.UpdateTicket(config.Ticket, message); err != nil {
			log.Fatal(err)
		}
	}
}
