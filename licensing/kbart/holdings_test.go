package kbart

import (
	"bufio"
	"compress/gzip"
	"os"
	"testing"
)

var (
	// $ gzip -dc testdata/kbart.txt.gz | sha1sum
	// d072bc9cef32ffbeaecc4c8c97562a1b9e47468c  -
	fixture  = "testdata/kbart.txt.gz"
	holdings *Holdings
)

// skipper is shared by tests and benchmarks.
type skipper interface {
	Skipf(format string, args ...any)
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
		zr, err := gzip.NewReader(file)
		if err != nil {
			s.Skipf("fixture: %v", err)
		}
		defer zr.Close()

		holdings = new(Holdings)
		if _, err := holdings.ReadFrom(bufio.NewReader(zr)); err != nil {
			s.Skipf("fixture: %v", err)
		}
	}
	return holdings
}

func TestSerialNumberMap(t *testing.T) {
	holdings := loadHoldings(t)
	m := holdings.SerialNumberMap()
	want := 37830
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

func TestWisoDatabaseMap(t *testing.T) {
	holdings := loadHoldings(t)
	m := holdings.WisoDatabaseMap()
	want := 511
	if len(m) != want {
		t.Errorf("WisoDatabaseMap: got %v, want %v", len(m), want)
	}
}
