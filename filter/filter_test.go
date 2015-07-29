package filter

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/miku/span/finc"
)

// TestHoldingsFilterApply refs #5675
func TestHoldingsFilterApply(t *testing.T) {
	b := []byte(`
	{
	  "rft.issn": [
	    "1364-2529",
	    "1470-1154"
	  ],
	  "x.date": "2014-08-13T00:00:00Z"
	}
	`)

	h := `
	<holding ezb_id = "23110">
	  <title><![CDATA[Rethinking History. (via EBSCO Host)]]></title>
	  <publishers><![CDATA[via EBSCO Host]]></publishers>
	  <EZBIssns>
	    <p-issn>1364-2529</p-issn>
	  </EZBIssns>
	  <entitlements>
	    <entitlement status = "subscribed">
	      <url>http%3A%2F%2Fsearch.ebscohost.com%2Fdirect.asp%3Fdb%3Daph%26jid%3D5C5%26scope%3Dsite</url>
	      <anchor>ebsco_aph</anchor>
	      <end>
	        <delay>-18M</delay>
	      </end>
	      <available><![CDATA[Academic Search Premier: 1998-03-01 -]]></available>
	    </entitlement>
	  </entitlements>
	</holding>	
	`

	var is finc.IntermediateSchema
	err := json.Unmarshal(b, &is)
	if err != nil {
		t.Error(err)
	}

	hf, err := NewHoldingFilter(strings.NewReader(h))
	if err != nil {
		t.Error(err)
	}

	var cases = []struct {
		t time.Time
		r bool
	}{
		// for a record with date 2014-08-13 and -18M wall expires on 2016-02-05
		{time.Date(2015, 7, 29, 0, 0, 0, 0, time.UTC), false},
		{time.Date(2016, 2, 4, 0, 0, 0, 0, time.UTC), false},
		{time.Date(2016, 2, 5, 0, 0, 0, 0, time.UTC), true},
		{time.Date(2020, 1, 10, 0, 0, 0, 0, time.UTC), true},
	}

	for _, c := range cases {
		hf.Ref = c.t
		r := hf.Apply(is)
		if r != c.r {
			t.Errorf("Apply with ref date %v got %+v, want %v", c.t, r, c.r)
		}
	}
}
