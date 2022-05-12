package dateutil

import (
	"testing"
	"time"
)

func TestMakeIntervalFunc(t *testing.T) {
	var cases = []struct {
		padLeft      PadFunc
		padRight     PadFunc
		start        time.Time
		end          time.Time
		numIntervals int
	}{
		{
			padLeft:      padLMinute,
			padRight:     padRMinute,
			start:        MustParse("2000-01-01 10:30"),
			end:          MustParse("2000-01-01 12:00"),
			numIntervals: 90,
		},
		{
			padLeft:      padLMinute,
			padRight:     padRMinute,
			start:        MustParse("2000-01-01 10:30:30"),
			end:          MustParse("2000-01-01 12:00:30"),
			numIntervals: 91,
		},
		{
			padLeft:      padLHour,
			padRight:     padRHour,
			start:        MustParse("2000-01-01 10:00"),
			end:          MustParse("2000-01-01 12:00"),
			numIntervals: 2,
		},
		{
			padLeft:      padLHour,
			padRight:     padRHour,
			start:        MustParse("2000-01-01 10:30"),
			end:          MustParse("2000-01-01 12:00"),
			numIntervals: 2,
		},
		{
			padLeft:      padLDay,
			padRight:     padRDay,
			start:        MustParse("2000-01-01"),
			end:          MustParse("2000-01-01"),
			numIntervals: 0,
		},
		{
			padLeft:      padLDay,
			padRight:     padRDay,
			start:        MustParse("2000-01-01"),
			end:          MustParse("1999-01-01"),
			numIntervals: 0,
		},
		{
			padLeft:      padLDay,
			padRight:     padRDay,
			start:        MustParse("2000-01-01"),
			end:          MustParse("2001-01-01"),
			numIntervals: 366,
		},
		{
			padLeft:      padLWeek,
			padRight:     padRWeek,
			start:        MustParse("2000-01-01"),
			end:          MustParse("2001-01-01"),
			numIntervals: 54,
		},
		{
			padLeft:      padLMonth,
			padRight:     padRMonth,
			start:        MustParse("2000-01-01 10:30"),
			end:          MustParse("2000-01-01 12:00"),
			numIntervals: 1,
		},
		{
			padLeft:      padLBiweek,
			padRight:     padRBiweek,
			start:        MustParse("2000-01-01 00:00"),
			end:          MustParse("2000-01-20 00:00"),
			numIntervals: 2,
		},
	}
	for i, c := range cases {
		var (
			f   = makeIntervalFunc(c.padLeft, c.padRight)
			ivs = f(c.start, c.end)
		)
		t.Logf("[%d] start: %v, end: %v", i, c.start, c.end)
		switch len(ivs) {
		case 0:
			t.Logf("[%d] []", i)
		case 1:
			t.Logf("[%d] [%v]", i, ivs[0])
		case 2:
			t.Logf("[%d] [%v, %v]", i, ivs[0], ivs[1])
		default:
			t.Logf("[%d] [%v, ..., %v]", i, ivs[0], ivs[len(ivs)-1])
		}
		if len(ivs) != c.numIntervals {
			t.Fatalf("[%d] got %d, want %d", i, len(ivs), c.numIntervals)
		}
	}
}
