// Package dateutil provides interval handling and a custom flag.
package dateutil

import (
	"time"

	"github.com/araddon/dateparse"
	"github.com/jinzhu/now"
)

var (
	// EveryMinute will chop up a timespan into 60s intervals;
	// https://english.stackexchange.com/questions/3091/weekly-daily-hourly-minutely.
	EveryMinute = makeIntervalFunc(padLMinute, padRMinute)
	Hourly      = makeIntervalFunc(padLHour, padRHour)
	Daily       = makeIntervalFunc(padLDay, padRDay)
	Weekly      = makeIntervalFunc(padLWeek, padRWeek)
	Monthly     = makeIntervalFunc(padLMonth, padRMonth)

	padLMinute = func(t time.Time) time.Time { return now.With(t).BeginningOfMinute() }
	padRMinute = func(t time.Time) time.Time { return now.With(t).EndOfMinute() }
	padLHour   = func(t time.Time) time.Time { return now.With(t).BeginningOfHour() }
	padRHour   = func(t time.Time) time.Time { return now.With(t).EndOfHour() }
	padLDay    = func(t time.Time) time.Time { return now.With(t).BeginningOfDay() }
	padRDay    = func(t time.Time) time.Time { return now.With(t).EndOfDay() }
	padLWeek   = func(t time.Time) time.Time { return now.With(t).BeginningOfWeek() }
	padRWeek   = func(t time.Time) time.Time { return now.With(t).EndOfWeek() }
	padLMonth  = func(t time.Time) time.Time { return now.With(t).BeginningOfMonth() }
	padRMonth  = func(t time.Time) time.Time { return now.With(t).EndOfMonth() }
)

// Interval groups start and end.
type Interval struct {
	Start time.Time
	End   time.Time
}

type (
	// padFunc allows to move a given time back and forth.
	padFunc func(t time.Time) time.Time
	// intervalFunc takes a start and endtime and returns a number of
	// intervals. How intervals are generated is flexible.
	intervalFunc func(s, e time.Time) []Interval
)

// makeIntervalFunc is a helper to create daily, weekly and other intervals.
// Given two shiftFuncs (to mark the beginning of an interval and the end), we
// return a function, that will allow us to generate intervals.
// TODO: We only need right pad, no?
func makeIntervalFunc(padLeft, padRight padFunc) intervalFunc {
	return func(s, e time.Time) (result []Interval) {
		if e.Before(s) || e.Equal(s) {
			return
		}
		e = e.Add(-1 * time.Second)
		var (
			l time.Time = s
			r time.Time
		)
		for {
			r = padRight(l)
			result = append(result, Interval{l, r})
			l = padLeft(r.Add(1 * time.Second))
			if l.After(e) {
				break
			}
		}
		return result
	}
}

// MustParse will panic on an unparsable date string.
func MustParse(value string) time.Time {
	t, err := dateparse.ParseStrict(value)
	if err != nil {
		panic(err)
	}
	return t
}
