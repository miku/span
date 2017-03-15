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

var fixture = "../../fixtures/kbart.txt"

func loadHoldings() (*Holdings, error) {
	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		return nil, err
	}
	file, err := os.Open(fixture)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	holdings := new(Holdings)
	if _, err := holdings.ReadFrom(bufio.NewReader(file)); err != nil {
		return nil, err
	}
	return holdings, nil
}

func TestByISSN(t *testing.T) {
	holdings, err := loadHoldings()
	if err != nil {
		t.Fatalf(err.Error())
	}
	entries := holdings.ByISSN("2079-8245")
	if len(entries) != 1 {
		t.Errorf("ByISSN: got %v, want 1", len(entries))
	}
	t.Logf("%d found: %s", len(entries), entries[0].PublicationTitle)

}

func BenchmarkByISSN(b *testing.B) {
	holdings, err := loadHoldings()
	if err != nil {
		b.Fatalf(err.Error())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		holdings.ByISSN("2079-8245")
	}
}

func TestFilter(t *testing.T) {
	holdings, err := loadHoldings()
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Test filter.
	entries := holdings.Filter(func(e licensing.Entry) bool {
		return strings.Contains(strings.ToLower(e.TitleURL), "wiso")
	})
	if len(entries) != 702 {
		t.Errorf("Filter: got %v, want %v", len(entries), 702)
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

	if len(names.Values()) != 534 {
		t.Errorf("Filter: got %v, want %v", len(names.Values()), 534)
	}
}
