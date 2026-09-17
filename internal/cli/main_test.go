package cli

import (
	"bytes"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestDispatch(t *testing.T) {
	cases := []struct {
		argv   []string
		want   []string
		legacy bool
	}{
		{nil, nil, false},
		{[]string{"span", "tag", "-c", "x"}, []string{"tag", "-c", "x"}, false},
		{[]string{"/usr/local/bin/span-tag", "-c", "x"}, []string{"tag", "-c", "x"}, true},
		{[]string{"span-crossref-sync", "-p", "zstd"}, []string{"crossref", "sync", "-p", "zstd"}, true},
		{[]string{"span-index", "query", "--size"}, []string{"index", "query", "--size"}, true},
		{[]string{"span-unknown", "x"}, []string{"x"}, false},
	}
	for _, c := range cases {
		got, legacy := Dispatch(c.argv)
		if !slices.Equal(got, c.want) || legacy != c.legacy {
			t.Errorf("Dispatch(%q) = %q, %v, want %q, %v", c.argv, got, legacy, c.want, c.legacy)
		}
	}
}

func TestRewriteArgs(t *testing.T) {
	cases := []struct {
		args   []string
		legacy bool
		want   []string
	}{
		{
			[]string{"tag", "-unfreeze", "f.zip", "-D", "in.is"}, true,
			[]string{"tag", "--unfreeze", "f.zip", "-D", "in.is"},
		},
		{
			// A value that looks like a flag is left alone.
			[]string{"tag", "-prefs", "-x", "-v"}, true,
			[]string{"tag", "--prefs", "-x", "-v"},
		},
		{
			[]string{"oa-filter", "-B", "-b", "25000", "-f", "k.tsv", "-fc", "fc.json", "-xsid", "48", "-oasid", "28", "-oasid", "30"}, true,
			[]string{"oa-filter", "-B", "-b", "25000", "-f", "k.tsv", "--fc", "fc.json", "--xsid", "48", "--oasid", "28", "--oasid", "30"},
		},
		{
			[]string{"crossref", "fast-snapshot", "-cache=false", "--o", "out.zst", "a.zst"}, true,
			[]string{"crossref", "fast-snapshot", "--cache=false", "-o", "out.zst", "a.zst"},
		},
		{
			// Help and version flags are known before cobra adds them.
			[]string{"crossref", "members", "-version"}, false,
			[]string{"crossref", "members", "--version"},
		},
		{
			// Inherited flags count, too.
			[]string{"index", "query", "-server", "http://x", "-size"}, false,
			[]string{"index", "query", "--server", "http://x", "--size"},
		},
		{
			// A single letter long name.
			[]string{"index", "query", "-q", "*:*"}, false,
			[]string{"index", "query", "--q", "*:*"},
		},
		{
			// Unknown multi-letter tokens are only rewritten for legacy names.
			[]string{"export", "-bogus"}, false,
			[]string{"export", "-bogus"},
		},
		{
			[]string{"export", "-bogus"}, true,
			[]string{"export", "--bogus"},
		},
		{
			[]string{"export", "--", "-with-fullrecord"}, true,
			[]string{"export", "--", "-with-fullrecord"},
		},
	}
	for _, c := range cases {
		got := RewriteArgs(NewRoot(), c.args, c.legacy)
		if !slices.Equal(got, c.want) {
			t.Errorf("RewriteArgs(%q, %v) = %q, want %q", c.args, c.legacy, got, c.want)
		}
	}
}

// TestLegacyNamesResolve checks that every legacy name leads to a runnable
// command.
func TestLegacyNamesResolve(t *testing.T) {
	root := NewRoot()
	for _, name := range slices.Sorted(maps.Keys(legacyNames)) {
		cmd, rest, err := root.Find(legacyNames[name])
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if len(rest) > 0 || cmd == root || !cmd.HasSubCommands() && !cmd.Runnable() {
			t.Errorf("%s: resolves to %q, rest %q", name, cmd.CommandPath(), rest)
		}
	}
}

// TestPackagingLegacyNames keeps the symlinks laid down by the Makefile and
// the packages in sync with the dispatch table.
func TestPackagingLegacyNames(t *testing.T) {
	want := slices.Sorted(maps.Keys(legacyNames))
	b, err := os.ReadFile("../../Makefile")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?s)LEGACY = \\\n(.*?)\n\n`).FindSubmatch(b)
	if m == nil {
		t.Fatal("LEGACY not found in Makefile")
	}
	got := strings.Fields(strings.ReplaceAll(string(m[1]), `\`, ""))
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("Makefile LEGACY = %q, want %q", got, want)
	}
	b, err = os.ReadFile("../../nfpm.yaml")
	if err != nil {
		t.Fatal(err)
	}
	got = nil
	for _, m := range regexp.MustCompile(`dst: /usr/local/bin/(span-[a-z-]+)\n\s+type: symlink`).FindAllSubmatch(b, -1) {
		got = append(got, string(m[1]))
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("nfpm.yaml symlinks = %q, want %q", got, want)
	}
}

// TestLegacyInvocations parses command lines as siskin and the docs use them,
// without running them.
func TestLegacyInvocations(t *testing.T) {
	cases := []string{
		"span-import -i crossref in.json",
		"span-import -w 2 -i crossref in.json",
		"span-tag -unfreeze f.zip in.is",
		"span-tag in.is -unfreeze f.zip",
		"span-tag -D -unfreeze f.zip in.is",
		"span-tag -c amsl.json -server localhost:8983 -prefs 49 -isi",
		"span-export -with-fullrecord -o solr5vu3",
		"span-export -o solr5vu3 in.is",
		"span-export -cpuprofile cpu.pprof -memprofile mem.pprof",
		"span-doisniffer -S",
		"span-update-labels -b 20000 -f labels.csv -s ,",
		"span-redact -b 100 -w 2 in.is",
		"span-oa-filter -B -b 25000 -f k.tsv -fc fc.json -xsid 48 -oasid 28 -oasid 30 -oasid 34 -m 1000 -verbose",
		"span-local-data -b 100",
		"span-freeze -b -f -expand e.json -o out.zip -no-proxy -okapi-url http://x -tenant de15 -limit 10",
		"span-folio -folio http://x -tenant t -limit 1 -cql q -r -u user:pass",
		"span-mail -f a@b -s subject -t c@d -t e@f -b body.txt -o out.txt",
		"span-compact -T /tmp -key id -sort-key date -strategy max -numeric -S 1G -o out.ndj in.ndj",
		"span-crossref-sync -p zstd -P feed-1- -i d -verbose -t 30m -s 2022-01-01 -e 2023-05-01 -c /tmp -ua x -mode s -x 3 -r 10 -q -debug",
		"span-crossref-snapshot -b 100000 -verbose -z -compress-program zstd -o out.zst -x ex.txt -E 10 -S 25% in.zst",
		"span-crossref-fast-snapshot -o out.zst -v -S 50% -X ex.txt -R -k -n 10 -w 2 -cache=false -cache-dir /tmp -cache-clear a.zst b.zst",
		"span-crossref-fastproc -f f.zip -o - -w 2 -b 10 -expand e.json -okapi-url http://x -tenant t -no-proxy=false -cache-ttl 1h -force a.zst",
		"span-crossref-members -offset 10 -rows 5 -base http://x -sleep 2s -q -retry 1",
		"span-crossref-table -b 10 -w 2",
		"span-index query --size --sid 49",
		"span-index query --q source_id:49 --size -s http://x",
		"span-index select -q *:* --rows 1 --fl id,title",
		"span-index compare --file 49.ldj --all --textile",
		"span-index report --name recent --rows 20",
		"span-index report --name issn --sid 49 --collection c --verbose",
		"span-index cleanup --until 2026-01-01 --sid 53",
	}
	for _, c := range cases {
		args, legacy := Dispatch(strings.Fields(c))
		root := NewRoot()
		args = RewriteArgs(root, args, legacy)
		cmd, rest, err := root.Find(args)
		if err != nil {
			t.Errorf("%s: %v", c, err)
			continue
		}
		cmd.InitDefaultHelpFlag()
		cmd.InitDefaultVersionFlag()
		if err := cmd.ParseFlags(rest); err != nil {
			t.Errorf("%s: %v (rewritten: %q)", c, err, args)
		}
	}
}

func TestRedactViaLegacyName(t *testing.T) {
	args, legacy := Dispatch([]string{"span-redact", "-b", "1", "-w", "1"})
	root := NewRoot()
	root.SetArgs(RewriteArgs(root, args, legacy))
	var out bytes.Buffer
	root.SetIn(strings.NewReader(`{"finc.id":"1","x.fulltext":"secret"}` + "\n"))
	root.SetOut(&out)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"finc.id":"1"`) || strings.Contains(out.String(), "secret") {
		t.Errorf("unexpected output: %s", out.String())
	}
}
