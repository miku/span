package holdings

import (
	"fmt"
	"testing"
	"time"
)

func TestParseDelay(t *testing.T) {
	var tests = []struct {
		s   string
		d   time.Duration
		err error
	}{
		{"-0M", time.Duration(0), nil},
		{"-1M", time.Duration(-2592000000000000), nil},
		{"-2M", time.Duration(-5184000000000000), nil},
		{"-1Y", time.Duration(-31536000000000000), nil},
		{"-1D", time.Duration(0), fmt.Errorf("unknown format: -1D")},
		{"-1", time.Duration(0), fmt.Errorf("unknown format: -1")},
		{"129", time.Duration(0), fmt.Errorf("unknown format: 129")},
		{"AB", time.Duration(0), fmt.Errorf("unknown format: AB")},
		{"-111m", time.Duration(0), fmt.Errorf("unknown format: -111m")},
		{"0.1M", time.Duration(0), fmt.Errorf("unknown format: 0.1M")},
	}

	for _, tt := range tests {
		d, err := ParseDelay(tt.s)
		if d != tt.d {
			t.Errorf("ParseDelay(%s) => %v, %v, want %v, %v", tt.s, d, err, tt.d, tt.err)
		}
		if err != nil {
			if tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("ParseDelay(%s) => %v, %v, want %v, %v", tt.s, d, err, tt.d, tt.err)
				}
			} else {
				t.Errorf("ParseDelay(%s) => %v, %v, want %v, %v", tt.s, d, err, tt.d, tt.err)
			}
		}
	}
}

func TestDelay(t *testing.T) {
	var tests = []struct {
		e   Entitlement
		d   time.Duration
		err error
	}{
		{Entitlement{FromDelay: "-1M"}, time.Duration(-2592000000000000), nil},
		{Entitlement{ToDelay: "-1M"}, time.Duration(-2592000000000000), nil},
		{Entitlement{FromDelay: "-1M", ToDelay: "-1M"}, time.Duration(-2592000000000000), nil},
		{Entitlement{FromDelay: "-1M", ToDelay: "-2M"}, time.Duration(-2592000000000000), nil},
		{Entitlement{FromDelay: "-2M", ToDelay: "-1M"}, time.Duration(-5184000000000000), nil},
	}
	for _, tt := range tests {
		d, err := tt.e.Delay()
		if d != tt.d {
			t.Errorf("%v.Delay() => %v, %v, want %v, %v", tt.e, d, err, tt.d, tt.err)
		}
		if err != nil {
			if tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("%v.Delay() => %v, %v, want %v, %v", tt.e, d, err, tt.d, tt.err)
				}
			} else {
				t.Errorf("%v.Delay() => %v, %v, want %v, %v", tt.e, d, err, tt.d, tt.err)
			}
		}
	}
}
