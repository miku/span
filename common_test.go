package span

import (
	"fmt"
	"reflect"
	"testing"
)

func TestUnescapeTrim(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{in: "Hello", out: "Hello"},
		{in: "Hello &#x000F6;", out: "Hello ö"},
	}

	for _, tt := range tests {
		r := UnescapeTrim(tt.in)
		if r != tt.out {
			t.Errorf("UnescapeTrim(%s): got %s, want %s", tt.in, r, tt.out)
		}
	}
}

func TestParseHoldingSpec(t *testing.T) {
	var tests = []struct {
		spec   string
		result map[string]string
		err    error
	}{
		{spec: "", result: nil, err: fmt.Errorf("invalid spec: ")},
		{spec: "X:Z", result: map[string]string{"X": "Z"}, err: nil},
		{spec: "X:Z,A:b", result: map[string]string{"X": "Z", "A": "b"}, err: nil},
	}

	for _, tt := range tests {
		result, err := ParseHoldingSpec(tt.spec)
		if !reflect.DeepEqual(result, tt.result) {
			t.Errorf("ParseHoldingSpec(%s): got %v, want %v", tt.spec, result, tt.result)
		}
		if err != tt.err {
			if err != nil && tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("ParseHoldingSpec(%s): got %v, want %s", tt.spec, err, tt.err)
				}
			} else {
				t.Errorf("ParseHoldingSpec(%s): got %v, want %s", tt.spec, err, tt.err)
			}
		}
	}
}
