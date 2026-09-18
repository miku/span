package crossrefcmd

import (
	"testing"
	"time"
)

func TestIntervals(t *testing.T) {
	date := func(s string) time.Time {
		t.Helper()
		v, err := time.Parse(time.DateTime, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	var cases = []struct {
		about      string
		f          func(s, e time.Time) []Interval
		start, end string
		want       int
		first      Interval
	}{
		{"day same", Daily, "2000-01-01 00:00:00", "2000-01-01 00:00:00", 0, Interval{}},
		{"day reversed", Daily, "2000-01-01 00:00:00", "1999-01-01 00:00:00", 0, Interval{}},
		{"day partial", Daily, "2000-01-01 10:30:00", "2000-01-02 12:00:00", 2,
			Interval{date("2000-01-01 10:30:00"), date("2000-01-01 23:59:59.999999999")}},
		{"day full year", Daily, "2000-01-01 00:00:00", "2001-01-01 00:00:00", 366,
			Interval{date("2000-01-01 00:00:00"), date("2000-01-01 23:59:59.999999999")}},
		// 2000-01-01 is a Saturday, weeks start on Sunday.
		{"week full year", Weekly, "2000-01-01 00:00:00", "2001-01-01 00:00:00", 54,
			Interval{date("2000-01-01 00:00:00"), date("2000-01-01 23:59:59.999999999")}},
		{"week from sunday", Weekly, "2000-01-02 00:00:00", "2000-01-09 00:00:00", 1,
			Interval{date("2000-01-02 00:00:00"), date("2000-01-08 23:59:59.999999999")}},
		{"month same day", Monthly, "2000-01-01 10:30:00", "2000-01-01 12:00:00", 1,
			Interval{date("2000-01-01 10:30:00"), date("2000-01-31 23:59:59.999999999")}},
		{"month leap february", Monthly, "2000-02-10 00:00:00", "2000-03-02 00:00:00", 2,
			Interval{date("2000-02-10 00:00:00"), date("2000-02-29 23:59:59.999999999")}},
	}
	for _, c := range cases {
		t.Run(c.about, func(t *testing.T) {
			ivs := c.f(date(c.start), date(c.end))
			if len(ivs) != c.want {
				t.Fatalf("got %d intervals, want %d: %v", len(ivs), c.want, ivs)
			}
			if len(ivs) > 0 && ivs[0] != c.first {
				t.Errorf("first interval: got %v, want %v", ivs[0], c.first)
			}
			for i := 1; i < len(ivs); i++ {
				if gap := ivs[i].Start.Sub(ivs[i-1].End); gap != time.Nanosecond {
					t.Errorf("gap between %v and %v: %v", ivs[i-1], ivs[i], gap)
				}
			}
		})
	}
}
