package parallel

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"testing"
)

var errFake1 = errors.New("fake error #1")

// LinesEqualSeparator returns true, if every line in a, when separated by
// separator, can be found in b.
func LinesEqualSeparator(a, b, sep string) bool {
	al := strings.Split(a, sep)
	bl := strings.Split(b, sep)
	if len(al) != len(bl) {
		return false
	}
	for _, line := range al {
		if !slices.Contains(bl, line) {
			return false
		}
	}
	return true
}

// LinesEqual returns true, if every line in a, when separated by a newline, can be found in b.
func LinesEqual(a, b string) bool {
	return LinesEqualSeparator(a, b, "\n")
}

func TestSimple(t *testing.T) {
	var cases = []struct {
		about    string
		r        io.Reader
		expected string
		f        TransformerFunc
		err      error
	}{
		{
			about:    "no input produces no output",
			r:        strings.NewReader(""),
			expected: "",
			f:        func(_ int64, b []byte) ([]byte, error) { return []byte{}, nil },
			err:      nil,
		},
		{
			about:    "order is not guaranteed",
			r:        strings.NewReader("a\nb\n"),
			expected: "B\nA\n",
			f:        func(_ int64, b []byte) ([]byte, error) { return bytes.ToUpper(b), nil },
			err:      nil,
		},
		{
			about:    "filter out items by returning nothing",
			r:        strings.NewReader("a\nb\n"),
			expected: "B\n",
			f: func(_ int64, b []byte) ([]byte, error) {
				if strings.TrimSpace(string(b)) == "a" {
					return []byte{}, nil
				}
				return bytes.ToUpper(b), nil
			},
			err: nil,
		},
		{
			about:    "empty lines skipped",
			r:        strings.NewReader("a\na\na\na\n\n\nb\n"),
			expected: "B\n",
			f: func(_ int64, b []byte) ([]byte, error) {
				if strings.TrimSpace(string(b)) == "a" {
					return []byte{}, nil
				}
				return bytes.ToUpper(b), nil
			},
			err: nil,
		},
		{
			about:    "on empty input transformer never called",
			r:        strings.NewReader(""),
			expected: "",
			f: func(_ int64, b []byte) ([]byte, error) {
				return nil, errFake1
			},
			err: nil,
		},
		{
			about:    "error does not come through if all lines skipped",
			r:        strings.NewReader("\n"),
			expected: "",
			f: func(_ int64, b []byte) ([]byte, error) {
				return nil, errFake1
			},
			err: nil,
		},
		{
			about:    "last line without trailing separator is processed",
			r:        strings.NewReader("a\nb"),
			expected: "A\nB",
			f:        func(_ int64, b []byte) ([]byte, error) { return bytes.ToUpper(b), nil },
			err:      nil,
		},
		{
			about:    "single line without trailing separator is processed",
			r:        strings.NewReader("a"),
			expected: "A",
			f:        func(_ int64, b []byte) ([]byte, error) { return bytes.ToUpper(b), nil },
			err:      nil,
		},
	}

	for _, c := range cases {
		t.Run(c.about, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewProcessor(c.r, &buf, c.f)
			err := p.Run()
			if err != c.err {
				t.Errorf("p.Run: got %v, want %v", err, c.err)
			}
			if !LinesEqual(buf.String(), c.expected) {
				t.Errorf("p.Run: got %v, want %v", buf.String(), c.expected)
			}
		})
	}
}

// TestConcurrentErrors exercises the error path with many workers and small
// batches; run with -race to verify error recording is synchronized.
func TestConcurrentErrors(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString("x\n")
	}
	p := NewProcessor(strings.NewReader(sb.String()), io.Discard, func(lineno int64, b []byte) ([]byte, error) {
		if lineno%3 == 0 {
			return nil, errFake1
		}
		return b, nil
	})
	p.BatchSize = 10
	if err := p.Run(); err != errFake1 {
		t.Errorf("p.Run: got %v, want %v", err, errFake1)
	}
}

// TestAllLinesProcessed checks that every input line shows up in the output,
// independent of batch boundaries and worker count.
func TestAllLinesProcessed(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 5000; i++ {
		sb.WriteString("x\n")
	}
	sb.WriteString("x") // no trailing newline
	var buf bytes.Buffer
	p := NewProcessor(strings.NewReader(sb.String()), &buf, func(_ int64, b []byte) ([]byte, error) {
		return []byte("y"), nil
	})
	p.BatchSize = 7
	if err := p.Run(); err != nil {
		t.Fatalf("p.Run: %v", err)
	}
	if got := buf.Len(); got != 5001 {
		t.Errorf("got %d lines, want 5001", got)
	}
}

// TestBatchedOutputKeepsEveryRecord checks that batching results on the way
// out does not lose, duplicate or corrupt records. Workers accumulate a whole
// batch before handing it over, so a boundary bug would show up here as a
// count mismatch rather than as garbled bytes.
func TestBatchedOutputKeepsEveryRecord(t *testing.T) {
	const numRecords = 5000
	var sb strings.Builder
	for i := 0; i < numRecords; i++ {
		fmt.Fprintf(&sb, "%d\n", i)
	}
	for _, numWorkers := range []int{1, 2, 8, 16} {
		for _, batchSize := range []int{1, 7, 1000, numRecords * 2} {
			t.Run(fmt.Sprintf("w=%d/b=%d", numWorkers, batchSize), func(t *testing.T) {
				var buf bytes.Buffer
				p := NewProcessor(strings.NewReader(sb.String()), &buf, func(_ int64, b []byte) ([]byte, error) {
					return b, nil
				})
				p.BatchSize = batchSize
				if err := p.RunWorkers(numWorkers); err != nil {
					t.Fatalf("p.Run: %v", err)
				}
				seen := make(map[string]int)
				for _, line := range strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n") {
					seen[line]++
				}
				if len(seen) != numRecords {
					t.Fatalf("got %d distinct records, want %d", len(seen), numRecords)
				}
				for i := 0; i < numRecords; i++ {
					if n := seen[strconv.Itoa(i)]; n != 1 {
						t.Fatalf("record %d appears %d times, want 1", i, n)
					}
				}
			})
		}
	}
}

// TestSingleWorkerPreservesOrder pins the one ordering guarantee callers can
// rely on today: with a single worker the output is in input order. The format
// golden tests convert this way, so they compare against stable bytes.
func TestSingleWorkerPreservesOrder(t *testing.T) {
	const numRecords = 1000
	var sb strings.Builder
	for i := 0; i < numRecords; i++ {
		fmt.Fprintf(&sb, "%d\n", i)
	}
	var buf bytes.Buffer
	p := NewProcessor(strings.NewReader(sb.String()), &buf, func(_ int64, b []byte) ([]byte, error) {
		return b, nil
	})
	p.BatchSize = 13
	if err := p.RunWorkers(1); err != nil {
		t.Fatalf("p.Run: %v", err)
	}
	if buf.String() != sb.String() {
		t.Error("single worker output is not in input order")
	}
}

// benchInput builds a line oriented input of roughly the given size.
func benchInput(numRecords int) string {
	var sb strings.Builder
	line := strings.Repeat("x", 250)
	for i := 0; i < numRecords; i++ {
		fmt.Fprintf(&sb, "%06d%s\n", i, line)
	}
	return sb.String()
}

// BenchmarkProcessor measures stage framing, not stage work: the transform is
// a copy, so what is left is the batching, channel handoffs and writes. This
// is the cost every stage pays per record on top of its own conversion.
func BenchmarkProcessor(b *testing.B) {
	input := benchInput(200000)
	for _, numWorkers := range []int{1, 2, 4, 8, 16} {
		b.Run(fmt.Sprintf("w=%d", numWorkers), func(b *testing.B) {
			b.SetBytes(int64(len(input)))
			b.ReportAllocs()
			for b.Loop() {
				p := NewProcessor(strings.NewReader(input), io.Discard, func(_ int64, x []byte) ([]byte, error) {
					return x, nil
				})
				p.BatchSize = 10000
				if err := p.RunWorkers(numWorkers); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
