package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/miku/span/solrutil"
)

// queryFlags captures every flag known to the `query` subcommand. Flags are
// grouped into filters (what to query) and breakdowns (how to summarise).
//
// The breakdowns are mutually exclusive: at most one of --size, --by-*,
// --field, --missing, --has, or one of the named shortcut facets may be set.
// When none is given, --size is implied.
type queryFlags struct {
	// filters
	q           string
	sid         string
	collection  string
	language    string
	format      string
	institution string
	since       string
	after       string
	before      string

	// breakdowns
	size      bool
	byField   string // value of --by-<NAME>; empty means none
	field     string // explicit --field for facet
	missing   string
	has       string
	limit     int
	shortcut  string // name of the shortcut flag (e.g. "formats")
	shortField string // resolved facet field for the shortcut
}

// shortcutFacets maps human-friendly --<flag> names to the underlying Solr
// facet field. Each generates a boolean flag with no value.
var shortcutFacets = []struct {
	flag, field string
}{
	{"formats", "format"},
	{"record-formats", "record_format"},
	{"languages", "language"},
	{"collections", "mega_collection"},
	{"publishers", "publisher"},
	{"institutions", "institution"},
	{"authors", "author_facet"},
	{"topics", "topic_facet"},
	{"journals", "container_title"},
	{"series", "series"},
	{"availability", "facet_avail"},
	{"urls", "url"},
	{"dewey", "dewey-raw"},
	{"publish-years", "publishDate"},
}

// byFields lists the fields exposed via --by-<name>.
var byFields = []struct {
	flag, field string
}{
	{"by-sid", "source_id"},
	{"by-source", "source_id"},
	{"by-format", "format"},
	{"by-record-format", "record_format"},
	{"by-language", "language"},
	{"by-collection", "mega_collection"},
	{"by-publisher", "publisher"},
	{"by-institution", "institution"},
	{"by-year", "publishDate"},
}

func runQuery(args []string) error {
	fs, server := newFlagSet("query")
	qf := &queryFlags{}

	// filters
	fs.StringVar(&qf.q, "q", "", "raw Solr query (overrides field filters)")
	fs.StringVar(&qf.sid, "sid", "", "filter by source_id")
	fs.StringVar(&qf.collection, "collection", "", "filter by mega_collection")
	fs.StringVar(&qf.language, "language", "", "filter by language")
	fs.StringVar(&qf.format, "format", "", "filter by format")
	fs.StringVar(&qf.institution, "institution", "", "filter by institution (ISIL)")
	fs.StringVar(&qf.since, "since", "", "filter last_indexed >= DATE (ISO or git-like, e.g. 1.day.ago)")
	fs.StringVar(&qf.after, "after", "", "filter publishDate > DATE")
	fs.StringVar(&qf.before, "before", "", "filter publishDate < DATE")

	// breakdowns
	fs.BoolVar(&qf.size, "size", false, "report numFound for the (filtered) query")
	fs.StringVar(&qf.field, "field", "", "facet on an arbitrary field")
	fs.StringVar(&qf.missing, "missing", "", "count docs where FIELD is empty")
	fs.StringVar(&qf.has, "has", "", "count docs where FIELD is non-empty")
	fs.IntVar(&qf.limit, "limit", 0, "limit number of facet rows (0 = unlimited)")

	// shortcut facet flags (boolean)
	shortBools := make(map[string]*bool, len(shortcutFacets))
	for _, sc := range shortcutFacets {
		shortBools[sc.flag] = fs.Bool(sc.flag, false, "facet: "+sc.field)
	}
	// --by-<name> facet flags (boolean)
	byBools := make(map[string]*bool, len(byFields))
	for _, b := range byFields {
		byBools[b.flag] = fs.Bool(b.flag, false, "facet by "+b.field)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Resolve the breakdown. Reject duplicates within each group.
	for _, sc := range shortcutFacets {
		if *shortBools[sc.flag] {
			if qf.shortcut != "" {
				return fmt.Errorf("multiple shortcut facet flags given: --%s and --%s", qf.shortcut, sc.flag)
			}
			qf.shortcut = sc.flag
			qf.shortField = sc.field
		}
	}
	for _, b := range byFields {
		if *byBools[b.flag] {
			if qf.byField != "" {
				return fmt.Errorf("multiple --by-* flags given")
			}
			qf.byField = b.field
		}
	}

	// Validate mutual exclusion. Build a list of which breakdown switches were set.
	chosen := []string{}
	if qf.size {
		chosen = append(chosen, "--size")
	}
	if qf.field != "" {
		chosen = append(chosen, "--field")
	}
	if qf.missing != "" {
		chosen = append(chosen, "--missing")
	}
	if qf.has != "" {
		chosen = append(chosen, "--has")
	}
	if qf.shortcut != "" {
		chosen = append(chosen, "--"+qf.shortcut)
	}
	if qf.byField != "" {
		chosen = append(chosen, "--by-*")
	}

	// --size + a --by-* / shortcut facet is allowed (both mean "facet on field"),
	// the --size flag is purely cosmetic in that combination. Flag it as not a
	// conflict.
	if qf.size && (qf.shortcut != "" || qf.byField != "") {
		// remove --size from chosen; the facet flag wins
		chosen = filter(chosen, "--size")
	}
	if len(chosen) > 1 {
		return fmt.Errorf("breakdown flags are mutually exclusive: %s", strings.Join(chosen, ", "))
	}

	// Build the filter query.
	q, err := buildQuery(qf)
	if err != nil {
		return err
	}

	idx := indexFor(*server)
	switch {
	case qf.missing != "":
		return printNumFound(idx, fmt.Sprintf("-%s:[* TO *]", qf.missing))
	case qf.has != "":
		return printNumFound(idx, fmt.Sprintf("%s:[* TO *]", qf.has))
	case qf.field != "":
		return printFacet(idx, q, qf.field, qf.limit)
	case qf.shortField != "":
		return printFacet(idx, q, qf.shortField, qf.limit)
	case qf.byField != "":
		return printFacet(idx, q, qf.byField, qf.limit)
	default:
		// --size or no breakdown flag: total count
		return printNumFound(idx, q)
	}
}

// buildQuery assembles a Solr query string from the filter flags. If --q was
// given it is used verbatim; otherwise the field filters are joined with AND.
// An empty filter set yields "*:*".
func buildQuery(qf *queryFlags) (string, error) {
	if qf.q != "" {
		return qf.q, nil
	}
	var clauses []string
	add := func(field, value string) {
		if value != "" {
			clauses = append(clauses, fmt.Sprintf("%s:%q", field, value))
		}
	}
	add("source_id", qf.sid)
	add("mega_collection", qf.collection)
	add("language", qf.language)
	add("format", qf.format)
	add("institution", qf.institution)

	if qf.since != "" {
		t, err := parseDate(qf.since)
		if err != nil {
			return "", fmt.Errorf("--since: %w", err)
		}
		clauses = append(clauses, fmt.Sprintf("last_indexed:[%s TO *]", formatSolrDate(t)))
	}
	if qf.after != "" {
		t, err := parseDate(qf.after)
		if err != nil {
			return "", fmt.Errorf("--after: %w", err)
		}
		clauses = append(clauses, fmt.Sprintf("publishDateSort:[%d TO *]", t.Year()))
	}
	if qf.before != "" {
		t, err := parseDate(qf.before)
		if err != nil {
			return "", fmt.Errorf("--before: %w", err)
		}
		clauses = append(clauses, fmt.Sprintf("publishDateSort:[* TO %d]", t.Year()))
	}
	if len(clauses) == 0 {
		return "*:*", nil
	}
	return strings.Join(clauses, " AND "), nil
}

func printNumFound(idx solrutil.Index, q string) error {
	n, err := idx.NumFound(q)
	if err != nil {
		return err
	}
	fmt.Println(n)
	return nil
}

func printFacet(idx solrutil.Index, query, field string, limit int) error {
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
			return b.count - a.count
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

func filter(ss []string, drop string) []string {
	out := ss[:0]
	for _, s := range ss {
		if s != drop {
			out = append(out, s)
		}
	}
	return out
}
