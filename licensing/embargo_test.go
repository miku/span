package licensing

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func mustParseDuration(s string) time.Duration {
	dur, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return dur
}

func mustParseTime(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

func TestEmbargoDuration(t *testing.T) {
	var cases = []struct {
		embargo Embargo
		dur     time.Duration
		err     error
	}{
		{embargo: Embargo("R1Y"), dur: mustParseDuration("8760h"), err: nil},
		{embargo: Embargo("R1M"), dur: mustParseDuration("730h"), err: nil},
		{embargo: Embargo("RaY"), dur: 0, err: ErrInvalidEmbargo},
		{embargo: Embargo("RRR"), dur: 0, err: ErrInvalidEmbargo},
	}
	for _, c := range cases {
		t.Run(string(c.embargo), func(t *testing.T) {
			dur, err := c.embargo.Duration()
			if err != c.err {
				t.Errorf("Duration: got %v, want %v", err, c.err)
			}
			if !reflect.DeepEqual(dur, c.dur) {
				t.Errorf("Duration: got %v, want %v", dur, c.dur)
			}
		})
	}
}

func TestEmbargoCompatible(t *testing.T) {
	var cases = []struct {
		about   string
		embargo Embargo
		t       time.Time
		rel     time.Time
		err     error
	}{
		{"P1D same day", Embargo("P1D"),
			mustParseTime("2006-01-02", "2000-01-01"),
			mustParseTime("2006-01-02", "2000-01-01"),
			ErrAfterMovingWall},
		{"P1D next day", Embargo("P1D"),
			mustParseTime("2006-01-02", "2000-01-01"),
			mustParseTime("2006-01-02", "2000-01-02"),
			nil},
		{"P1D two days later", Embargo("P1D"),
			mustParseTime("2006-01-02", "2000-01-01"),
			mustParseTime("2006-01-02", "2000-01-03"),
			nil},
		{"P1D one year later", Embargo("P1D"),
			mustParseTime("2006-01-02", "2000-01-01"),
			mustParseTime("2006-01-02", "2001-01-01"),
			nil},
		{"P12M one year later", Embargo("P12M"),
			mustParseTime("2006-01-02", "2000-01-01"),
			mustParseTime("2006-01-02", "2001-01-01"),
			nil},
		{"P1Y one year later", Embargo("P1Y"),
			mustParseTime("2006-01-02", "2000-01-01"),
			mustParseTime("2006-01-02", "2001-01-01"),
			nil},
		{"R1Y access begin", Embargo("R1Y"),
			mustParseTime("2006-01-02", "2000-01-03"),
			mustParseTime("2006-01-02", "2001-01-01"),
			nil},
	}
	for _, c := range cases {
		t.Run(c.about, func(t *testing.T) {
			err := c.embargo.CompatibleTo(c.t, c.rel)
			if err != c.err {
				t.Errorf("CompatibleTo(%v, %v, %v): got %v, want %v", c.embargo, c.t, c.rel, err, c.err)
			}
		})
	}
}

func TestEmbargoAccessBeginsAtWall(t *testing.T) {
	var cases = []struct {
		e                  Embargo
		accessBeginsAtWall bool
	}{
		{Embargo("1"), false},
		{Embargo("R1"), true},
		{Embargo("R1D"), true},
		{Embargo("R10M"), true},
		{Embargo("P10M"), false},
		{Embargo("?10M"), false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s=%v", c.e, c.accessBeginsAtWall), func(t *testing.T) {
			got := c.e.AccessBeginsAtWall()
			if got != c.accessBeginsAtWall {
				t.Errorf("embargo.AccessBeginsAtWall() got %v, want %v", got, c.accessBeginsAtWall)
			}
		})
	}
}
