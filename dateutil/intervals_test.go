package dateutil

import (
	"fmt"
	"testing"
	"time"
)

func TestMakeIntervalFunc(t *testing.T) {
	var cases = []struct {
		about        string
		padLeft      PadFunc
		padRight     PadFunc
		start        time.Time
		end          time.Time
		numIntervals int
	}{
		{"minute 90min span", padLMinute, padRMinute, MustParse("2000-01-01 10:30"), MustParse("2000-01-01 12:00"), 90},
		{"minute 91min span", padLMinute, padRMinute, MustParse("2000-01-01 10:30:30"), MustParse("2000-01-01 12:00:30"), 91},
		{"hour exact", padLHour, padRHour, MustParse("2000-01-01 10:00"), MustParse("2000-01-01 12:00"), 2},
		{"hour partial", padLHour, padRHour, MustParse("2000-01-01 10:30"), MustParse("2000-01-01 12:00"), 2},
		{"day same", padLDay, padRDay, MustParse("2000-01-01"), MustParse("2000-01-01"), 0},
		{"day reversed", padLDay, padRDay, MustParse("2000-01-01"), MustParse("1999-01-01"), 0},
		{"day full year", padLDay, padRDay, MustParse("2000-01-01"), MustParse("2001-01-01"), 366},
		{"week full year", padLWeek, padRWeek, MustParse("2000-01-01"), MustParse("2001-01-01"), 54},
		{"month same day", padLMonth, padRMonth, MustParse("2000-01-01 10:30"), MustParse("2000-01-01 12:00"), 1},
		{"biweek 20 days", padLBiweek, padRBiweek, MustParse("2000-01-01 00:00"), MustParse("2000-01-20 00:00"), 2},
	}
	for _, c := range cases {
		t.Run(c.about, func(t *testing.T) {
			f := makeIntervalFunc(c.padLeft, c.padRight)
			ivs := f(c.start, c.end)
			if len(ivs) != c.numIntervals {
				t.Errorf("got %d intervals, want %d", len(ivs), c.numIntervals)
			}
			if t.Failed() {
				switch len(ivs) {
				case 0:
					t.Logf("start: %v, end: %v, intervals: []", c.start, c.end)
				case 1:
					t.Logf("start: %v, end: %v, intervals: [%v]", c.start, c.end, ivs[0])
				default:
					t.Logf("start: %v, end: %v, intervals: [%v, ..., %v]", c.start, c.end, ivs[0], ivs[len(ivs)-1])
				}
			}
		})
	}
}

func BenchmarkMakeIntervalFunc(b *testing.B) {
	benchmarks := []struct {
		name     string
		padLeft  PadFunc
		padRight PadFunc
		start    time.Time
		end      time.Time
	}{
		{"day/year", padLDay, padRDay, MustParse("2000-01-01"), MustParse("2001-01-01")},
		{"minute/90min", padLMinute, padRMinute, MustParse("2000-01-01 10:30"), MustParse("2000-01-01 12:00")},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			f := makeIntervalFunc(bm.padLeft, bm.padRight)
			for i := 0; i < b.N; i++ {
				_ = f(bm.start, bm.end)
			}
		})
	}
}

func BenchmarkMustParse(b *testing.B) {
	inputs := []string{"2000-01-01", "2000-01-01 10:30", "2000-01-01 10:30:30"}
	for _, in := range inputs {
		b.Run(fmt.Sprintf("len=%d", len(in)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = MustParse(in)
			}
		})
	}
}
