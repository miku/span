// span-index runs prebuilt queries against a SOLR index and renders results.
//
// Usage:
//
//	span-index -t numdocs
//	span-index -t sources
//	span-index -t since -p "date:2024-01-01"
//	span-index -t formats -p "limit:10"
//	span-index -t docs -p "source_id:68"
//	span-index -l
package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/miku/span/solrutil"
)

var (
	server   = flag.String("s", "http://localhost:8983/solr/biblio", "SOLR server")
	taskName = flag.String("t", "", "task to run (use -l to see available tasks)")
	list     = flag.Bool("l", false, "list available tasks")
	pp       params
)

// params collects repeated -p flags into a slice.
type params []string

func (p *params) String() string { return strings.Join(*p, ", ") }
func (p *params) Set(v string) error {
	*p = append(*p, v)
	return nil
}

// parseParams turns []{key:value} into a map.
func parseParams(pp params) map[string]string {
	m := make(map[string]string)
	for _, p := range pp {
		k, v, ok := strings.Cut(p, ":")
		if ok {
			m[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return m
}

// task describes a named query with a renderer.
type task struct {
	Name string
	Help string
	Run  func(index solrutil.Index, p map[string]string) error
}

// tasks is the registry of available tasks.
var tasks = []task{
	{
		Name: "numdocs",
		Help: "total number of documents in the index",
		Run: func(index solrutil.Index, p map[string]string) error {
			n, err := index.NumFound("*:*")
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "sources",
		Help: "list source ids with document counts",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, "*:*", "source_id", p)
		},
	},
	{
		Name: "docs-by-source",
		Help: "number of docs for a given source_id; -p source_id:68",
		Run: func(index solrutil.Index, p map[string]string) error {
			sid, ok := p["source_id"]
			if !ok {
				return fmt.Errorf("parameter source_id required")
			}
			n, err := index.NumFound(fmt.Sprintf("source_id:%q", sid))
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "since",
		Help: "docs indexed since a date; -p date:2024-01-01T00:00:00Z",
		Run: func(index solrutil.Index, p map[string]string) error {
			date, ok := p["date"]
			if !ok {
				return fmt.Errorf("parameter date required (e.g. 2024-01-01T00:00:00Z)")
			}
			q := fmt.Sprintf("last_indexed:[%s TO *]", date)
			n, err := index.NumFound(q)
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "publishers",
		Help: "list publishers with counts",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "publisher", p)
		},
	},
	{
		Name: "formats",
		Help: "list all formats with counts",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "format", p)
		},
	},
	{
		Name: "record-formats",
		Help: "list all record formats with counts",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "record_format", p)
		},
	},
	{
		Name: "collections",
		Help: "list mega_collection names with counts",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "mega_collection", p)
		},
	},
	{
		Name: "dewey",
		Help: "list dewey-raw values with counts; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "dewey-raw", p)
		},
	},
	{
		Name: "publish-years",
		Help: "histogram of publishDate values; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "publishDate", p)
		},
	},
	{
		Name: "languages",
		Help: "list languages with counts",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "language", p)
		},
	},
	{
		Name: "institutions",
		Help: "list institutions (ISIL) with counts",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "institution", p)
		},
	},
	{
		Name: "urls",
		Help: "list URLs; -p limit:100",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "url", p)
		},
	},
	{
		Name: "facet",
		Help: "generic facet query; -p field:FIELDNAME [-p q:QUERY] [-p limit:N]",
		Run: func(index solrutil.Index, p map[string]string) error {
			field, ok := p["field"]
			if !ok {
				return fmt.Errorf("parameter field required")
			}
			return printFacet(index, queryFromParams(p), field, p)
		},
	},
	{
		Name: "numfound",
		Help: "run arbitrary query and return numFound; -p q:\"source_id:68 AND format:Article\"",
		Run: func(index solrutil.Index, p map[string]string) error {
			q := queryFromParams(p)
			n, err := index.NumFound(q)
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "select",
		Help: "run arbitrary select query; -p q:QUERY [-p rows:N]",
		Run: func(index solrutil.Index, p map[string]string) error {
			q := queryFromParams(p)
			rows := "10"
			if v, ok := p["rows"]; ok {
				rows = v
			}
			vs := url.Values{}
			vs.Set("q", q)
			vs.Set("rows", rows)
			vs.Set("wt", "json")
			resp, err := index.Select(vs)
			if err != nil {
				return err
			}
			for _, doc := range resp.Response.Docs {
				fmt.Println(string(doc))
			}
			return nil
		},
	},
	{
		Name: "authors",
		Help: "top authors with counts; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "author_facet", p)
		},
	},
	{
		Name: "topics",
		Help: "top subjects/topics with counts; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "topic_facet", p)
		},
	},
	{
		Name: "journals",
		Help: "top journal/container titles with counts; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "container_title", p)
		},
	},
	{
		Name: "series",
		Help: "list series with counts; -p limit:20",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "series", p)
		},
	},
	{
		Name: "availability",
		Help: "facet_avail breakdown (Online, Free, etc.)",
		Run: func(index solrutil.Index, p map[string]string) error {
			return printFacet(index, queryFromParams(p), "facet_avail", p)
		},
	},
	{
		Name: "missing",
		Help: "count docs where a field is empty; -p field:FIELDNAME",
		Run: func(index solrutil.Index, p map[string]string) error {
			field, ok := p["field"]
			if !ok {
				return fmt.Errorf("parameter field required")
			}
			q := fmt.Sprintf("-%s:[* TO *]", field)
			n, err := index.NumFound(q)
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "has",
		Help: "count docs where a field is non-empty; -p field:FIELDNAME",
		Run: func(index solrutil.Index, p map[string]string) error {
			field, ok := p["field"]
			if !ok {
				return fmt.Errorf("parameter field required")
			}
			q := fmt.Sprintf("%s:[* TO *]", field)
			n, err := index.NumFound(q)
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "source-collections",
		Help: "collections for a given source; -p source_id:49",
		Run: func(index solrutil.Index, p map[string]string) error {
			sid, ok := p["source_id"]
			if !ok {
				return fmt.Errorf("parameter source_id required")
			}
			return printFacet(index, fmt.Sprintf("source_id:%q", sid), "mega_collection", p)
		},
	},
	{
		Name: "source-formats",
		Help: "format breakdown for a given source; -p source_id:49",
		Run: func(index solrutil.Index, p map[string]string) error {
			sid, ok := p["source_id"]
			if !ok {
				return fmt.Errorf("parameter source_id required")
			}
			return printFacet(index, fmt.Sprintf("source_id:%q", sid), "format", p)
		},
	},
	{
		Name: "source-languages",
		Help: "language breakdown for a given source; -p source_id:49",
		Run: func(index solrutil.Index, p map[string]string) error {
			sid, ok := p["source_id"]
			if !ok {
				return fmt.Errorf("parameter source_id required")
			}
			return printFacet(index, fmt.Sprintf("source_id:%q", sid), "language", p)
		},
	},
	{
		Name: "recent",
		Help: "most recently indexed docs; -p rows:10",
		Run: func(index solrutil.Index, p map[string]string) error {
			rows := "10"
			if v, ok := p["rows"]; ok {
				rows = v
			}
			vs := url.Values{}
			vs.Set("q", queryFromParams(p))
			vs.Set("rows", rows)
			vs.Set("sort", "last_indexed desc")
			vs.Set("fl", "id,source_id,title,last_indexed")
			vs.Set("wt", "json")
			resp, err := index.Select(vs)
			if err != nil {
				return err
			}
			for _, doc := range resp.Response.Docs {
				fmt.Println(string(doc))
			}
			return nil
		},
	},
	{
		Name: "between",
		Help: "docs published between two years; -p from:2020 -p to:2024",
		Run: func(index solrutil.Index, p map[string]string) error {
			from, ok := p["from"]
			if !ok {
				return fmt.Errorf("parameter from required (e.g. 2020)")
			}
			to, ok := p["to"]
			if !ok {
				return fmt.Errorf("parameter to required (e.g. 2024)")
			}
			q := fmt.Sprintf("publishDate:[%s TO %s]", from, to)
			n, err := index.NumFound(q)
			if err != nil {
				return err
			}
			fmt.Println(n)
			return nil
		},
	},
	{
		Name: "institution-sources",
		Help: "source breakdown for a given institution (ISIL); -p institution:DE-14",
		Run: func(index solrutil.Index, p map[string]string) error {
			isil, ok := p["institution"]
			if !ok {
				return fmt.Errorf("parameter institution required (e.g. DE-14)")
			}
			return printFacet(index, fmt.Sprintf("institution:%q", isil), "source_id", p)
		},
	},
}

// queryFromParams builds a SOLR query from common parameter conventions. If a
// "q" parameter is given, use that directly. Otherwise combine known field
// parameters (source_id, format, mega_collection, etc.) with AND.
func queryFromParams(p map[string]string) string {
	if q, ok := p["q"]; ok {
		return q
	}
	var clauses []string
	for _, field := range []string{"source_id", "format", "record_format", "mega_collection", "language", "institution"} {
		if v, ok := p[field]; ok {
			clauses = append(clauses, fmt.Sprintf("%s:%q", field, v))
		}
	}
	if len(clauses) == 0 {
		return "*:*"
	}
	return strings.Join(clauses, " AND ")
}

// printFacet runs a facet query and prints the result as a tab-separated table
// sorted by count descending. Respects a "limit" parameter.
func printFacet(index solrutil.Index, query, field string, p map[string]string) error {
	limit := 0
	if v, ok := p["limit"]; ok {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid limit: %v", err)
		}
		limit = n
	}
	idx := index
	if limit > 0 {
		idx.FacetLimit = limit
	}
	resp, err := idx.FacetQuery(query, field)
	if err != nil {
		return err
	}
	fm, err := resp.Facets()
	if err != nil {
		return err
	}
	type entry struct {
		key   string
		count int
	}
	entries := make([]entry, 0, len(fm))
	for k, v := range fm {
		if v > 0 {
			entries = append(entries, entry{k, v})
		}
	}
	slices.SortFunc(entries, func(a, b entry) int {
		if a.count != b.count {
			return b.count - a.count // descending
		}
		return strings.Compare(a.key, b.key)
	})
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%d\n", e.key, e.count)
	}
	return w.Flush()
}

// resolveTask finds a task by exact name or unambiguous prefix.
func resolveTask(name string) (*task, error) {
	for i := range tasks {
		if tasks[i].Name == name {
			return &tasks[i], nil
		}
	}
	var matches []*task
	for i := range tasks {
		if strings.HasPrefix(tasks[i].Name, name) {
			matches = append(matches, &tasks[i])
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return nil, fmt.Errorf("unknown task: %s", name)
	default:
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return nil, fmt.Errorf("ambiguous task %q, matches: %s", name, strings.Join(names, ", "))
	}
}

func main() {
	flag.Var(&pp, "p", "parameter as key:value (repeatable)")
	flag.Parse()
	if *list {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, t := range tasks {
			fmt.Fprintf(w, "%s\t%s\n", t.Name, t.Help)
		}
		w.Flush()
		os.Exit(0)
	}
	if *taskName == "" {
		fmt.Fprintln(os.Stderr, "use -t NAME to run a task, or -l to see available tasks")
		os.Exit(1)
	}
	t, err := resolveTask(*taskName)
	if err != nil {
		log.Fatal(err)
	}
	index := solrutil.Index{Server: solrutil.PrependHTTP(*server)}
	if err := t.Run(index, parseParams(pp)); err != nil {
		log.Fatal(err)
	}
}
