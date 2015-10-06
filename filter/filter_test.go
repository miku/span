//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                 by The Finc Authors, http://finc.info
//                 by Martin Czygan, <martin.czygan@uni-leipzig.de>
//
// This file is part of some open source application.
//
// Some open source application is free software: you can redistribute
// it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation, either
// version 3 of the License, or (at your option) any later version.
//
// Some open source application is distributed in the hope that it will
// be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
//
// @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
//
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
		// A record with date 2014-08-13 and -18M wall gets available on 2016-02-05.
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

// TestHoldingsFilterZeroDelay test IS record from the future and zero delay.
func TestHoldingsFilterZeroDelay(t *testing.T) {
	b := []byte(`
    {
      "rft.issn": [
        "1364-2529"
      ],
      "x.date": "2014-08-13T00:00:00Z"
    }
    `)

	h := `
    <holding ezb_id = "X">
      <EZBIssns>
        <p-issn>1364-2529</p-issn>
      </EZBIssns>
      <entitlements>
        <entitlement status = "subscribed">
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
		{time.Date(1970, 1, 10, 0, 0, 0, 0, time.UTC), true},
		{time.Date(2014, 8, 13, 0, 0, 0, 0, time.UTC), true},
		{time.Date(2016, 8, 13, 0, 0, 0, 0, time.UTC), true},
		{time.Date(2800, 8, 13, 0, 0, 0, 0, time.UTC), true},
	}

	for _, c := range cases {
		hf.Ref = c.t
		r := hf.Apply(is)
		if r != c.r {
			t.Errorf("Apply with ref date %v got %v, want %v", c.t, r, c.r)
		}
	}
}

// TestHoldingsFilterMixedDelay test IS record with inconsistent delays.
func TestHoldingsFilterMixedDelay(t *testing.T) {
	b := []byte(`
    {
      "rft.issn": [
        "1364-2529"
      ],
      "x.date": "2014-01-01T00:00:00Z"
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
          <begin>
            <delay>-1Y</delay>
          </begin>
          <end>
            <delay>-24M</delay>
          </end>
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
		{time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC), false},
		{time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC), false},
		{time.Date(2014, 12, 27, 0, 0, 0, 0, time.UTC), false},
		{time.Date(2014, 12, 28, 0, 0, 0, 0, time.UTC), true},
		{time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), true},
		{time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC), true},
		{time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC), true},
	}

	for _, c := range cases {
		hf.Ref = c.t
		r := hf.Apply(is)
		if r != c.r {
			t.Errorf("Apply with ref date %v got %v, want %v", c.t, r, c.r)
		}
	}
}
