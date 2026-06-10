package licensing

import "testing"

// benchEntry resembles a typical KBART row with date and embargo constraints.
var benchEntry = Entry{
	PublicationTitle: "Journal of Benchmarks",
	PrintIdentifier:  "2079-8245",
	FirstIssueDate:   "1998-01-01",
	LastIssueDate:    "2008-06-30",
	FirstVolume:      "1",
	LastVolume:       "25",
	Embargo:          "P1Y",
}

// BenchmarkCoversParsed measures Covers on entry copies that carry the
// eagerly parsed date cache; this is the situation in
// filter.HoldingsFilter.Apply since kbart.ReadFrom pre-parses dates.
func BenchmarkCoversParsed(b *testing.B) {
	entry := benchEntry
	entry.ParseDates()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := entry // copy, like a lookup map value
		_ = e.Covers("2004-07-01", "10", "2")
	}
}

// BenchmarkCoversUnparsed measures Covers on entry copies without the date
// cache, so every call re-parses both issue dates (the behavior before
// eager parsing was added).
func BenchmarkCoversUnparsed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		e := benchEntry
		_ = e.Covers("2004-07-01", "10", "2")
	}
}
