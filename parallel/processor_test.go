package parallel

import (
	"bytes"
	"errors"
	"io"
	"slices"
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
