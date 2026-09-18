package kbart

import (
	"io"
	"strings"
	"testing"

	"github.com/miku/span/licensing"
)

const kbartRow = `
publication_title	print_identifier	online_identifier	date_first_issue_online	num_first_vol_online	num_first_issue_online	date_last_issue_online	num_last_vol_online	num_last_issue_online	title_url	first_author	title_id	embargo_info	coverage_depth	coverage_notes	publisher_name	own_anchor	package:collection	il_relevance	il_nationwide	il_electronic_transmission	il_comment	all_issns	zdb_id
 Bill of Rights Journal (via Hein Online)	0006-2499		1968	1		1996	29		http://heinonline.org/HOL/Index?index=journals/blorij&collection=journals		227801		Volltext		via Hein Online		HEIN:hein	Keine Fernleihe				0006-2499	2805467-2
`

func TestDecode(t *testing.T) {
	var cases = []struct {
		about string
		input string
		want  []licensing.Entry
	}{
		{
			"two columns",
			"\npublication_title\tprint_identifier\nHello\t123\n",
			[]licensing.Entry{{PublicationTitle: "Hello", PrintIdentifier: "123"}},
		},
		{
			"full kbart row",
			kbartRow,
			[]licensing.Entry{{
				PublicationTitle:      "Bill of Rights Journal (via Hein Online)",
				PrintIdentifier:       "0006-2499",
				FirstIssueDate:        "1968",
				FirstVolume:           "1",
				LastIssueDate:         "1996",
				LastVolume:            "29",
				TitleURL:              "http://heinonline.org/HOL/Index?index=journals/blorij&collection=journals",
				TitleID:               "227801",
				CoverageDepth:         "Volltext",
				PublisherName:         "via Hein Online",
				PackageCollection:     "HEIN:hein",
				InterlibraryRelevance: "Keine Fernleihe",
				AllSerialNumbers:      "0006-2499",
				ZDBID:                 "2805467-2",
			}},
		},
		{
			// A single ISSN list, refs #11579.
			"single indented column",
			"\n\tonline_identifier\n\t2560-6050\n",
			[]licensing.Entry{{OnlineIdentifier: "2560-6050"}},
		},
		{
			"blank lines, crlf, extra and unknown columns",
			"publication_title\tunknown\n\n\r\nA\tx\n\nB\ty\tz\r\n",
			[]licensing.Entry{{PublicationTitle: "A"}, {PublicationTitle: "B"}},
		},
		{"empty", "", nil},
	}
	for _, c := range cases {
		t.Run(c.about, func(t *testing.T) {
			dec := newDecoder(strings.NewReader(c.input))
			var got []licensing.Entry
			for {
				var entry licensing.Entry
				err := dec.decode(&entry)
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				got = append(got, entry)
			}
			if len(got) != len(c.want) {
				t.Fatalf("got %d entries, want %d: %+v", len(got), len(c.want), got)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("[%d] got %+v, want %+v", i, got[i], c.want[i])
				}
			}
		})
	}
}

func BenchmarkDecode(b *testing.B) {
	for b.Loop() {
		var entry licensing.Entry
		if err := newDecoder(strings.NewReader(kbartRow)).decode(&entry); err != nil {
			b.Fatal(err)
		}
	}
}
