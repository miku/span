package licensing

import (
	"reflect"
	"testing"
	"time"
)

func TestISSNList(t *testing.T) {
	var cases = []struct {
		entry  Entry
		result []string
	}{
		{
			Entry{PrintIdentifier: "1234-5678"},
			[]string{"1234-5678"},
		},
		{
			Entry{OnlineIdentifier: "1234-5678"},
			[]string{"1234-5678"},
		},
		{
			Entry{PrintIdentifier: "22", OnlineIdentifier: "1234-5678"},
			[]string{"1234-5678"},
		},
		{
			Entry{PrintIdentifier: "22", AllSerialNumbers: "1234-5678; Hello 2222-222X"},
			[]string{"1234-5678", "2222-222X"},
		},
		{
			Entry{
				OnlineIdentifier: "4444-4444",
				PrintIdentifier:  "3333-3333",
				AllSerialNumbers: "1234-5678; Hello 2222-222X",
			},
			[]string{"1234-5678", "2222-222X", "3333-3333", "4444-4444"},
		},
	}

	for _, c := range cases {
		result := c.entry.ISSNList()
		if !reflect.DeepEqual(result, c.result) {
			t.Errorf("ISSNList: got %v, want %v", result, c.result)
		}
	}
}

func TestContainsDate(t *testing.T) {
	var cases = []struct {
		entry Entry
		value string
		err   error
	}{
		{
			entry: Entry{
				FirstIssueDate: "1992-06-23",
				LastIssueDate:  "1998-04-10",
			},
			value: "1997",
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "",
				LastIssueDate:  "1998-04-10",
			},
			value: "1997",
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "",
				LastIssueDate:  "1998-04-10",
			},
			value: "1998-04",
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "",
				LastIssueDate:  "",
			},
			value: "1998-04",
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "2000",
				LastIssueDate:  "",
			},
			value: "1999-12-31",
			err:   ErrBeforeFirstIssueDate,
		},
		{
			entry: Entry{
				FirstIssueDate: "2000",
				LastIssueDate:  "",
			},
			value: "",
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "",
				LastIssueDate:  "",
			},
			value: "",
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "1992-02-10",
				LastIssueDate:  "1992-09-10",
			},
			value: "1992",
			err:   nil,
		},
	}

	for _, c := range cases {
		err := c.entry.containsDate(c.value)
		if err != c.err {
			t.Errorf("containsDate(%v, %v): got %v, want %v", c.entry, c.value, err, c.err)
		}
	}
}

func BenchmarkContainsDate(b *testing.B) {
	entry := Entry{
		FirstIssueDate: "1992-06-23",
		LastIssueDate:  "1998-04-10",
	}
	v := "1992"
	for i := 0; i < b.N; i++ {
		entry.containsDate(v)
	}
}

func TestCovers(t *testing.T) {
	var cases = []struct {
		about  string
		entry  Entry
		date   string
		volume string
		issue  string
		err    error
	}{
		{"an empty date is not accepted",
			Entry{}, "", "", "", ErrInvalidDate},
		{"an empty does not impose constraints",
			Entry{}, "2000", "", "", nil},
		{"given date lies before entry interval",
			Entry{FirstIssueDate: "1990-01-01", LastIssueDate: "2008-01-01"}, "1989", "", "", ErrBeforeFirstIssueDate},
		{"given date lies after entry interval",
			Entry{FirstIssueDate: "1990-01-01", LastIssueDate: "2008-01-01"}, "2008-02", "", "", ErrAfterLastIssueDate},
		{"date ok, but issue too early",
			Entry{FirstIssueDate: "2000", FirstVolume: "3", FirstIssue: "21", LastIssueDate: "2008"},
			"2000", "3", "20", ErrBeforeFirstIssue},
		{"date ok, but volume too early",
			Entry{FirstIssueDate: "2000", FirstVolume: "3", LastIssueDate: "2008"},
			"2000", "1", "", ErrBeforeFirstVolume},
		{"date ok, but issue too late",
			Entry{FirstIssueDate: "2000", LastIssueDate: "2008", LastVolume: "2", LastIssue: "12"},
			"2008", "2", "13", ErrAfterLastIssue},
		{"date ok, but volume too late",
			Entry{FirstIssueDate: "2000", LastIssueDate: "2008", LastVolume: "2"},
			"2008", "3", "", ErrAfterLastVolume},
		{"date too early, same granularity",
			Entry{FirstIssueDate: "2001-05-05"}, "2001-05-04", "", "", ErrBeforeFirstIssueDate},
		{"date ok, use the coarser granularity",
			Entry{FirstIssueDate: "2001-05-05"}, "2001-05", "", "", nil},
		{"date ok, use the coarser granularity",
			Entry{FirstIssueDate: "2001-05"}, "2001-05-04", "", "", nil},
		{"date too early",
			Entry{FirstIssueDate: "2001"}, "2000", "", "", ErrBeforeFirstIssueDate},
		{"date ok, use the coarser granularity",
			Entry{FirstIssueDate: "2001"}, "2001-05-05", "", "", nil},
		{"date ok",
			Entry{FirstIssueDate: "2001"}, "2001", "", "", nil},
		{
			"date ok, but moving wall hit",
			Entry{
				FirstIssueDate: time.Now().Add(-168 * time.Hour).Format("2006-01-02"),
				Embargo:        "P1Y",
			}, time.Now().Format("2006-01-02"), "", "", ErrAfterMovingWall,
		},
		{
			"date ok and moving wall hit",
			Entry{
				FirstIssueDate: "2000-01-01",
				Embargo:        "P1Y",
			}, time.Now().Add(-4000 * time.Hour).Format("2006-01-02"), "", "", ErrAfterMovingWall,
		},
		{
			"date ok and moving wall fine",
			Entry{
				FirstIssueDate: "2000-01-01",
				Embargo:        "P1Y",
			}, time.Now().Add(-10000 * time.Hour).Format("2006-01-02"), "", "", nil,
		},
		{
			"date ok, first volume after record volume",
			Entry{FirstIssueDate: "2000", FirstVolume: "6"}, "2001-05-05", "4", "", nil,
		},
		{
			"date ok, last volume before record volume",
			Entry{LastIssueDate: "2002", LastVolume: "3"}, "2001-05-05", "4", "", nil,
		},
		{
			"date ok, first volume after record volume, day granularity",
			Entry{FirstIssueDate: "2001-04-01", FirstVolume: "6"}, "2001-05-05", "4", "", ErrBeforeFirstVolume,
		},
		{
			"date ok, last volume before record volume, day granularity",
			Entry{LastIssueDate: "2001-06-01", LastVolume: "3"}, "2001-05-05", "4", "", ErrAfterLastVolume,
		},
		{
			"extended date",
			Entry{FirstIssueDate: "1870-05"}, "1879-12-01T00:00:00Z", "", "", nil,
		},
	}
	for _, c := range cases {
		err := c.entry.Covers(c.date, c.volume, c.issue)
		if err != c.err {
			t.Errorf("Covers(%#v, %v, %v, %v): got %v, want %v (%s)", c.entry, c.date, c.volume, c.issue, err, c.err, c.about)
		}
	}
}

func BenchmarkCovers(b *testing.B) {
	benchmarks := []struct {
		name   string
		entry  Entry
		date   string
		volume string
		issue  string
	}{
		{
			name: "full",
			entry: Entry{
				FirstIssueDate: "2000",
				FirstVolume:    "4",
				FirstIssue:     "28",
				LastIssueDate:  "2008",
				LastVolume:     "8",
				LastIssue:      "42",
			},
			date: "2008", volume: "8", issue: "43",
		},
		{
			name: "embargo",
			entry: Entry{
				FirstIssueDate: "2000",
				FirstVolume:    "4",
				FirstIssue:     "28",
				LastIssueDate:  "2008",
				LastVolume:     "8",
				LastIssue:      "42",
				Embargo:        "P60D",
			},
			date: "2008", volume: "8", issue: "43",
		},
		{
			name: "partial",
			entry: Entry{
				FirstIssueDate: "2000",
				LastIssueDate:  "2008",
				LastVolume:     "1",
			},
			date: "2008", volume: "8", issue: "43",
		},
		{
			name: "empty", entry: Entry{}, date: "", volume: "", issue: "",
		},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bm.entry.Covers(bm.date, bm.volume, bm.issue)
			}
		})
	}
}

// === RUN   TestCovers
// --- PASS: TestCovers (0.00s)
// BenchmarkContainsDate-4   	10000000	       221 ns/op
// BenchmarkCovers/full-4    	 3000000	       466 ns/op
// BenchmarkCovers/embargo-4 	 1000000	      1742 ns/op
// BenchmarkCovers/partial-4 	 5000000	       362 ns/op
// PASS
// ok  	github.com/miku/span/licensing	8.267s
