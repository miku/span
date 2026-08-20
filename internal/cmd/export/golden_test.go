package export

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	json "github.com/segmentio/encoding/json"
)

// Golden tests for the export formats, the other end of the hub. The input is
// testdata/input.is, the concatenation of every reshape golden file, so a
// change to a source format shows up here too -- this is the shape of the
// documents that reach the index.
//
//	testdata/input.is                 intermediate schema records
//	testdata/golden.<format>.ndjson   what each exporter makes of them
//
// Regenerate after an intended change with:
//
//	go test ./internal/cmd/export -run TestGolden -update
//
// which rebuilds input.is from the reshape goldens first, so adding a source
// format sample there automatically widens the coverage here.

var update = flag.Bool("update", false, "update golden files in testdata/")

const (
	testdataDir = "testdata"
	inputFile   = "testdata/input.is"

	// reshapeGoldens is where the source side keeps its converted records.
	// Reaching across packages for testdata is unusual, but the alternative
	// is a hand-maintained copy that silently goes stale whenever a format
	// gains a sample.
	reshapeGoldens = "../reshape/testdata/*/golden.ndjson"
)

// excludedFromInput names reshape goldens that must not feed the export tests.
// dummy is the minimal example format and emits no finc.id, which every real
// record has.
var excludedFromInput = map[string]bool{"dummy": true}

// exportFormats are the format names to pin, including the solr5vu3v12 alias,
// which is solr5vu3 with the fullrecord field populated.
var exportFormats = []string{"formeta", "solr5vu3", "solr5vu3v12"}

func TestGolden(t *testing.T) {
	if *update {
		if err := os.WriteFile(inputFile, buildInput(t), 0644); err != nil {
			t.Fatalf("write %s: %v", inputFile, err)
		}
	}
	for _, name := range exportFormats {
		t.Run(name, func(t *testing.T) {
			got := export(t, name)
			goldenPath := filepath.Join(testdataDir, "golden."+name+".ndjson")
			if *update {
				if err := os.WriteFile(goldenPath, got, 0644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				t.Logf("wrote %s (%d records)", goldenPath, bytes.Count(got, []byte("\n")))
				return
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("%v; run: go test ./internal/cmd/export -run TestGolden -update", err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("output differs from golden\n%s", describeDiff(want, got))
			}
		})
	}
}

// TestInputIsCurrent fails if a source format gained or changed a sample
// without the export side being regenerated, so downstream coverage cannot
// quietly fall behind the reshape goldens.
func TestInputIsCurrent(t *testing.T) {
	want := buildInput(t)
	got, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("%v; run: go test ./internal/cmd/export -run TestGolden -update", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s is stale against %s (%d records on disk, %d in the reshape goldens); "+
			"run: go test ./internal/cmd/export -run TestGolden -update",
			inputFile, reshapeGoldens, len(splitLines(got)), len(splitLines(want)))
	}
}

// buildInput concatenates the reshape golden files, in glob order, which is
// sorted by format name.
func buildInput(t *testing.T) []byte {
	t.Helper()
	paths, err := filepath.Glob(reshapeGoldens)
	if err != nil {
		t.Fatalf("glob %s: %v", reshapeGoldens, err)
	}
	if len(paths) == 0 {
		t.Fatalf("no reshape goldens at %s", reshapeGoldens)
	}
	var buf bytes.Buffer
	for _, path := range paths {
		if excludedFromInput[filepath.Base(filepath.Dir(path))] {
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		buf.Write(b)
	}
	return buf.Bytes()
}

// TestGoldenCoversEveryExporter fails if an exporter is added to the registry
// without a golden file, so the pinned set cannot silently fall behind.
func TestGoldenCoversEveryExporter(t *testing.T) {
	for _, name := range FormatNames() {
		if !slices.Contains(exportFormats, name) {
			t.Errorf("exporter %q has no golden test; add it to exportFormats", name)
		}
	}
}

// TestGoldenSolrInvariants pins properties every SOLR document must have,
// whatever the source format it came from.
func TestGoldenSolrInvariants(t *testing.T) {
	for _, line := range splitLines(export(t, "solr5vu3")) {
		var doc map[string]any
		if err := json.Unmarshal(line, &doc); err != nil {
			t.Fatalf("not JSON: %v", err)
		}
		id, _ := doc["id"].(string)
		if id == "" {
			t.Errorf("document has no id: %.120q", line)
			continue
		}
		if s, _ := doc["record_format"].(string); s != "is" {
			t.Errorf("%s: record_format = %q, want is", id, s)
		}
		if s, _ := doc["source_id"].(string); s == "" {
			t.Errorf("%s: source_id is empty", id)
		}
		// Without -with-fullrecord the field carries a blob reference rather
		// than the record itself, refs #8031.
		if s, _ := doc["fullrecord"].(string); s != "blob:"+id {
			t.Errorf("%s: fullrecord = %.60q, want blob:%s", id, s, id)
		}
	}
}

// TestGoldenDeterministic checks that exporting twice yields the same bytes.
// The SOLR document derives facets from set valued fields, which is where
// unordered iteration tends to leak into output.
func TestGoldenDeterministic(t *testing.T) {
	for _, name := range exportFormats {
		t.Run(name, func(t *testing.T) {
			first := export(t, name)
			for i := range 3 {
				if got := export(t, name); !bytes.Equal(got, first) {
					t.Fatalf("run %d differs from run 0\n%s", i+1, describeDiff(first, got))
				}
			}
		})
	}
}

// export runs an exporter over the shared input, single threaded, so the
// output order is the input order.
func export(t *testing.T, name string) []byte {
	t.Helper()
	f, err := os.Open(inputFile)
	if err != nil {
		t.Fatalf("open input: %v", err)
	}
	defer f.Close()
	var buf bytes.Buffer
	cfg := Config{Format: name, BatchSize: 1, NumWorkers: 1}
	if err := Run(cfg, f, &buf); err != nil {
		t.Fatalf("Run(%s): %v", name, err)
	}
	return buf.Bytes()
}

func splitLines(b []byte) [][]byte {
	var lines [][]byte
	for line := range bytes.SplitSeq(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) > 0 {
			lines = append(lines, line)
		}
	}
	return lines
}

// describeDiff renders a field level report of how two output streams differ.
// Formeta is not JSON, so it falls back to a truncated literal comparison.
func describeDiff(want, got []byte) string {
	var b strings.Builder
	wantLines, gotLines := splitLines(want), splitLines(got)
	if len(wantLines) != len(gotLines) {
		fmt.Fprintf(&b, "  record count: want %d, got %d\n", len(wantLines), len(gotLines))
	}
	for i := range min(len(wantLines), len(gotLines)) {
		if bytes.Equal(wantLines[i], gotLines[i]) {
			continue
		}
		fmt.Fprintf(&b, "  record %d:\n", i+1)
		var wantRec, gotRec map[string]any
		if json.Unmarshal(wantLines[i], &wantRec) != nil || json.Unmarshal(gotLines[i], &gotRec) != nil {
			fmt.Fprintf(&b, "    want %.200q\n    got  %.200q\n", wantLines[i], gotLines[i])
			continue
		}
		for _, k := range sortedKeys(wantRec, gotRec) {
			w, wok := wantRec[k]
			g, gok := gotRec[k]
			switch {
			case !gok:
				fmt.Fprintf(&b, "    - %s: %s\n", k, render(w))
			case !wok:
				fmt.Fprintf(&b, "    + %s: %s\n", k, render(g))
			case render(w) != render(g):
				fmt.Fprintf(&b, "    ~ %s: %s -> %s\n", k, render(w), render(g))
			}
		}
	}
	b.WriteString("  regenerate with: go test ./internal/cmd/export -run TestGolden -update")
	return b.String()
}

func sortedKeys(maps ...map[string]any) []string {
	seen := make(map[string]bool)
	var keys []string
	for _, m := range maps {
		for k := range m {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)
	return keys
}

func render(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	if len(b) > 160 {
		return string(b[:157]) + "..."
	}
	return string(b)
}
