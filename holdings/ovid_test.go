package holdings

import (
	"io"
	"reflect"
	"strings"
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
		{"-1M", time.Duration(-1) * month, nil},
		{"-2M", time.Duration(-2) * month, nil},
		{"-1Y", time.Duration(-1) * year, nil},
		{"-1D", time.Duration(0), errUnknownFormat},
		{"-1", time.Duration(0), errUnknownFormat},
		{"129", time.Duration(0), errUnknownFormat},
		{"AB", time.Duration(0), errUnknownFormat},
		{"-111m", time.Duration(0), errUnknownFormat},
		{"0.1M", time.Duration(0), errUnknownFormat},
	}

	for _, tt := range tests {
		d, err := parseDelay(tt.s)
		if d != tt.d {
			t.Errorf("parseDelay(%s) => %v, %v, want %v, %v", tt.s, d, err, tt.d, tt.err)
		}
		if err != nil {
			if tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("parseDelay(%s) => %v, %v, want %v, %v", tt.s, d, err, tt.d, tt.err)
				}
			} else {
				t.Errorf("parseDelay(%s) => %v, %v, want %v, %v", tt.s, d, err, tt.d, tt.err)
			}
		}
	}
}

func TestParseHoldings(t *testing.T) {
	var tests = []struct {
		r        io.Reader
		licenses Licenses
	}{
		{strings.NewReader(`
<holding ezb_id = "1">
  <title><![CDATA[Journal of Molecular Modeling]]></title>
  <publishers><![CDATA[Springer]]></publishers>
  <EZBIssns>
    <p-issn>1610-2940</p-issn>
    <e-issn>0948-5023</e-issn>
  </EZBIssns>
  <entitlements>
    <entitlement status = "subscribed">
      <url>http%3A%2F%2Flink.springer.com%2Fjournal%2F894</url>
      <anchor>natli_springer</anchor>
      <begin>
        <year>1995</year>
        <volume>1</volume>
      </begin>
      <end>
        <year>2002</year>
        <volume>8</volume>
      </end>
      <available><![CDATA[Nationallizenz]]></available>
    </entitlement>
    <entitlement status = "subscribed">
      <url>http%3A%2F%2Flink.springer.com%2Fjournal%2F894</url>
      <anchor>springer</anchor>
      <available><![CDATA[Konsortiallizenz - Gesamter Zeitraum]]></available>
    </entitlement>
  </entitlements>
</holding>`), Licenses{"1610-2940": []License{
			License("1995000001000000:2002000008000000:0"),
			License("0000000000000000:ZZZZZZZZZZZZZZZZ:0")},
			"0948-5023": []License{
				License("1995000001000000:2002000008000000:0"),
				License("0000000000000000:ZZZZZZZZZZZZZZZZ:0"),
			}},
		}}
	for _, tt := range tests {
		licenses, err := ParseHoldings(tt.r)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(licenses, tt.licenses) {
			t.Errorf("ParseHoldings(%v) => %+v, want %+v", tt.r, licenses, tt.licenses)
		}
	}
}
