package licensing

import "testing"
import "time"
import "reflect"

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
		{
			embargo: Embargo("R1Y"), dur: mustParseDuration("8760h"), err: nil,
		},
		{
			embargo: Embargo("RaY"), dur: 0, err: ErrInvalidEmbargo,
		},
	}
	for _, c := range cases {
		dur, err := c.embargo.Duration()
		if err != c.err {
			t.Errorf("Duration: got %v, want %v", err, c.err)
		}
		if !reflect.DeepEqual(dur, c.dur) {
			t.Errorf("Duration: got %v, want %v", dur, c.dur)
		}
	}
}

func TestEmbargoCompatible(t *testing.T) {
	var cases = []struct {
		embargo Embargo
		t       time.Time
		rel     time.Time
		ok      bool
		err     error
	}{
		{
			embargo: Embargo("P1D"), // Access ends one day earlier.
			t:       mustParseTime("2006-01-02", "2000-01-01"),
			rel:     mustParseTime("2006-01-02", "2000-01-01"),
			err:     ErrAfterMovingWall,
		},
		{
			embargo: Embargo("P1D"), // Access ends one day earlier.
			t:       mustParseTime("2006-01-02", "2000-01-01"),
			rel:     mustParseTime("2006-01-02", "2000-01-02"),
			err:     nil,
		},
		{
			embargo: Embargo("P1D"), // Access ends one day earlier.
			t:       mustParseTime("2006-01-02", "2000-01-01"),
			rel:     mustParseTime("2006-01-02", "2000-01-03"),
			err:     nil,
		},
		{
			embargo: Embargo("P1D"), // Access ends one day earlier.
			t:       mustParseTime("2006-01-02", "2000-01-01"),
			rel:     mustParseTime("2006-01-02", "2001-01-01"),
			err:     nil,
		},
		{
			embargo: Embargo("P12M"), // Access ends 12 months from relative date.
			t:       mustParseTime("2006-01-02", "2000-01-01"),
			rel:     mustParseTime("2006-01-02", "2001-01-01"),
			err:     nil,
		},
		{
			embargo: Embargo("R1Y"),                            // Access begin one year from relative date.
			t:       mustParseTime("2006-01-02", "2000-01-03"), // Glitch due to fixed number of hours.
			rel:     mustParseTime("2006-01-02", "2001-01-01"),
			err:     nil,
		},
	}
	for _, c := range cases {
		err := c.embargo.CompatibleTo(c.t, c.rel)
		if err != c.err {
			t.Errorf("CompatibleTo(%v, %v, %v): got %v, want %v", c.embargo, c.t, c.rel, err, c.err)
		}
	}
}
