// Package folio implements the core of span-folio: it talks to the FOLIO API
// to fetch metadata collections and renders them for inspection.
package folio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"text/tabwriter"

	"github.com/miku/span/folio"
	"github.com/miku/span/strutil"
)

// Client is the subset of the FOLIO API used by this command. It is satisfied
// by *folio.API and can be stubbed in tests.
type Client interface {
	Authenticate(username, password string) error
	MetadataCollections(opts folio.MetadataCollectionsOpts) (*folio.MetadataCollectionsResponse, error)
}

// Config holds the tunables for a folio run.
type Config struct {
	User     string
	Password string
	CQL      string
	Limit    int
	Raw      bool
}

// Run authenticates against the FOLIO API, fetches metadata collections and
// renders them to w.
func Run(cfg Config, client Client, w io.Writer) error {
	if cfg.User == "" || cfg.Password == "" {
		return errors.New("incomplete credentials")
	}
	if err := client.Authenticate(cfg.User, cfg.Password); err != nil {
		return err
	}
	log.Println("[ok] auth")
	resp, err := client.MetadataCollections(folio.MetadataCollectionsOpts{
		CQL:   cfg.CQL,
		Limit: cfg.Limit,
	})
	if err != nil {
		return err
	}
	return render(resp, cfg.Raw, w)
}

// render writes the metadata collections response to w, either as raw JSON
// (one object per line) or as a truncated tab-separated table.
func render(resp *folio.MetadataCollectionsResponse, raw bool, w io.Writer) error {
	if raw {
		for _, v := range resp.FincConfigMetadataCollections {
			b, err := json.Marshal(v)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(w, string(b)); err != nil {
				return err
			}
		}
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
	for _, entry := range resp.FincConfigMetadataCollections {
		fmt.Fprintf(tw, "%s\t%s\t%s\n",
			strutil.Truncate(entry.Label, 40),
			strutil.Truncate(strings.Join(entry.SolrMegaCollections, ", "), 40),
			strutil.Truncate(entry.MdSource.Name, 40))
	}
	return tw.Flush()
}
