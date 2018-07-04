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
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/miku/span/solrutil"
	log "github.com/sirupsen/logrus"
)

const (
	Check = "\u2713"
	Cross = "\u274C"
)

var (
	server  = flag.String("server", "http://localhost:8983/solr/biblio", "location of SOLR server")
	textile = flag.Bool("t", false, "emit a textile table")
	ascii   = flag.Bool("a", false, "emit ascii table")
)

// prependHTTP prepends http, if necessary.
func prependHTTP(s string) string {
	if !strings.HasPrefix(s, "http") {
		return fmt.Sprintf("http://%s", s)
	}
	return s
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

	// Cases like "access_facet:"Electronic Resources" für alle Records".
	// Multiple values are alternatives.
	allowedKeyCases := []struct {
		Query  string
		Field  string
		Values []string
	}{
		{"source_id:30", "format", []string{"eBook", "ElectronicArticle"}},
		{"source_id:30", "format_de15", []string{"Book, E-Book", "Article, E-Article"}},
		{"source_id:48", "language", []string{"German", "English"}},
		{"source_id:49", "facet_avail", []string{"Online", "Free"}},
		{"source_id:55", "facet_avail", []string{"Online", "Free"}},
	}

	for _, c := range allowedKeyCases {
		if err = index.AllowedKeys(c.Query, c.Field, c.Values...); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(c.Query),
			Link:             index.FacetLink(c.Query, c.Field),
			SolrField:        c.Field,
			FixedResult:      true,
			Passed:           err == nil,
			Comment:          ErrorOrComment(err, fmt.Sprintf("%s %s %s", c.Query, c.Field, c.Values)),
		})
	}

	// Cases like "facet_avail:Online UND facet_avail:Free für alle Records".
	// All records must have one or more facet values.
	allRecordsCases := []struct {
		Query  string
		Field  string
		Values []string
	}{
		{"source_id:28", "format", []string{"ElectronicArticle"}},
		{"source_id:28", "format_de15", []string{"Article, E-Article"}},
		{"source_id:28", "facet_avail", []string{"Online", "Free"}},
		{"source_id:28", "access_facet", []string{"Electronic Resources"}},
		{"source_id:28", "mega_collection", []string{"DOAJ Directory of Open Access Journals"}},
		{"source_id:28", "finc_class_facet", []string{"not assigned"}},

		{"source_id:30", "facet_avail", []string{"Online", "Free"}},
		{"source_id:30", "access_facet", []string{"Electronic Resources"}},
		{"source_id:30", "mega_collection", []string{"SSOAR Social Science Open Access Repository"}},

		{"source_id:34", "format", []string{"ElectronicThesis"}},
		{"source_id:34", "format_de15", []string{"Thesis"}},
		{"source_id:34", "facet_avail", []string{"Online", "Free"}},
		{"source_id:34", "access_facet", []string{"Electronic Resources"}},
		{"source_id:34", "mega_collection", []string{"PQDT Open"}},

		{"source_id:48", "format", []string{"ElectronicArticle"}},
		{"source_id:48", "format_de15", []string{"Article, E-Article"}},
		{"source_id:48", "facet_avail", []string{"Online"}},
		{"source_id:48", "access_facet", []string{"Electronic Resources"}},

		{"source_id:49", "facet_avail", []string{"Online"}},
		{"source_id:49", "access_facet", []string{"Electronic Resources"}},
		{"source_id:49", "language", []string{"English"}},

		{"source_id:50", "format", []string{"ElectronicArticle"}},
		{"source_id:50", "format_de15", []string{"Article, E-Article"}},
		{"source_id:50", "facet_avail", []string{"Online"}},
		{"source_id:50", "access_facet", []string{"Electronic Resources"}},
		{"source_id:50", "mega_collection", []string{"DeGruyter SSH"}},

		{"source_id:53", "format", []string{"ElectronicArticle"}},
		{"source_id:53", "format_de15", []string{"Article, E-Article"}},
		{"source_id:53", "facet_avail", []string{"Online"}},
		{"source_id:53", "access_facet", []string{"Electronic Resources"}},
		{"source_id:53", "mega_collection", []string{"CEEOL Central and Eastern European Online Library"}},

		{"source_id:55", "format", []string{"ElectronicArticle"}},
		{"source_id:55", "format_de15", []string{"Article, E-Article"}},
		{"source_id:55", "facet_avail", []string{"Online"}},
		{"source_id:55", "access_facet", []string{"Electronic Resources"}},

		{"source_id:60", "format", []string{"ElectronicArticle"}},
		{"source_id:60", "format_de15", []string{"Article, E-Article"}},
		{"source_id:60", "facet_avail", []string{"Online"}},
		{"source_id:60", "access_facet", []string{"Electronic Resources"}},
		{"source_id:60", "mega_collection", []string{"Thieme E-Journals"}},
		{"source_id:60", "facet_avail", []string{"Online"}},

		{"source_id:85", "format", []string{"ElectronicArticle"}},
		{"source_id:85", "format_de15", []string{"Article, E-Article"}},
		{"source_id:85", "facet_avail", []string{"Online"}},
		{"source_id:85", "access_facet", []string{"Electronic Resources"}},
		{"source_id:85", "language", []string{"English"}},
		{"source_id:85", "mega_collection", []string{"Elsevier Journals"}},

		{"source_id:87", "format", []string{"ElectronicArticle"}},
		{"source_id:87", "format_de15", []string{"Article, E-Article"}},
		{"source_id:87", "facet_avail", []string{"Online", "Free"}},
		{"source_id:87", "access_facet", []string{"Electronic Resources"}},
		{"source_id:87", "language", []string{"English"}},
		{"source_id:87", "mega_collection", []string{"International Journal of Communication"}},

		{"source_id:89", "format", []string{"ElectronicArticle"}},
		{"source_id:89", "format_de15", []string{"Article, E-Article"}},
		{"source_id:89", "facet_avail", []string{"Online"}},
		{"source_id:89", "access_facet", []string{"Electronic Resources"}},
		{"source_id:89", "language", []string{"English"}},
		{"source_id:89", "mega_collection", []string{"IEEE Xplore Library"}},

		{"source_id:101", "format", []string{"ElectronicArticle"}},
		{"source_id:101", "format_de15", []string{"Article, E-Article"}},
		{"source_id:101", "facet_avail", []string{"Online"}},
		{"source_id:101", "access_facet", []string{"Electronic Resources"}},
		{"source_id:101", "mega_collection", []string{"Kieler Beiträge zur Filmmusikforschung"}},
		{"source_id:101", "finc_class_facet", []string{"not assigned"}},

		{"source_id:105", "format", []string{"ElectronicArticle"}},
		{"source_id:105", "format_de15", []string{"Article, E-Article"}},
		{"source_id:105", "facet_avail", []string{"Online"}},
		{"source_id:105", "access_facet", []string{"Electronic Resources"}},
		{"source_id:105", "mega_collection", []string{"Springer Journals"}},
		{"source_id:105", "finc_class_facet", []string{"not assigned"}},
	}

	for _, c := range allRecordsCases {
		if err = index.EqualSizeTotal(c.Query, c.Field, c.Values...); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(c.Query),
			Link:             index.FacetLink(c.Query, c.Field),
			SolrField:        c.Field,
			FixedResult:      true,
			Passed:           err == nil,
			Comment:          ErrorOrComment(err, fmt.Sprintf("%s %s %s", c.Query, c.Field, c.Values)),
		})
	}

	// Cases like "facet_avail:Free für mindestens 0,5% aller Records".
	ratioCases := []struct {
		Query    string
		Field    string
		Value    string
		MinRatio float64
	}{
		{"source_id:49", "facet_avail", "Free", 0.8},
		{"source_id:55", "facet_avail", "Free", 2.2},
		{"source_id:105", "facet_avail", "Free", 0.5},
	}
	for _, c := range ratioCases {
		if err = index.MinRatioPct(c.Query, c.Field, c.Value, c.MinRatio); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(c.Query),
			Link:             index.FacetLink(c.Query, c.Field),
			SolrField:        c.Field,
			FixedResult:      true,
			Passed:           err == nil,
			Comment: ErrorOrComment(err,
				fmt.Sprintf("%s %s %s %0.4f", c.Query, c.Field, c.Value, c.MinRatio)),
		})
	}

	// Cases like "facet_avail:Free für mindestens 50 Records".
	minCountCases := []struct {
		Query    string
		Field    string
		Value    string
		MinCount int
	}{
		{"source_id:89", "facet_avail", "Free", 50},
	}
	for _, c := range minCountCases {
		if err = index.MinCount(c.Query, c.Field, c.Value, c.MinCount); err != nil {
			log.Println(err)
		}
		results = append(results, Result{
			SourceIdentifier: MustParseSourceIdentifier(c.Query),
			Link:             index.FacetLink(c.Query, c.Field),
			SolrField:        c.Field,
			FixedResult:      true,
			Passed:           err == nil,
			Comment: ErrorOrComment(err,
				fmt.Sprintf("%s %s %s %d", c.Query, c.Field, c.Value, c.MinCount)),
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
