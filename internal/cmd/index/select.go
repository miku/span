package index

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// runSelect handles "span-index select", which forwards a raw Solr query and
// streams matching documents as one JSON object per line.
func newSelectCmd(c *common) *cobra.Command {
	cmd := &cobra.Command{Use: "select", Short: "Raw SOLR query, prints JSON docs", Args: cobra.NoArgs}
	fs := cmd.Flags()
	q := fs.StringP("query", "q", "*:*", "Solr query")
	rows := fs.Int("rows", 10, "max docs to return")
	sort := fs.String("sort", "", "Solr sort spec, e.g. \"last_indexed desc\"")
	fl := fs.String("fl", "", "comma-separated field list")
	start := fs.Int("start", 0, "offset into the result set")
	cmd.Example = examples(
		`span index select -q "source_id:49 AND format:Article"`,
		`span index select -q "*:*" --rows 1 --fl id,title`,
		`span index select -q "*:*" --sort "last_indexed desc" --rows 5`,
		`span index select -q "issn:1234-5678" --rows 100 --start 100`,
	)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if *q == "" {
			return fmt.Errorf("--query is required")
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
		resp, err := c.index().Select(vs)
		if err != nil {
			return err
		}
		for _, doc := range resp.Response.Docs {
			fmt.Fprintln(os.Stdout, string(doc))
		}
		return nil
	}
	return cmd
}
