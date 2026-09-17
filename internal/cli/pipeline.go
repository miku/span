package cli

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/miku/span/internal/cmd/export"
	"github.com/miku/span/internal/cmd/oafilter"
	"github.com/miku/span/internal/cmd/redact"
	"github.com/miku/span/internal/cmd/reshape"
	"github.com/miku/span/internal/cmd/tag"
	"github.com/miku/span/internal/cmd/updatelabels"
	"github.com/miku/span/solrutil"
	"github.com/spf13/cobra"
)

// The commands in this file make up the main pipeline: import converts raw
// data into the intermediate schema, tag attaches ISILs, export writes SOLR
// documents. The others are filters over intermediate schema streams.

func newImportCmd() *cobra.Command {
	var (
		cfg     = reshape.Config{BatchSize: 10000, NumWorkers: runtime.NumCPU()}
		list    bool
		logfile string
		prof    profile
	)
	cmd := &cobra.Command{
		Use:   "import [flags] [file...]",
		Short: "Convert raw metadata into the intermediate schema",
		Long: `Reshapes raw metadata from various source formats into the intermediate
schema. Use --list to see the available input formats.`,
		Example: "  span import -i crossref < slice.ndj > out.is",
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				for _, k := range reshape.FormatNames() {
					fmt.Fprintln(cmd.OutOrStdout(), k)
				}
				return nil
			}
			if logfile != "" {
				f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					return err
				}
				defer f.Close()
				slog.SetDefault(slog.New(slog.NewJSONHandler(f, nil)))
			}
			r, closeInputs, err := openInputs(args, cmd.InOrStdin())
			if err != nil {
				return err
			}
			defer closeInputs()
			return prof.run(func() error {
				w := bufio.NewWriter(cmd.OutOrStdout())
				return errors.Join(reshape.Run(cfg, r, w), w.Flush())
			})
		},
	}
	f := cmd.Flags()
	f.StringVarP(&cfg.Name, "format", "i", "", "input format name")
	f.BoolVar(&list, "list", false, "list input formats")
	f.IntVarP(&cfg.NumWorkers, "workers", "w", cfg.NumWorkers, "number of workers")
	f.IntVarP(&cfg.BatchSize, "batch-size", "b", cfg.BatchSize, "batch size")
	f.StringVar(&logfile, "logfile", "", "path to logfile to append to, otherwise stderr")
	f.BoolVar(&cfg.Verbose, "verbose", false, "be verbose")
	prof.addFlags(f, true)
	return cmd
}

func newTagCmd() *cobra.Command {
	var (
		cfg                      = tag.DefaultConfig()
		config, unfreeze, expand string
		prof                     profile
	)
	cmd := &cobra.Command{
		Use:   "tag [flags] [file...]",
		Short: "Attach institution labels (ISIL) to intermediate schema records",
		Long: `Runs a forest of filters (from a JSON config, -c, or a frozen filterconfig,
--unfreeze) over every intermediate schema record to attach institution (ISIL)
tags. Can optionally query SOLR to deduplicate on the fly.`,
		Example: `  span tag -c '{"DE-15": {"any": {}}}' < input.is > output.is
  span tag --unfreeze filterconfig.zip input.is > output.is`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config == "" && unfreeze == "" {
				return errors.New("config file required")
			}
			if cfg.Server != "" {
				cfg.Server = solrutil.PrependHTTP(cfg.Server)
			}
			return prof.run(func() error {
				tagger, cleanup, err := tag.LoadTagger(config, unfreeze, expand, cfg.Verbose)
				if err != nil {
					return err
				}
				defer cleanup()
				r, closeInputs, err := openInputs(args, cmd.InOrStdin())
				if err != nil {
					return err
				}
				defer closeInputs()
				return tag.Run(cfg, tagger, r, cmd.OutOrStdout())
			})
		},
	}
	f := cmd.Flags()
	f.StringVarP(&config, "config", "c", "", "JSON config file for filters")
	f.IntVarP(&cfg.BatchSize, "batch-size", "b", cfg.BatchSize, "batch size")
	f.IntVarP(&cfg.NumWorkers, "workers", "w", cfg.NumWorkers, "number of workers")
	f.StringVar(&unfreeze, "unfreeze", "", "unfreeze filterconfig from a frozen file")
	f.BoolVar(&cfg.Verbose, "verbose", false, "verbose output")
	f.StringVar(&cfg.Server, "server", "", "if not empty, query SOLR to deduplicate on-the-fly")
	f.StringVar(&cfg.Prefs, "prefs", cfg.Prefs, "most preferred source id first, for deduplication")
	f.BoolVar(&cfg.IgnoreSameIdentifier, "isi", false, "when doing deduplication, ignore matches in index with the same id")
	f.BoolVarP(&cfg.DropDangling, "drop-dangling", "D", false, "drop dangling documents that do not have any isil attached")
	f.StringVar(&expand, "expand", "", "JSON file mapping meta-ISILs to lists of ISILs to expand into")
	prof.addFlags(f, true)
	return cmd
}

func newExportCmd() *cobra.Command {
	var (
		cfg  = export.DefaultConfig()
		list bool
		prof profile
	)
	cmd := &cobra.Command{
		Use:   "export [flags] [file...]",
		Short: "Convert intermediate schema records into SOLR documents",
		Long: `Converts intermediate schema records into a destination format, mostly SOLR
import documents. Use --list to see the available output formats.`,
		Example: "  span export --with-fullrecord -o solr5vu3 < tagged.is > solr.ndj",
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				fmt.Fprintln(cmd.OutOrStdout(), strings.Join(export.FormatNames(), "\n"))
				return nil
			}
			r, closeInputs, err := openInputs(args, cmd.InOrStdin())
			if err != nil {
				return err
			}
			defer closeInputs()
			return prof.run(func() error {
				return export.Run(cfg, r, cmd.OutOrStdout())
			})
		},
	}
	f := cmd.Flags()
	f.IntVarP(&cfg.BatchSize, "batch-size", "b", cfg.BatchSize, "batch size")
	f.IntVarP(&cfg.NumWorkers, "workers", "w", cfg.NumWorkers, "number of workers")
	f.StringVarP(&cfg.Format, "format", "o", cfg.Format, "output format")
	f.BoolVar(&list, "list", false, "list output formats")
	f.BoolVar(&cfg.WithFullrecord, "with-fullrecord", false, "populate fullrecord field with originating intermediate schema record")
	prof.addFlags(f, true)
	return cmd
}

func newRedactCmd() *cobra.Command {
	cfg := redact.DefaultConfig()
	cmd := &cobra.Command{
		Use:   "redact [flags] [file...]",
		Short: "Clear the fulltext field of intermediate schema records",
		Long: `Redacts intermediate schema records by clearing the fulltext field. Like jq's
del, but parallel and faster.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, closeInputs, err := openInputs(args, cmd.InOrStdin())
			if err != nil {
				return err
			}
			defer closeInputs()
			return redact.Run(cfg, r, cmd.OutOrStdout())
		},
	}
	cmd.Flags().IntVarP(&cfg.BatchSize, "batch-size", "b", cfg.BatchSize, "batch size")
	cmd.Flags().IntVarP(&cfg.NumWorkers, "workers", "w", cfg.NumWorkers, "number of workers")
	return cmd
}

func newUpdateLabelsCmd() *cobra.Command {
	var (
		cfg                  = updatelabels.DefaultConfig()
		labelFile, separator string
	)
	cmd := &cobra.Command{
		Use:   "update-labels [flags] < input",
		Short: "Set x.labels from a mapping of IDs to ISILs",
		Long: `Updates each intermediate schema record's x.labels field from a mapping of
IDs to ISILs. The mapping is held in memory. Without --file, nothing is done.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if labelFile == "" {
				return nil
			}
			f, err := os.Open(labelFile)
			if err != nil {
				return err
			}
			defer f.Close()
			labelMap, err := updatelabels.LoadLabelMap(f, separator)
			if err != nil {
				return err
			}
			w := bufio.NewWriter(cmd.OutOrStdout())
			return errors.Join(updatelabels.Run(cfg, labelMap, cmd.InOrStdin(), w), w.Flush())
		},
	}
	f := cmd.Flags()
	f.StringVarP(&labelFile, "file", "f", "", "path to comma separated file with ID and ISIL")
	f.StringVarP(&separator, "separator", "s", ",", "separator value")
	f.IntVarP(&cfg.BatchSize, "batch-size", "b", cfg.BatchSize, "batch size")
	f.IntVarP(&cfg.NumWorkers, "workers", "w", cfg.NumWorkers, "number of workers")
	return cmd
}

func newOAFilterCmd() *cobra.Command {
	cfg := oafilter.Config{BatchSize: 5000, BatchMemoryLimit: 209715200}
	cmd := &cobra.Command{
		Use:   "oa-filter [flags] < input",
		Short: "Mark records as open access based on a KBART file",
		Long: `Marks records as open access (sets x.oa to true) when a given KBART holding
file validates the record.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return oafilter.Run(cfg, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	f := cmd.Flags()
	f.StringVarP(&cfg.KbartFile, "kbart", "f", "", "path to a single KBART file")
	f.StringVar(&cfg.FreeContentFile, "fc", "", "path to a .../list?do=freeContent AMSL response JSON file")
	f.IntVarP(&cfg.BatchSize, "batch-size", "b", cfg.BatchSize, "batch size")
	f.BoolVar(&cfg.Verbose, "verbose", false, "extended output")
	f.Int64VarP(&cfg.BatchMemoryLimit, "batch-memory-limit", "m", cfg.BatchMemoryLimit, "memory limit per batch")
	f.BoolVarP(&cfg.BestEffort, "best-effort", "B", false, "ignore unmarshaling errors")
	f.StringArrayVar(&cfg.ExcludeSids, "xsid", nil, "exclude a given SID from checks, x.oa will always be false (repeatable)")
	f.StringArrayVar(&cfg.OpenAccessSids, "oasid", nil, "always set x.oa true for a given sid (repeatable)")
	return cmd
}
