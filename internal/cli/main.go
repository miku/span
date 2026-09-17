// Package cli implements span's command line: one binary, one cobra command
// per verb.
//
// span used to install one executable per tool (span-import, span-tag, ...).
// The old names survive as symlinks to the single "span" binary. It looks at
// the name it was invoked under and runs the matching verb, so existing
// scripts, cron jobs and siskin tasks keep working unchanged.
package cli

import (
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/miku/span"
	"github.com/miku/span/internal/cmd/index"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// legacyNames maps each executable span used to install onto the command path
// that replaces it. The Makefile and nfpm.yaml lay down exactly these names
// as symlinks; a test keeps them in sync.
var legacyNames = map[string][]string{
	"span-compact":                {"compact"},
	"span-crossref-fast-snapshot": {"crossref", "fast-snapshot"},
	"span-crossref-fastproc":      {"crossref", "fastproc"},
	"span-crossref-members":       {"crossref", "members"},
	"span-crossref-snapshot":      {"crossref", "snapshot"},
	"span-crossref-sync":          {"crossref", "sync"},
	"span-crossref-table":         {"crossref", "table"},
	"span-doisniffer":             {"doisniffer"},
	"span-export":                 {"export"},
	"span-folio":                  {"folio"},
	"span-freeze":                 {"freeze"},
	"span-import":                 {"import"},
	"span-index":                  {"index"},
	"span-local-data":             {"local-data"},
	"span-mail":                   {"mail"},
	"span-oa-filter":              {"oa-filter"},
	"span-redact":                 {"redact"},
	"span-tag":                    {"tag"},
	"span-update-labels":          {"update-labels"},
}

// Main runs span and exits with a non-zero status on error.
func Main() {
	args, legacy := Dispatch(os.Args)
	root := NewRoot()
	root.SetArgs(RewriteArgs(root, args, legacy))
	if c, err := root.ExecuteC(); err != nil {
		log.Printf("%s: %v", c.CommandPath(), err)
		os.Exit(1)
	}
}

// NewRoot builds the command tree.
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "span",
		Short: "Convert, license and export library metadata",
		Long: `span converts metadata from various sources into an intermediate schema,
attaches institution labels and exports documents for SOLR.

Earlier versions installed one executable per tool (span-import, span-tag, ...).
These names are still installed as symlinks to this binary and take the flags
they always did.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	crossref := &cobra.Command{
		Use:   "crossref",
		Short: "Harvest, deduplicate and convert crossref data",
	}
	crossref.AddCommand(
		newCrossrefFastSnapshotCmd(),
		newCrossrefFastprocCmd(),
		newCrossrefMembersCmd(),
		newCrossrefSnapshotCmd(),
		newCrossrefSyncCmd(),
		newCrossrefTableCmd(),
	)
	root.AddCommand(
		newImportCmd(),
		newTagCmd(),
		newExportCmd(),
		newCompactCmd(),
		newDoisnifferCmd(),
		newFolioCmd(),
		newFreezeCmd(),
		newLocalDataCmd(),
		newMailCmd(),
		newOAFilterCmd(),
		newRedactCmd(),
		newUpdateLabelsCmd(),
		index.NewCommand(),
		crossref,
	)
	setVersion(root, span.AppVersion)
	root.SetVersionTemplate("{{.Version}}\n")
	return root
}

// setVersion gives every command its own --version flag (and -v, where the
// command does not use -v for something else), as most legacy tools had one.
func setVersion(cmd *cobra.Command, version string) {
	cmd.Version = version
	for _, sub := range cmd.Commands() {
		setVersion(sub, version)
	}
}

// Dispatch maps an invocation under a legacy name onto the command path that
// replaces it. argv is the whole command line, argv[0] included; it returns
// the args for the root command and whether a legacy name was used.
func Dispatch(argv []string) (args []string, legacy bool) {
	if len(argv) == 0 {
		return nil, false
	}
	name := strings.TrimSuffix(filepath.Base(argv[0]), ".exe")
	path, ok := legacyNames[name]
	if !ok {
		return argv[1:], false
	}
	return append(slices.Clone(path), argv[1:]...), true
}

// RewriteArgs bridges the stdlib flag package the legacy tools used and the
// pflag package cobra uses. With stdlib flags, "-unfreeze f" and "--unfreeze f"
// are the same thing, and so are "-o x" and "--o x". pflag reads "-unfreeze"
// as a cluster of shorthands and knows single letters only as shorthands, so
// long names get a second dash and single letters lose one.
//
// Under a legacy name every single-dash multi-letter token is rewritten, since
// the stdlib flag package had no clusters. Under "span" itself only known
// flags are rewritten, so that clusters like "-bf" keep working.
func RewriteArgs(root *cobra.Command, args []string, legacy bool) []string {
	target, _, err := root.Find(args)
	if err != nil || target == nil {
		return args
	}
	target.InitDefaultHelpFlag()
	target.InitDefaultVersionFlag()
	flags := pflag.NewFlagSet(target.Name(), pflag.ContinueOnError)
	flags.AddFlagSet(target.Flags())
	flags.AddFlagSet(target.InheritedFlags())
	return rewrite(args, flags, legacy)
}

// rewrite walks the command line one token at a time. Everything after "--"
// is left as it is, and so is the value of a flag that takes one.
func rewrite(args []string, flags *pflag.FlagSet, legacy bool) []string {
	out := make([]string, 0, len(args))
	var skip bool
	for i, arg := range args {
		if arg == "--" {
			return append(out, args[i:]...)
		}
		if skip {
			out, skip = append(out, arg), false
			continue
		}
		token, consumes := rewriteToken(arg, flags, legacy)
		out, skip = append(out, token), consumes
	}
	return out
}

// rewriteToken rewrites one token, reporting whether the flag it names will
// take the next token as its value.
func rewriteToken(arg string, flags *pflag.FlagSet, legacy bool) (string, bool) {
	if len(arg) < 2 || arg[0] != '-' {
		return arg, false // A positional or "-" itself.
	}
	dashes := 1
	if arg[1] == '-' {
		dashes = 2
	}
	name, hasValue := arg[dashes:], false
	if eq := strings.IndexByte(name, '='); eq >= 0 {
		name, hasValue = name[:eq], true
	}
	if name == "" {
		return arg, false
	}
	if len(name) == 1 {
		// "--o" was fine with stdlib flags; pflag only knows "-o", unless
		// the long name itself is a single letter.
		if f := flags.ShorthandLookup(name); f != nil {
			return arg[dashes-1:], !hasValue && takesValue(f)
		}
		if f := flags.Lookup(name); f != nil {
			return "--" + arg[dashes:], !hasValue && takesValue(f)
		}
		return arg, false
	}
	if dashes == 2 {
		return arg, !hasValue && takesValue(flags.Lookup(name))
	}
	if f := flags.Lookup(name); f != nil {
		return "-" + arg, !hasValue && takesValue(f)
	}
	if legacy {
		return "-" + arg, false // Unknown, but it was never a cluster either.
	}
	return arg, false
}

// takesValue reports whether a flag consumes the next token. pflag marks the
// flags that do not, the booleans, with a default for bare use.
func takesValue(f *pflag.Flag) bool { return f != nil && f.NoOptDefVal == "" }
