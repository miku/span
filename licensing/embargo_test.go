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

func TestEmbargoDuration(t *testing.T) {
	var cases = []struct {
		embargo Embargo
		dur     time.Duration
		err     error
	}{
		{
			embargo: Embargo("R1Y"), dur: mustParseDuration("-8760h"), err: nil,
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
