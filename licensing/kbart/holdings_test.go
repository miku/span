package kbart

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/miku/span/container"
	"github.com/miku/span/licensing"
)

var (
	// $ sha1sum fixtures/kbart.txt
	// d072bc9cef32ffbeaecc4c8c97562a1b9e47468c  fixtures/kbart.txt
	fixture  = "../../fixtures/kbart.txt"
	holdings *Holdings
)

// skipper is shared by tests and benchmarks.
type skipper interface {
	Skipf(format string, args ...interface{})
}

func loadHoldings(s skipper) *Holdings {
	if holdings == nil {
		if _, err := os.Stat(fixture); os.IsNotExist(err) {
			s.Skipf("fixture: %v", err)
		}
		file, err := os.Open(fixture)
		if err != nil {
			s.Skipf("fixture: %v", err)
		}
		defer file.Close()

		holdings = new(Holdings)
		if _, err := holdings.ReadFrom(bufio.NewReader(file)); err != nil {
			s.Skipf("fixture: %v", err)
		}
	}
	return holdings
}

func TestFilter(t *testing.T) {
	holdings := loadHoldings(t)

	// Test filter.
	entries := holdings.Filter(func(e licensing.Entry) bool {
		return strings.Contains(strings.ToLower(e.TitleURL), "wiso")
	})
	want := 702
	if len(entries) != want {
		t.Errorf("Filter: got %v, want %v", len(entries), want)
	}

	// Test database name extraction.
	p := regexp.MustCompile(`[A-Z]{3,4}`)
	names := container.NewStringSet()
	for _, e := range entries {
		matches := p.FindAllString(e.TitleURL, -1)
		for _, m := range matches {
			names.Add(m)
		}
	}
	want = 534
	if len(names.Values()) != want {
		t.Errorf("Filter: got %v, want %v", len(names.Values()), want)
	}
}

func TestSerialNumberMap(t *testing.T) {
	holdings := loadHoldings(t)
	m := holdings.SerialNumberMap()
	want := 84089
	if len(m) != want {
		t.Errorf("SerialNumberMap: got %v, want %v", len(m), want)
	}
}

func BenchmarkSerialNumberMap(b *testing.B) {
	holdings := loadHoldings(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		holdings.SerialNumberMap()
	}
}

func BenchmarkLookupViaSerialNumberMap(b *testing.B) {
	holdings := loadHoldings(b)
	m := holdings.SerialNumberMap()
	b.ResetTimer()

	issn := "2079-8245"

	for i := 0; i < b.N; i++ {
		v := m[issn]
		_ = len(v) // Dummyop.
	}
}

func BenchmarkLookupViaFilter(b *testing.B) {
	holdings := loadHoldings(b)
	b.ResetTimer()

	issn := "2079-8245"
	f := func(e licensing.Entry) bool {
		if e.PrintIdentifier == issn || e.OnlineIdentifier == issn {
			return true
		}
		return false
	}

	for i := 0; i < b.N; i++ {
		holdings.Filter(f)
	}
}

// === RUN   TestFilter
// --- PASS: TestFilter (4.29s)
// === RUN   TestSerialNumberMap
// --- PASS: TestSerialNumberMap (0.45s)
// BenchmarkSerialNumberMap-4                      2           514861084   ns/op
// BenchmarkLookupViaSerialNumberMap-4             100000000          21.5 ns/op
// BenchmarkLookupViaFilter-4                      100          13340319   ns/op
// PASS
// ok    github.com/miku/span/licensing/kbart    12.653s

func TestWisoDatabaseMap(t *testing.T) {
	holdings := loadHoldings(t)
	m := holdings.WisoDatabaseMap()
	want := 534
	if len(m) != want {
		t.Errorf("WisoDatabaseMap: got %v, want %v", len(m), want)
	}
}
