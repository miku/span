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
	  "finc.format": "ElectronicArticle",
	  "finc.mega_collection": "Informa UK Limited (CrossRef)",
	  "finc.record_id": "ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4MC8xMzY0MjUyOS4yMDE0LjkzMjA3OQ",
	  "finc.source_id": "49",
	  "ris.type": "EJOUR",
	  "rft.atitle": "History, power and visual communication artefacts",
	  "rft.epage": "22",
	  "rft.genre": "article",
	  "rft.issn": [
	    "1364-2529",
	    "1470-1154"
	  ],
	  "rft.jtitle": "Rethinking History",
	  "rft.tpages": "21",
	  "rft.pages": "1-22",
	  "rft.pub": [
	    "Informa UK Limited"
	  ],
	  "rft.date": "2014-08-13",
	  "x.date": "2014-08-13T00:00:00Z",
	  "rft.spage": "1",
	  "authors": [
	    {
	      "rft.aulast": "Hepworth",
	      "rft.aufirst": "Katherine"
	    }
	  ],
	  "doi": "10.1080/13642529.2014.932079",
	  "languages": [
	    "eng"
	  ],
	  "url": [
	    "http://dx.doi.org/10.1080/13642529.2014.932079"
	  ],
	  "version": "0.9",
	  "x.subjects": [
	    "History"
	  ],
	  "x.type": "journal-article"
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

	hf.Ref = time.Now()
	result := hf.Apply(is)
	if result != false {
		t.Errorf("Apply got %+v, want %v", result, false)
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
