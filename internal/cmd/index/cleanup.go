package index

import (
	"fmt"
	"os"
	"strings"

	"github.com/segmentio/encoding/json"
)

// runCleanup handles "span-index cleanup". It NEVER mutates the index: it only
// writes a delete-by-query command to stdout for the operator to review and run.
//
// The command targets records by last_indexed (the same query the `query
// --until` count produces), optionally scoped by source_id, collection, etc.
// To avoid foot-guns it refuses to emit a command unless a time bound (--until)
// or a raw --q was given, so a bare invocation can never become "delete all".
func runCleanup(args []string) error {
	fs, server, debug := newFlagSet("cleanup")
	qf := &queryFlags{}
	fs.StringVar(&qf.q, "q", "", "raw Solr delete query (overrides the field filters below)")
	fs.StringVar(&qf.until, "until", "", "delete records with last_indexed <= DATE (ISO or git-like, e.g. 30.days.ago)")
	fs.StringVar(&qf.since, "since", "", "also require last_indexed >= DATE (lower bound of a window)")
	fs.StringVar(&qf.sid, "sid", "", "scope to source_id")
	fs.StringVar(&qf.collection, "collection", "", "scope to mega_collection")
	fs.StringVar(&qf.language, "language", "", "scope to language")
	fs.StringVar(&qf.format, "format", "", "scope to format")
	fs.StringVar(&qf.institution, "institution", "", "scope to institution (ISIL)")
	setExamples(fs,
		"span-index cleanup --until 2026-01-01",
		"span-index cleanup --until 30.days.ago --sid 53",
		`span-index cleanup --q "source_id:53 AND last_indexed:[* TO 2025-01-01T00:00:00Z]"`,
		"# review the printed command, then pipe to a shell to run it:",
		"span-index cleanup --until 2026-01-01 --sid 53 | sh",
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	// A cleanup without a time bound (or explicit raw query) would delete every
	// record matching the remaining filters regardless of index date — almost
	// never what's wanted. Require an explicit bound.
	if qf.until == "" && qf.q == "" {
		return fmt.Errorf("cleanup needs --until DATE (or a raw --q); refusing to build a delete without a time bound")
	}

	q, err := buildQuery(qf)
	if err != nil {
		return err
	}
	if strings.TrimSpace(q) == "*:*" {
		return fmt.Errorf("refusing to emit a delete-all command (query resolved to *:*)")
	}

	idx := indexFor(*server, *debug)

	// Show the blast radius as a comment. This is a read-only count; emitting the
	// command does not depend on it succeeding, but if the index is unreachable
	// the operator could not run the delete anyway, so we surface the error.
	n, err := idx.NumFound(q)
	if err != nil {
		return fmt.Errorf("counting matches for %q: %w", q, err)
	}

	cmd, err := renderCleanupCommand(idx.Server, q, n)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, cmd)
	return nil
}

// renderCleanupCommand builds the stdout text for a cleanup: a leading comment
// stating how many docs match, followed by a copy-paste-safe curl delete-by-query
// against the collection's update handler (with commit=true). The query is sent
// as a JSON delete body so any quotes inside it are correctly escaped, and the
// whole payload is single-quoted for the shell.
func renderCleanupCommand(server, q string, n int64) (string, error) {
	body := struct {
		Delete struct {
			Query string `json:"query"`
		} `json:"delete"`
	}{}
	body.Delete.Query = q
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	updateURL := fmt.Sprintf("%s/update?commit=true", server)
	return fmt.Sprintf("# %d docs match %s\ncurl %s \\\n  -H 'Content-Type: application/json' \\\n  -d %s",
		n, q, shellSingleQuote(updateURL), shellSingleQuote(string(payload))), nil
}

// shellSingleQuote wraps s in single quotes for safe use as one shell argument,
// escaping any embedded single quotes via the '\'' idiom. This keeps the emitted
// curl command copy-paste safe even when the query contains quotes or apostrophes.
func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
