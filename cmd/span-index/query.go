package main

import (
	"fmt"
	"net/url"
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
	until       string
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
	fs, server, debug := newFlagSet("query")
	qf := &queryFlags{}

	// filters
	fs.StringVar(&qf.q, "q", "", "raw Solr query (overrides field filters)")
	fs.StringVar(&qf.sid, "sid", "", "filter by source_id")
	fs.StringVar(&qf.collection, "collection", "", "filter by mega_collection")
	fs.StringVar(&qf.language, "language", "", "filter by language")
	fs.StringVar(&qf.format, "format", "", "filter by format")
	fs.StringVar(&qf.institution, "institution", "", "filter by institution (ISIL)")
	fs.StringVar(&qf.since, "since", "", "filter last_indexed >= DATE (ISO or git-like, e.g. 1.day.ago)")
	fs.StringVar(&qf.until, "until", "", "filter last_indexed <= DATE (ISO or git-like, e.g. 30.days.ago)")
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

	setExamples(fs,
		"span-index query --size",
		"span-index query --size --sid 49",
		"span-index query --by-sid",
		"span-index query --formats --sid 49",
		"span-index query --since 1.day.ago",
		"span-index query --until 30.days.ago --size",
		"span-index query --until 2026-01-01 --sid 53 --size",
		"span-index query --after 2026-01-01 --before 2026-02-01 --sid 49",
		"span-index query --missing doi",
		"span-index query --missing record_id --by-sid",
		"span-index query --has issn --sid 49",
		"span-index query --has issn --by-sid",
		`span-index query --q "source_id:49 AND format:Article" --size`,
	)

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

	// --missing and --has are filters (composed via fq), not breakdowns; they
	// may combine with --size or with a facet breakdown.
	breakdowns := []string{}
	if qf.size {
		breakdowns = append(breakdowns, "--size")
	}
	if qf.field != "" {
		breakdowns = append(breakdowns, "--field")
	}
	if qf.shortcut != "" {
		breakdowns = append(breakdowns, "--"+qf.shortcut)
	}
	if qf.byField != "" {
		breakdowns = append(breakdowns, "--by-*")
	}
	// --size + a --by-* / shortcut facet is allowed (both mean "facet on field"),
	// the --size flag is purely cosmetic in that combination.
	if qf.size && (qf.shortcut != "" || qf.byField != "") {
		breakdowns = filter(breakdowns, "--size")
	}
	if len(breakdowns) > 1 {
		return fmt.Errorf("breakdown flags are mutually exclusive: %s", strings.Join(breakdowns, ", "))
	}

	// Build the filter query.
	q, err := buildQuery(qf)
	if err != nil {
		return err
	}

	idx := indexFor(*server, *debug)

	// Compose fq clauses from --missing / --has. Pre-validate fields against
	// the schema so we fail fast with a useful message.
	var fqs []string
	if qf.missing != "" {
		if err := checkExistenceQueryable(idx, qf.missing); err != nil {
			return err
		}
		fqs = append(fqs, fmt.Sprintf("-%s:*", qf.missing))
	}
	if qf.has != "" {
		if err := checkExistenceQueryable(idx, qf.has); err != nil {
			return err
		}
		fqs = append(fqs, fmt.Sprintf("%s:*", qf.has))
	}

	// Pick the breakdown field, if any.
	facetField := ""
	switch {
	case qf.field != "":
		facetField = qf.field
	case qf.shortField != "":
		facetField = qf.shortField
	case qf.byField != "":
		facetField = qf.byField
	}

	if facetField != "" {
		return printFacetFq(idx, q, fqs, facetField, qf.limit)
	}
	return printNumFoundFq(idx, q, fqs)
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
	if qf.until != "" {
		t, err := parseDate(qf.until)
		if err != nil {
			return "", fmt.Errorf("--until: %w", err)
		}
		clauses = append(clauses, fmt.Sprintf("last_indexed:[* TO %s]", formatSolrDate(t)))
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

// checkExistenceQueryable verifies that name is defined in the Solr schema
// and that Solr can answer "field has a value / is missing" queries on it.
// Lucene's FieldExistsQuery requires docValues, norms, or termVectors.
func checkExistenceQueryable(idx solrutil.Index, name string) error {
	fields, err := idx.SchemaFieldsFull()
	if err != nil {
		return fmt.Errorf("schema lookup failed: %w", err)
	}
	var field *solrutil.SchemaField
	names := make([]string, 0, len(fields))
	for i := range fields {
		names = append(names, fields[i].Name)
		if fields[i].Name == name {
			field = &fields[i]
		}
	}
	if field == nil {
		var near []string
		lname := strings.ToLower(name)
		for _, f := range names {
			if strings.Contains(strings.ToLower(f), lname) {
				near = append(near, f)
			}
		}
		slices.Sort(names)
		if len(near) > 0 {
			slices.Sort(near)
			return fmt.Errorf("field %q is not in the schema; did you mean: %s", name, strings.Join(near, ", "))
		}
		return fmt.Errorf("field %q is not in the schema (have %d fields: %s)", name, len(names), strings.Join(names, ", "))
	}
	if !field.Indexed {
		return fmt.Errorf("field %q (type %s) is not indexed; cannot run existence queries against it.\n%s",
			name, field.Type, existenceCandidates(fields))
	}
	if !field.CanCheckExistence() {
		return fmt.Errorf("field %q (type %s) has no docValues, norms, or termVectors; Solr cannot count missing/has values for it. Enable docValues=true on the field (or its field type) in the schema to make this query work.\n%s",
			name, field.Type, existenceCandidates(fields))
	}
	return nil
}

// existenceCandidates renders a one-line summary of fields that *can* be used
// with --missing / --has, so the user knows what to try instead.
func existenceCandidates(fields []solrutil.SchemaField) string {
	var ok []string
	for _, f := range fields {
		if f.CanCheckExistence() {
			ok = append(ok, f.Name)
		}
	}
	if len(ok) == 0 {
		return "no fields in this schema support missing/has queries"
	}
	slices.Sort(ok)
	return fmt.Sprintf("fields that support missing/has queries (%d): %s", len(ok), strings.Join(ok, ", "))
}

// printNumFoundFq runs q with zero or more filter queries (fq) and prints the
// resulting numFound. Filter queries are used for negations like
// -field:* to sidestep the pure-negative-query trap of Solr's standard parser.
func printNumFoundFq(idx solrutil.Index, q string, fqs []string) error {
	vals := url.Values{}
	vals.Add("q", q)
	for _, fq := range fqs {
		vals.Add("fq", fq)
	}
	vals.Add("rows", "0")
	vals.Add("wt", "json")
	resp, err := idx.Select(vals)
	if err != nil {
		return err
	}
	fmt.Println(resp.Response.NumFound)
	return nil
}

// printFacetFq runs a facet on field over q with optional filter queries (fq).
// Used so --by-* / --field / shortcut facets compose with --missing and --has.
func printFacetFq(idx solrutil.Index, q string, fqs []string, field string, limit int) error {
	vals := url.Values{}
	vals.Add("q", q)
	for _, fq := range fqs {
		vals.Add("fq", fq)
	}
	vals.Add("facet", "true")
	vals.Add("facet.field", field)
	facetLimit := limit
	if facetLimit <= 0 {
		facetLimit = solrutil.DefaultFacetLimit
	}
	vals.Add("facet.limit", fmt.Sprintf("%d", facetLimit))
	vals.Add("rows", "0")
	vals.Add("wt", "json")
	resp, err := idx.Select(vals)
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
