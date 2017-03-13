package kbart

import (
	"bufio"
	"os"
	"testing"
	"time"
)

var fixture = "../../fixtures/kbart.txt"

func TestByISSN(t *testing.T) {
	started := time.Now()
	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		t.Skipf("skipping, since fixture not found: %s", fixture)
	}
	file, err := os.Open(fixture)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer file.Close()
	holdings := new(Holdings)
	if _, err := holdings.ReadFrom(bufio.NewReader(file)); err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("loading took: %s", time.Since(started))
	entries := holdings.ByISSN("2079-8245")
	t.Logf("%d found: %s", len(entries), entries[0].PublicationTitle)

}

func BenchmarkByISSN(b *testing.B) {
	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		b.Skipf("skipping, since fixture not found: %s", fixture)
	}
	file, err := os.Open(fixture)
	if err != nil {
		b.Fatalf(err.Error())
	}
	defer file.Close()
	holdings := new(Holdings)
	if _, err := holdings.ReadFrom(bufio.NewReader(file)); err != nil {
		b.Fatalf(err.Error())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		holdings.ByISSN("2079-8245")
	}
}
