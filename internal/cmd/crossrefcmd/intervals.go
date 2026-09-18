package crossrefcmd

import (
	"fmt"
	"time"
)

// Interval groups start and end, both inclusive.
type Interval struct {
	Start time.Time
	End   time.Time
}

// String renders an interval.
func (iv Interval) String() string {
	return fmt.Sprintf("%s %s", iv.Start.Format(time.RFC3339), iv.End.Format(time.RFC3339))
}

// Daily, Weekly and Monthly split the time between start and end into
// intervals. The first interval begins at start, every interval ends right
// before the next day, week (starting on Sunday) or month begins, and the last
// interval covers end.
func Daily(start, end time.Time) []Interval   { return intervals(start, end, nextDay) }
func Weekly(start, end time.Time) []Interval  { return intervals(start, end, nextWeek) }
func Monthly(start, end time.Time) []Interval { return intervals(start, end, nextMonth) }

func nextDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d+1, 0, 0, 0, 0, t.Location())
}

func nextWeek(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d+7-int(t.Weekday()), 0, 0, 0, 0, t.Location())
}

func nextMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m+1, 1, 0, 0, 0, 0, t.Location())
}

// intervals chops up the time between start and end, using next to find the
// beginning of the following interval.
func intervals(start, end time.Time, next func(time.Time) time.Time) (result []Interval) {
	if !end.After(start) {
		return nil
	}
	last := end.Add(-time.Second)
	for l := start; ; {
		n := next(l)
		result = append(result, Interval{l, n.Add(-time.Nanosecond)})
		if l = n; l.After(last) {
			return result
		}
	}
}
