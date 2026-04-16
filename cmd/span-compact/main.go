// span-compact deduplicates NDJSON streams on a user-defined field, keeping
// one record per key chosen by a selectable strategy. Built for the
// 10M-100M record range: uses an external sort instead of in-memory hash
// maps, so memory usage stays bounded. Works with stdin or a file (auto
// decompresses .gz and .zst) and writes NDJSON to stdout or -o.
//
// Strategies:
//
//	first    keep the first record seen for a key
//	last     keep the last record seen
//	random   keep a uniformly random record (reservoir over the group)
//	min      keep the record with the smallest -sort-key value
//	max      keep the record with the largest -sort-key value
//
// Examples:
//
//	span-compact -key id -strategy last input.ndj
//	zstdcat big.ndj.zst | span-compact -key doi -strategy max \
//	    -sort-key indexed_at -numeric > out.ndj
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/miku/span"
	"github.com/segmentio/encoding/json"
)

var (
	keyField    = flag.String("key", "id", "JSON field to deduplicate on")
	sortField   = flag.String("sort-key", "", "JSON field used by -strategy min|max")
	strategy    = flag.String("strategy", "last", "first|last|random|min|max")
	numericSort = flag.Bool("numeric", false, "treat -sort-key as numeric")
	outputFile  = flag.String("o", "", "output file (default: stdout)")
	sortMem     = flag.String("S", "50%", "sort -S memory buffer")
	sortTmp     = flag.String("T", "", "sort -T temp directory")
	showVersion = flag.Bool("v", false, "print version")
)

func main() {
	log.SetFlags(0)
	flag.Parse()
	if *showVersion {
		fmt.Println(span.AppVersion)
		return
	}
	switch *strategy {
	case "first", "last", "random":
	case "min", "max":
		if *sortField == "" {
			log.Fatal("strategy min|max requires -sort-key")
		}
	default:
		log.Fatalf("unknown strategy: %q (want first|last|random|min|max)", *strategy)
	}

	in, closeIn, err := openInput(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer closeIn()

	out, closeOut, err := openOutput(*outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer closeOut()

	if err := run(in, out); err != nil {
		log.Fatal(err)
	}
}

// run wires the input stream through an external sort and emits the
// first record of each sorted group, which (given the per-strategy sort
// key) is the winning record.
func run(in io.Reader, out io.Writer) error {
	args := []string{"-t", "\t", "-k1,1"}
	switch *strategy {
	case "first":
		args = append(args, "-k2,2n")
	case "last":
		args = append(args, "-k2,2nr")
	case "random":
		args = append(args, "-k2,2n")
	case "min":
		if *numericSort {
			args = append(args, "-k2,2n")
		} else {
			args = append(args, "-k2,2")
		}
	case "max":
		if *numericSort {
			args = append(args, "-k2,2nr")
		} else {
			args = append(args, "-k2,2r")
		}
	}
	args = append(args, "-S", *sortMem)
	if *sortTmp != "" {
		args = append(args, "-T", *sortTmp)
	}
	cmd := exec.Command("sort", args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	cmd.Stderr = os.Stderr
	sortIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	sortOut, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	feedErr := make(chan error, 1)
	go func() { feedErr <- feed(in, sortIn) }()

	if err := emit(sortOut, out); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		<-feedErr
		return err
	}
	if err := <-feedErr; err != nil {
		_ = cmd.Wait()
		return err
	}
	return cmd.Wait()
}

// feed reads NDJSON from r and writes `key\tcol2\tline` rows to w. col2
// encodes the per-strategy ordering so a single ascending/descending sort
// on (col1, col2) places the winning record first in each group.
func feed(r io.Reader, w io.WriteCloser) error {
	defer w.Close()
	bw := bufio.NewWriterSize(w, 1<<20)
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<27)
	var (
		buf  bytes.Buffer
		line int64
	)
	for sc.Scan() {
		b := sc.Bytes()
		if len(b) == 0 {
			continue
		}
		key, sortVal, err := extract(b)
		if err != nil {
			return fmt.Errorf("line %d: %w", line+1, err)
		}
		var col2 string
		switch *strategy {
		case "first", "last":
			col2 = fmt.Sprintf("%d", line)
		case "random":
			col2 = fmt.Sprintf("%d", rand.Int64())
		case "min", "max":
			col2 = sortVal
		}
		buf.Reset()
		buf.Write(key)
		buf.WriteByte('\t')
		buf.WriteString(col2)
		buf.WriteByte('\t')
		buf.Write(b)
		buf.WriteByte('\n')
		if _, err := bw.Write(buf.Bytes()); err != nil {
			return err
		}
		line++
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return bw.Flush()
}

// extract pulls the dedup key (raw JSON bytes) and, if requested, the
// sort key (as a string) from a single NDJSON record. Using raw JSON for
// the dedup key avoids the cost of unquoting strings and keeps numeric
// vs string keys unambiguous.
func extract(line []byte) (key []byte, sortVal string, err error) {
	var m map[string]json.RawMessage
	if err = json.Unmarshal(line, &m); err != nil {
		return nil, "", err
	}
	raw, ok := m[*keyField]
	if !ok {
		return nil, "", fmt.Errorf("missing key field %q", *keyField)
	}
	key = bytes.TrimSpace(raw)
	if *sortField != "" {
		sraw, ok := m[*sortField]
		if !ok {
			return nil, "", fmt.Errorf("missing sort field %q", *sortField)
		}
		sortVal = rawToSortString(bytes.TrimSpace(sraw))
	}
	return key, sortVal, nil
}

// rawToSortString returns a string suitable for sort(1) comparison. For
// JSON strings we strip the surrounding quotes so the lexicographic
// comparison matches the underlying value; for everything else we use
// the raw bytes.
func rawToSortString(raw []byte) string {
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return s
		}
	}
	return string(raw)
}

// emit reads sorted `key\tcol2\tline` rows from r and writes the
// original JSON line of the first occurrence of each key group to w.
func emit(r io.Reader, w io.Writer) error {
	bw := bufio.NewWriterSize(w, 1<<20)
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<27)
	var (
		cur   []byte
		first = true
	)
	for sc.Scan() {
		row := sc.Bytes()
		i := bytes.IndexByte(row, '\t')
		if i < 0 {
			continue
		}
		key := row[:i]
		rest := row[i+1:]
		j := bytes.IndexByte(rest, '\t')
		if j < 0 {
			continue
		}
		doc := rest[j+1:]
		if first || !bytes.Equal(key, cur) {
			if _, err := bw.Write(doc); err != nil {
				return err
			}
			if err := bw.WriteByte('\n'); err != nil {
				return err
			}
			cur = append(cur[:0], key...)
			first = false
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return bw.Flush()
}

// openInput opens path for reading, auto-detecting compression by
// extension, or returns stdin if path is empty.
func openInput(path string) (io.Reader, func(), error) {
	if path == "" {
		return os.Stdin, func() {}, nil
	}
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

func openOutput(path string) (io.Writer, func(), error) {
	if path == "" {
		return os.Stdout, func() {}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { f.Close() }, nil
}
