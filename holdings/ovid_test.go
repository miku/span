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
</holding>`), Licenses{
			"1610-2940": []License{
				License("1995000001000000:2002000008000000:0"),
				License("0000000000000000:ZZZZZZZZZZZZZZZZ:0")},
			"0948-5023": []License{
				License("1995000001000000:2002000008000000:0"),
				License("0000000000000000:ZZZZZZZZZZZZZZZZ:0"),
			}},
		}, {
			strings.NewReader(`
<holding ezb_id = "119">
  <title><![CDATA[Traumatology]]></title>
  <publishers><![CDATA[Sage Publications ; HighWire Press ; Academy of Traumatology]]></publishers>
  <EZBIssns>
    <p-issn>1534-7656</p-issn>
    <e-issn>1085-9373</e-issn>
  </EZBIssns>
  <entitlements>
    <entitlement status = "subscribed">
      <url>http%3A%2F%2Ftmt.sagepub.com%2F</url>
      <anchor>sage_premier</anchor>
      <begin>
        <year>1995</year>
        <volume>1</volume>
      </begin>
      <end>
        <year>2013</year>
        <volume>19</volume>
      </end>
      <available><![CDATA[DFG-geförderte Allianz-Lizenz]]></available>
    </entitlement>
    <entitlement status = "subscribed">
      <url>http%3A%2F%2Ftmt.sagepub.com%2F</url>
      <anchor>natli_sage_archive</anchor>
      <begin>
        <year>1995</year>
        <volume>1</volume>
      </begin>
      <end>
        <year>2013</year>
        <volume>19</volume>
      </end>
      <available><![CDATA[Nationallizenz]]></available>
    </entitlement>
  </entitlements>
</holding>`), Licenses{
				"1534-7656": []License{
					License("1995000001000000:2013000019000000:0")},
				"1085-9373": []License{
					License("1995000001000000:2013000019000000:0")}},
		},
		{
			strings.NewReader(
				`<holding ezb_id = "20">
  <title><![CDATA[Behavioral and Brain Sciences]]></title>
  <publishers><![CDATA[Cambridge University Press]]></publishers>
  <EZBIssns>
    <p-issn>0140-525X</p-issn>
    <e-issn>1469-1825</e-issn>
  </EZBIssns>
  <entitlements>
    <entitlement status = "subscribed">
      <url>http%3A%2F%2Fjournals.cambridge.org%2FBBS</url>
      <anchor>natli_cup</anchor>
      <begin>
        <year>2012</year>
        <volume>35</volume>
        <delay>-2Y</delay>
      </begin>
      <end>
        <delay>-2Y</delay>
      </end>
      <available><![CDATA[Nationallizenz]]></available>
    </entitlement>
    <entitlement status = "subscribed">
      <url>http%3A%2F%2Fjournals.cambridge.org%2Fjid_BBS</url>
      <anchor>cambridge</anchor>
      <begin>
        <year>1997</year>
        <volume>20</volume>
        <issue>1</issue>
      </begin>
      <end>
        <year>2004</year>
        <volume>27</volume>
        <issue>6</issue>
      </end>
    </entitlement>
  </entitlements>
</holding>`), Licenses{
				"0140-525X": []License{
					License("2012000035000000:ZZZZZZZZZZZZZZZZ:-62208000000000000"),
					License("1997000020000001:2004000027000006:0")},
				"1469-1825": []License{
					License("2012000035000000:ZZZZZZZZZZZZZZZZ:-62208000000000000"),
					License("1997000020000001:2004000027000006:0")}},
		},
	}
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

func TestCovers(t *testing.T) {
	var cases = []struct {
		license   License
		signature string
		result    bool
	}{
		{License("2012000035000000:ZZZZZZZZZZZZZZZZ:-62208000000000000"), "2013000035000000", true},
		{License("2012000035000000:ZZZZZZZZZZZZZZZZ:-62208000000000000"), "2011000035000000", false},
		{License("2012000035000000:ZZZZZZZZZZZZZZZZ:-62208000000000000"), "2012000035000000", true},
		{License("2012000035000000:ZZZZZZZZZZZZZZZZ:-62208000000000000"), "2012000034000000", false},
		{License("2012000035000000:ZZZZZZZZZZZZZZZZ:-62208000000000000"), "2018000034000000", false}, // this test will fail in ~2020 :)
	}
	for _, c := range cases {
		r := c.license.Covers(c.signature)
		if r != c.result {
			t.Errorf("Covers(%s) => got %v, want %v", c.signature, r, c.result)
		}
	}
}
