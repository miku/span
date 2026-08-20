package reshape

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

	"github.com/segmentio/encoding/json"
)

// Golden tests for the input formats. The whole contract of a format package
// is a pure function -- vendor bytes in, intermediate schema out -- so it can
// be pinned with a sample input and its expected output:
//
//	testdata/<format>/input.<ext>     a small, real sample record or two
//	testdata/<format>/golden.ndjson   the intermediate schema it converts to
//
// One test walks the format registry, so adding a format under test means
// adding two files and (if listed) removing a line from formatsWithoutSample.
// Regenerate golden files after an intended change with:
//
//	go test ./internal/cmd/reshape -run TestGolden -update
//
// and review the diff -- a schema change should show up there rather than in
// the production index.

var update = flag.Bool("update", false, "update golden files in testdata/")

const testdataDir = "testdata"

// formatsWithoutSample lists registry formats that have no sample input yet,
// with a note on what is needed. These are skipped instead of failing, so the
// suite stays green while the list shrinks. A format not listed here must have
// testdata, which keeps new formats from being added without one.
var formatsWithoutSample = map[string]string{
	"ceeol-marcxml": "need a CEEOL MARCXML sample",
	"dblp":          "need a dblp.xml <article> sample",
	"doaj":          "need a DOAJ API v1 article sample (ndjson)",
	"doaj-oai":      "need a DOAJ OAI oai_dc sample",
	"elsevier-tar":  "need a small Elsevier shipment tar",
	"genderopen":    "need a GenderOpen OAI sample",
	"highwire":      "need a Highwire OAI sample",
	"ieee":          "need an IEEE publication XML sample",
	"imslp":         "need an IMSLP data sample (single text record)",
	"ios":           "need an IOS Press article XML sample",
	"jstor":         "need a JSTOR article XML sample",
	"mediarep-dim":  "need a mediarep DIM sample",
	"olms":          "need an OLMS oai_dc sample",
	"olms-mets":     "need an OLMS METS sample",
	"ssoar":         "need an SSOAR OAI sample",
	"thieme-nlm":    "need a Thieme NLM sample",
	"zvdd":          "need a ZVDD Dublin Core sample",
}

// TestGolden converts each format's sample input and compares the result
// against its golden file.
func TestGolden(t *testing.T) {
	for _, name := range FormatNames() {
		t.Run(name, func(t *testing.T) {
			input, err := sampleInput(name)
			if err != nil {
				if note, known := formatsWithoutSample[name]; known {
					t.Skipf("no sample input: %s", note)
				}
				t.Fatalf("%v; add %s/%s/input.<ext> or list the format in formatsWithoutSample",
					err, testdataDir, name)
			}
			got := convert(t, name, input)
			goldenPath := filepath.Join(testdataDir, name, "golden.ndjson")
			if *update {
				if err := os.WriteFile(goldenPath, got, 0644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				t.Logf("wrote %s (%d records)", goldenPath, bytes.Count(got, []byte("\n")))
				return
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("%v; run: go test ./internal/cmd/reshape -run TestGolden -update", err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("%s: output differs from golden\n%s", input, describeDiff(want, got))
			}
		})
	}
}

// notASource lists formats that are not backed by a real data source and so
// carry no identity fields. Only the documented minimal example belongs here.
var notASource = map[string]bool{"dummy": true}

// TestGoldenInvariants checks properties that hold for every intermediate
// schema record, whatever the source format. Golden files pin what a format
// does emit; this pins what every format must emit.
func TestGoldenInvariants(t *testing.T) {
	for _, name := range FormatNames() {
		t.Run(name, func(t *testing.T) {
			input, err := sampleInput(name)
			if err != nil {
				t.Skip("no sample input")
			}
			if notASource[name] {
				t.Skip("example format, carries no source identity")
			}
			records := decodeRecords(t, convert(t, name, input))
			if len(records) == 0 {
				t.Fatal("no records converted")
			}
			for i, rec := range records {
				for _, field := range []string{"finc.id", "finc.source_id", "version"} {
					if s, _ := rec[field].(string); s == "" {
						t.Errorf("record %d: %s is empty", i, field)
					}
				}
				if s, _ := rec["version"].(string); s != "" && s != "0.9" && s != "1.0" {
					t.Errorf("record %d: unexpected schema version %q", i, s)
				}
			}
		})
	}
}

// TestGoldenDeterministic checks that a format converts to the same bytes
// twice in a row, single threaded. Byte stability is a prerequisite for the
// planned incremental ("has this record changed?") update path, and a format
// that reaches for a map iteration or time.Now() would break it here first.
func TestGoldenDeterministic(t *testing.T) {
	for _, name := range FormatNames() {
		t.Run(name, func(t *testing.T) {
			input, err := sampleInput(name)
			if err != nil {
				t.Skip("no sample input")
			}
			first := convert(t, name, input)
			for i := range 3 {
				if got := convert(t, name, input); !bytes.Equal(got, first) {
					t.Fatalf("run %d differs from run 0\n%s", i+1, describeDiff(first, got))
				}
			}
		})
	}
}

// TestFormatsWithoutSampleIsAccurate keeps the skip list from going stale: it
// may only name real formats, and none of them may already have testdata.
func TestFormatsWithoutSampleIsAccurate(t *testing.T) {
	names := FormatNames()
	for name, note := range formatsWithoutSample {
		if !slices.Contains(names, name) {
			t.Errorf("formatsWithoutSample names %q, which is not a registered format", name)
		}
		if note == "" {
			t.Errorf("formatsWithoutSample[%q] has no note on what is needed", name)
		}
		if path, err := sampleInput(name); err == nil {
			t.Errorf("formatsWithoutSample names %q, but %s exists; remove the entry", name, path)
		}
	}
}

// sampleInput returns the path to the single sample input for a format.
func sampleInput(name string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(testdataDir, name, "input.*"))
	if err != nil {
		return "", err
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no sample input for format %q", name)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("format %q has %d sample inputs, want exactly one: %v", name, len(matches), matches)
	}
}

// convert runs a format over its sample input the way span-import would, but
// single threaded, so the output order is the input order.
func convert(t *testing.T, name, input string) []byte {
	t.Helper()
	f, err := os.Open(input)
	if err != nil {
		t.Fatalf("open input: %v", err)
	}
	defer f.Close()
	var buf bytes.Buffer
	cfg := Config{Name: name, BatchSize: 1, NumWorkers: 1}
	if err := Run(cfg, f, &buf); err != nil {
		t.Fatalf("Run(%s): %v", name, err)
	}
	return buf.Bytes()
}

// decodeRecords parses ndjson output into generic maps.
func decodeRecords(t *testing.T, b []byte) []map[string]any {
	t.Helper()
	var records []map[string]any
	for i, line := range splitLines(b) {
		var rec map[string]any
		if err := json.Unmarshal(line, &rec); err != nil {
			t.Fatalf("line %d is not JSON: %v (%.120q)", i+1, err, line)
		}
		records = append(records, rec)
	}
	return records
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

// describeDiff renders a field level report of how two ndjson streams differ.
// Golden files are one long line per record, which git diffs badly, so the
// test does the work of saying which field changed rather than which byte.
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
	b.WriteString("  regenerate with: go test ./internal/cmd/reshape -run TestGolden -update")
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
