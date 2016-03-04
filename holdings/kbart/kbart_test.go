package kbart

import (
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kr/pretty"
	"github.com/miku/span/holdings"
)

func TestEmbargeDisallowEarlier(t *testing.T) {
	var cases = []struct {
		e               embargo
		disallowEarlier bool
	}{
		{embargo("1"), false},
		{embargo("R1"), true},
		{embargo("R1D"), true},
		{embargo("R10M"), true},
		{embargo("?10M"), false},
	}

	for _, c := range cases {
		got := c.e.DisallowEarlier()
		if got != c.disallowEarlier {
			t.Errorf("embargo.DisallowEarlier() got %v, want %v", got, c.disallowEarlier)
		}
	}
}

func TestEmbargoDuration(t *testing.T) {
	var cases = []struct {
		e   embargo
		d   time.Duration
		err error
	}{
		{embargo("1"), time.Duration(0), nil},
		{embargo("R1"), time.Duration(0), nil},
		{embargo("R1D"), -24 * time.Hour, nil},
		{embargo("R10M"), -7200 * time.Hour, nil},
		{embargo("?10M"), time.Duration(0), nil},
	}

	for _, c := range cases {
		got, err := c.e.AsDuration()
		if err != c.err {
			t.Errorf("embargo.AsDuration() got %v, want %v", err, c.err)
		}
		if got != c.d {
			t.Errorf("embargo.DisallowEarlier() got %v, want %v", got, c.d)
		}
	}
}

func TestFromReader(t *testing.T) {
	var cases = []struct {
		r       io.Reader
		entries holdings.Entries
		err     error
	}{
		{r: strings.NewReader("first line is discarded"),
			entries: make(holdings.Entries),
			err:     nil},
		{r: strings.NewReader("xxx\nyyy\nzzz"),
			entries: make(holdings.Entries),
			err:     nil},
		// Beware: KBART files must end with newline, otherwise the last row is ignored.
		{r: strings.NewReader(`publication_title	print_identifier	online_identifier	date_first_issue_online	num_first_vol_online	num_first_issue_online	date_last_issue_online	num_last_vol_online	num_last_issue_online	title_url	first_author	title_id	embargo_info	coverage_depth	coverage_notes	publisher_name	own_anchor	il_relevance	il_nationwide	il_electronic_transmission	il_comment	all_issns	zdb_id
Bill of Rights Journal (via Hein Online)	0006-2499		1968	1		1996	29		http://heinonline.org/HOL/Index?index=journals/blorij&collection=journals		227801		Volltext		via Hein Online		Keine Fernleihe				0006-2499	2805467-2
`),
			entries: holdings.Entries{
				"0006-2499": []holdings.License{
					holdings.Entry{
						Begin: holdings.Signature{
							Date:   "1968",
							Volume: "1",
							Issue:  "",
						},
						End: holdings.Signature{
							Date:   "1996",
							Volume: "29",
							Issue:  "",
						},
						Embargo:                time.Duration(0),
						EmbargoDisallowEarlier: false,
					}}},
			err: nil},
		// Beware: KBART files must end with newline, otherwise the last row is ignored.
		{r: strings.NewReader(`publication_title	print_identifier	online_identifier	date_first_issue_online	num_first_vol_online	num_first_issue_online	date_last_issue_online	num_last_vol_online	num_last_issue_online	title_url	first_author	title_id	embargo_info	coverage_depth	coverage_notes	publisher_name	own_anchor	il_relevance	il_nationwide	il_electronic_transmission	il_comment	all_issns	zdb_id

Bill of Rights Journal (via Hein Online)	0006-2499		1968	1		1996	29		http://heinonline.org/HOL/Index?index=journals/blorij&collection=journals		227801		Volltext		via Hein Online		Keine Fernleihe				0006-2499	2805467-2
`),
			entries: holdings.Entries{
				"0006-2499": []holdings.License{
					holdings.Entry{
						Begin: holdings.Signature{
							Date:   "1968",
							Volume: "1",
							Issue:  "",
						},
						End: holdings.Signature{
							Date:   "1996",
							Volume: "29",
							Issue:  "",
						},
						Embargo:                time.Duration(0),
						EmbargoDisallowEarlier: false,
					}}},
			err: nil},
	}

	for _, c := range cases {

		reader := NewReader(c.r)
		reader.SkipFirstRow = true

		entries, err := reader.ReadEntries()
		if err != c.err {
			t.Errorf("ReadEntries got %+v, want %+v", err, c.err)
		}
		if !reflect.DeepEqual(entries, c.entries) {
			t.Errorf("ReadEntries got %+v, want %+v", entries, c.entries)
			for _, s := range pretty.Diff(c.entries, entries) {
				t.Errorf(s)
			}
		}
	}
}
