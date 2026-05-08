package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

// runSelect handles "span-index select", which forwards a raw Solr query and
// streams matching documents as one JSON object per line.
func runSelect(args []string) error {
	fs, server := newFlagSet("select")
	q := fs.String("q", "*:*", "Solr query")
	rows := fs.Int("rows", 10, "max docs to return")
	sort := fs.String("sort", "", "Solr sort spec, e.g. \"last_indexed desc\"")
	fl := fs.String("fl", "", "comma-separated field list")
	start := fs.Int("start", 0, "offset into the result set")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *q == "" {
		return fmt.Errorf("-q is required")
	}
	vs := url.Values{}
	vs.Set("q", *q)
	vs.Set("rows", strconv.Itoa(*rows))
	vs.Set("start", strconv.Itoa(*start))
	vs.Set("wt", "json")
	if *sort != "" {
		vs.Set("sort", *sort)
	}
	if *fl != "" {
		vs.Set("fl", *fl)
	}
	resp, err := indexFor(*server).Select(vs)
	if err != nil {
		return err
	}
	for _, doc := range resp.Response.Docs {
		fmt.Fprintln(os.Stdout, string(doc))
	}
	return nil
}
