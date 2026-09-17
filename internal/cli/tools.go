package cli

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/dchest/safefile"
	"github.com/klauspost/compress/zstd"
	"github.com/miku/span/folio"
	"github.com/miku/span/internal/cmd/compact"
	"github.com/miku/span/internal/cmd/doisniffer"
	spanfolio "github.com/miku/span/internal/cmd/folio"
	freezecmd "github.com/miku/span/internal/cmd/freeze"
	"github.com/miku/span/internal/cmd/localdata"
	"github.com/miku/span/internal/cmd/mailcmd"
	"github.com/sethgrid/pester"
	"github.com/spf13/cobra"
)

func newCompactCmd() *cobra.Command {
	var (
		cfg    = compact.DefaultConfig()
		output string
	)
	cmd := &cobra.Command{
		Use:   "compact [flags] [file]",
		Short: "Deduplicate an NDJSON stream on a field",
		Long: `Deduplicates an NDJSON stream on a chosen field, keeping one record per key
via a selectable strategy. Uses an external sort, so memory stays bounded even
for 10M-100M record inputs. Reads stdin or a file (.gz and .zst are
decompressed).

Strategies:

  first    keep the first record seen for a key
  last     keep the last record seen
  random   keep a uniformly random record (reservoir over the group)
  min      keep the record with the smallest --sort-key value
  max      keep the record with the largest --sort-key value`,
		Example: `  span compact --key id --strategy last input.ndj
  zstdcat big.ndj.zst | span compact --key doi --strategy max --sort-key indexed_at --numeric > out.ndj`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}
			var in io.Reader = cmd.InOrStdin()
			if len(args) == 1 {
				r, closeIn, err := openCompressed(args[0])
				if err != nil {
					return err
				}
				defer closeIn()
				in = r
			}
			if output == "" {
				return compact.Run(cfg, in, cmd.OutOrStdout())
			}
			f, err := os.Create(output)
			if err != nil {
				return err
			}
			return errors.Join(compact.Run(cfg, in, f), f.Close())
		},
	}
	f := cmd.Flags()
	f.StringVar(&cfg.KeyField, "key", cfg.KeyField, "JSON field to deduplicate on")
	f.StringVar(&cfg.SortField, "sort-key", "", "JSON field used by --strategy min|max")
	f.StringVar(&cfg.Strategy, "strategy", cfg.Strategy, "first|last|random|min|max")
	f.BoolVar(&cfg.NumericSort, "numeric", false, "treat --sort-key as numeric")
	f.StringVarP(&output, "output", "o", "", "output file (default: stdout)")
	f.StringVarP(&cfg.SortMem, "buffer-size", "S", cfg.SortMem, "sort -S memory buffer")
	f.StringVarP(&cfg.SortTmp, "temporary-directory", "T", "", "sort -T temp directory")
	return cmd
}

// openCompressed opens path for reading, decompressing by extension.
func openCompressed(path string) (io.Reader, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	switch {
	case strings.HasSuffix(path, ".gz"):
		gr, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, nil, err
		}
		return gr, func() { gr.Close(); f.Close() }, nil
	case strings.HasSuffix(path, ".zst"), strings.HasSuffix(path, ".zstd"):
		zr, err := zstd.NewReader(f)
		if err != nil {
			f.Close()
			return nil, nil, err
		}
		return zr, func() { zr.Close(); f.Close() }, nil
	default:
		return f, func() { f.Close() }, nil
	}
}

func newDoisnifferCmd() *cobra.Command {
	var (
		cfg             = doisniffer.DefaultConfig()
		noSkipUnmatched bool
	)
	cmd := &cobra.Command{
		Use:   "doisniffer [flags] < input",
		Short: "Find DOIs in VuFind SOLR JSON documents",
		Long: `Sniffs DOIs out of VuFind SOLR JSON documents and, optionally, annotates each
document with the DOI it found.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.SkipUnmatched = !noSkipUnmatched
			return doisniffer.Run(cfg, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	f := cmd.Flags()
	f.BoolVarP(&noSkipUnmatched, "no-skip-unmatched", "S", false, "do not skip unmatched documents")
	f.StringVarP(&cfg.UpdateKey, "update-key", "k", "doi_str_mv", "update key")
	f.StringVarP(&cfg.IdentifierKey, "identifier-key", "i", "id", "identifier key")
	f.StringVarP(&cfg.IgnoreKeys, "ignore-keys", "K", "barcode,dewey", "ignore keys (regexp), comma separated")
	f.IntVarP(&cfg.NumWorkers, "workers", "w", runtime.NumCPU(), "number of workers")
	f.IntVarP(&cfg.BatchSize, "batch-size", "b", 5000, "batch size")
	return cmd
}

func newLocalDataCmd() *cobra.Command {
	cfg := localdata.DefaultConfig()
	cmd := &cobra.Command{
		Use:   "local-data [flags] < input",
		Short: "Extract selected fields from a JSON stream",
		Long: `Extracts selected fields from a JSON stream, similar to jq but faster for this
narrow task.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return localdata.Run(cfg, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	cmd.Flags().IntVarP(&cfg.BatchSize, "batch-size", "b", cfg.BatchSize, "batch size")
	return cmd
}

func newMailCmd() *cobra.Command {
	var cfg mailcmd.Config
	var textfile, output string
	cmd := &cobra.Command{
		Use:   "mail [flags]",
		Short: "Send an email via SMTP or write the composed message to a file",
		Long: `Sends an email via SMTP or writes the composed message to a file. Without
--output the message is sent via the SMTP server named in SPAN_SMTP_SERVER
(host only; port 25 is used).`,
		Example: `  span mail -f sender@example.com -s "Test subject" \
      -t recipient1@example.com -t recipient2@example.com -b body.txt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.SMTPServer = os.Getenv("SPAN_SMTP_SERVER")
			if err := cfg.Validate(); err != nil {
				return err
			}
			if textfile == "" {
				return errors.New("-b/--textfile is required")
			}
			body, err := os.ReadFile(textfile)
			if err != nil {
				return fmt.Errorf("reading body file: %w", err)
			}
			cfg.Body = string(body)
			if output == "" {
				return mailcmd.Run(cfg, nil, mailcmd.DefaultSender(cfg.SMTPServer))
			}
			f, err := os.Create(output)
			if err != nil {
				return err
			}
			return errors.Join(mailcmd.Run(cfg, f, nil), f.Close())
		},
	}
	f := cmd.Flags()
	f.StringVarP(&cfg.From, "sender", "f", "", "the value of the From: header (required)")
	f.StringVarP(&cfg.Subject, "subject", "s", "", "the value of the Subject: header (required)")
	f.StringArrayVarP(&cfg.To, "recipient", "t", nil, "a To: header value (at least one required, repeatable)")
	f.StringVarP(&textfile, "textfile", "b", "", "the textfile for the body (required)")
	f.StringVarP(&output, "output", "o", "", "write the composed message to file instead of sending")
	return cmd
}

func newFolioCmd() *cobra.Command {
	var (
		cfg              = spanfolio.Config{}
		base, tenant, up string
	)
	cmd := &cobra.Command{
		Use:   "folio [flags]",
		Short: "Fetch metadata collections from the FOLIO API (debugging aid)",
		Long: `Talks to the FOLIO API to fetch ISIL, metadata collections and related
attachment information for a tenant. Work in progress.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if up != "" {
				var ok bool
				if cfg.User, cfg.Password, ok = strings.Cut(up, ":"); !ok {
					return errors.New("-u requires user:password")
				}
			}
			api := &folio.API{Base: base, Tenant: tenant, Client: pester.New()}
			return spanfolio.Run(cfg, api, cmd.OutOrStdout())
		},
	}
	f := cmd.Flags()
	f.StringVar(&base, "folio", "https://okapi.erm.staging.folio.finc.info", "folio endpoint")
	f.StringVar(&tenant, "tenant", "de_15", "folio tenant")
	f.IntVar(&cfg.Limit, "limit", 100000, "limit for lists")
	f.StringVar(&cfg.CQL, "cql", `(selectedBy=("*"))`, `cql query, e.g. (selectedBy=("DE-15")`)
	f.BoolVarP(&cfg.Raw, "raw", "r", false, "raw output")
	f.StringVarP(&up, "user", "u", "", "user:password for api")
	return cmd
}

func newFreezeCmd() *cobra.Command {
	var (
		legacy   freezecmd.LegacyConfig
		fcfg     = freezecmd.FolioConfig{OkapiURL: os.Getenv("OKAPI_URL")}
		output   string
		useFolio bool
	)
	cmd := &cobra.Command{
		Use:   "freeze [flags] -o output.zip",
		Short: "Freeze a filter configuration and all referenced files into a zip",
		Long: `Freezes a filter configuration and every file it references into a single zip,
so a tagging run can be reproduced offline. Reads a URL-bearing blob from stdin
(legacy) or builds the config from the FOLIO API (-f).

Output zip structure:

  /blob          original input or generated filterconfig JSON
  /mapping.json  URL to local path mapping
  /files/<sha1>  downloaded content`,
		Example: `  curl -s https://queue.acm.org/ | span freeze -b -o acm.zip
  OKAPI_TOKEN=xxx span freeze -f --okapi-url https://... -o folio.zip`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if output == "" {
				return errors.New("output file required")
			}
			if useFolio {
				fcfg.Token = os.Getenv("OKAPI_TOKEN")
				fcfg.Expand = legacy.Expand
				return freezecmd.RunFolio(fcfg, output)
			}
			client := http.DefaultClient
			if fcfg.NoProxy {
				client = &http.Client{
					Transport: &http.Transport{
						Proxy:                 nil,
						DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
						TLSHandshakeTimeout:   10 * time.Second,
						ResponseHeaderTimeout: 30 * time.Second,
					},
				}
			}
			file, err := safefile.Create(output, 0644)
			if err != nil {
				return err
			}
			defer file.Close()
			if err := freezecmd.RunLegacy(legacy, client, cmd.InOrStdin(), file); err != nil {
				return err
			}
			return file.Commit()
		},
	}
	f := cmd.Flags()
	f.StringVarP(&output, "output", "o", "", "output file")
	f.BoolVarP(&legacy.BestEffort, "best-effort", "b", false, "report errors but do not stop")
	f.BoolVarP(&useFolio, "folio", "f", false, "use FOLIO API instead of stdin")
	f.BoolVar(&fcfg.NoProxy, "no-proxy", false, "ignore system proxy settings")
	f.StringVar(&fcfg.OkapiURL, "okapi-url", fcfg.OkapiURL, "OKAPI base URL (env: OKAPI_URL)")
	f.StringVar(&fcfg.Tenant, "tenant", "de15", "FOLIO tenant")
	f.IntVar(&fcfg.Limit, "limit", 100000, "API pagination limit")
	f.StringVar(&legacy.Expand, "expand", "", "JSON or file mapping meta-ISILs to lists of ISILs to expand into")
	return cmd
}
