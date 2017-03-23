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

func TestContainsVolume(t *testing.T) {
	var cases = []struct {
		entry  Entry
		volume string
		err    error
	}{
		{Entry{}, "", nil},
		{Entry{}, "10", nil},
		{Entry{FirstVolume: "1"}, "10", nil},
		{Entry{FirstVolume: "Vol. 1"}, "10", nil},
		{Entry{FirstVolume: "Vol. 10"}, "10", nil},
		{Entry{FirstVolume: "Vol. X"}, "10", nil},
		{Entry{LastVolume: "10"}, "10", nil},
		{Entry{LastVolume: "10th volume"}, "10", nil},
		{Entry{LastVolume: "11th volume"}, "10", nil},
		{Entry{LastVolume: "9"}, "10", ErrAfterLastVolume},
		{Entry{FirstVolume: "11", LastVolume: ""}, "10", ErrBeforeFirstVolume},
		{Entry{FirstVolume: "11", LastVolume: ""}, "", nil},
	}
	for _, c := range cases {
		err := c.entry.containsVolume(c.volume)
		if err != c.err {
			t.Errorf("containsVolume(%v, %v): got %v, want %v", c.entry, c.volume, err, c.err)
		}
	}
}

func TestContainsIssue(t *testing.T) {
	var cases = []struct {
		entry Entry
		issue string
		err   error
	}{
		{Entry{}, "", nil},
		{Entry{}, "10", nil},
		{Entry{FirstIssue: "1"}, "10", nil},
		{Entry{LastIssue: "9"}, "10", ErrAfterLastIssue},
		{Entry{FirstIssue: "11", LastIssue: ""}, "10", ErrBeforeFirstIssue},
		{Entry{FirstIssue: "11", LastIssue: ""}, "", nil},
	}
	for _, c := range cases {
		err := c.entry.containsIssue(c.issue)
		if err != c.err {
			t.Errorf("containsIssue(%v, %v): got %v, want %v", c.entry, c.issue, err, c.err)
		}
	}
}

func TestCovers(t *testing.T) {
	var cases = []struct {
		entry  Entry
		date   string
		volume string
		issue  string
		err    error
	}{
		{Entry{}, "", "", "", ErrInvalidDate},
		{Entry{}, "2000", "", "", nil},
		{Entry{FirstIssueDate: "1990-01-01", LastIssueDate: "2008-01-01"}, "1989", "", "", ErrBeforeFirstIssueDate},
		{Entry{FirstIssueDate: "1990-01-01", LastIssueDate: "2008-01-01"}, "2008-02", "", "", ErrAfterLastIssueDate},
		{Entry{FirstIssueDate: "2000", FirstVolume: "3", FirstIssue: "21", LastIssueDate: "2008"}, "2000", "3", "20", ErrBeforeFirstIssue},
		{Entry{FirstIssueDate: "2000", FirstVolume: "3", LastIssueDate: "2008"}, "2000", "1", "", ErrBeforeFirstVolume},
		{Entry{FirstIssueDate: "2000", LastIssueDate: "2008", LastVolume: "2", LastIssue: "12"}, "2008", "2", "13", ErrAfterLastIssue},
		{Entry{FirstIssueDate: "2000", LastIssueDate: "2008", LastVolume: "2"}, "2008", "3", "", ErrAfterLastVolume},
		{Entry{FirstIssueDate: "2001-05-05"}, "2001-05-04", "", "", ErrBeforeFirstIssueDate},
		{Entry{FirstIssueDate: "2001-05-05"}, "2001-05", "", "", nil},
		{Entry{FirstIssueDate: "2001-05"}, "2001-05-04", "", "", nil},
		{Entry{FirstIssueDate: "2001"}, "2000", "", "", ErrBeforeFirstIssueDate},
		{Entry{FirstIssueDate: "2001"}, "2001-05-05", "", "", nil},
		{Entry{FirstIssueDate: "2001"}, "2001", "", "", nil},
		{
			Entry{
				FirstIssueDate: time.Now().Add(-168 * time.Hour).Format("2006-01-02"),
				Embargo:        "P1Y",
			}, time.Now().Format("2006-01-02"), "", "", ErrAfterMovingWall,
		},
	}
	for _, c := range cases {
		err := c.entry.Covers(c.date, c.volume, c.issue)
		if err != c.err {
			t.Errorf("Covers(%v, %v, %v, %v): got %v, want %v", c.entry, c.date, c.volume, c.issue, err, c.err)
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
