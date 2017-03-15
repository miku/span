package csv

import (
	"encoding/csv"
	"reflect"
	"strings"
	"testing"
)

var (
	testOne = `
publication_title	print_identifier
Hello	123
`

	testTwo = `
publication_title	print_identifier	online_identifier	date_first_issue_online	num_first_vol_online	num_first_issue_online	date_last_issue_online	num_last_vol_online	num_last_issue_online	title_url	first_author	title_id	embargo_info	coverage_depth	coverage_notes	publisher_name	own_anchor	package:collection	il_relevance	il_nationwide	il_electronic_transmission	il_comment	all_issns	zdb_id
 Bill of Rights Journal (via Hein Online)	0006-2499		1968	1		1996	29		http://heinonline.org/HOL/Index?index=journals/blorij&collection=journals		227801		Volltext		via Hein Online		HEIN:hein	Keine Fernleihe				0006-2499	2805467-2
`
)

type TestSimple struct {
	Title string `csv:"publication_title"`
	ID    string `csv:"print_identifier"`
}

type TestRepetition struct {
	Title       string `csv:"publication_title"`
	Publication string `csv:"publication_title"`
	ID          string `csv:"print_identifier"`
}

type TestEntry struct {
	PublicationTitle                   string `csv:"publication_title"`
	PrintIdentifier                    string `csv:"print_identifier"`
	OnlineIdentifier                   string `csv:"online_identifier"`
	FirstIssueDate                     string `csv:"date_first_issue_online"`
	FirstVolume                        string `csv:"num_first_vol_online"`
	FirstIssue                         string `csv:"num_first_issue_online"`
	LastIssueDate                      string `csv:"date_last_issue_online"`
	LastVolume                         string `csv:"num_last_vol_online"`
	LastIssue                          string `csv:"num_last_issue_online"`
	TitleURL                           string `csv:"title_url"`
	FirstAuthor                        string `csv:"first_author"`
	TitleID                            string `csv:"title_id"`
	Embargo                            string `csv:"embargo_info"`
	CoverageDepth                      string `csv:"coverage_depth"`
	CoverageNotes                      string `csv:"coverage_notes"`
	PublisherName                      string `csv:"publisher_name"`
	OwnAnchor                          string `csv:"own_anchor"`
	PackageCollection                  string `csv:"package:collection"`
	InterlibraryRelevance              string `csv:"il_relevance"`
	InterlibraryNationwide             string `csv:"il_nationwide"`
	InterlibraryElectronicTransmission string `csv:"il_electronic_transmission"`
	InterlibraryComment                string `csv:"il_comment"`
	AllSerialNumbers                   string `csv:"all_issns"`
	ZDBID                              string `csv:"zdb_id"`
	Location                           string `csv:"location"`
	TitleNotes                         string `csv:"title_notes"`
	StaffNotes                         string `csv:"staff_notes"`
	VendorID                           string `csv:"vendor_id"`
	OCLCCollectionName                 string `csv:"oclc_collection_name"`
	OCLCCollectionID                   string `csv:"oclc_collection_id"`
	OCLCEntryID                        string `csv:"oclc_entry_id"`
	OCLCLinkScheme                     string `csv:"oclc_link_scheme"`
	OCLCNumber                         string `csv:"oclc_number"`
	Action                             string `csv:"ACTION"`
}

func TestDecode(t *testing.T) {
	r := csv.NewReader(strings.NewReader(testOne))
	r.Comma = '\t'

	dec := NewDecoder(r)

	var example TestSimple
	if err := dec.Decode(&example); err != nil {
		t.Errorf(err.Error())
	}

	expected := TestSimple{Title: "Hello", ID: "123"}
	if !reflect.DeepEqual(example, expected) {
		t.Errorf("Decode: got %#v, want %#v", example, expected)
	}
}

func TestDecodeRepetitions(t *testing.T) {
	r := csv.NewReader(strings.NewReader(testOne))
	r.Comma = '\t'

	dec := NewDecoder(r)

	var example TestRepetition
	if err := dec.Decode(&example); err != nil {
		t.Errorf(err.Error())
	}

	expected := TestRepetition{Title: "Hello", Publication: "Hello", ID: "123"}
	if !reflect.DeepEqual(example, expected) {
		t.Errorf("Decode: got %#v, want %#v", example, expected)
	}
}

func TestDecodeKbart(t *testing.T) {
	r := csv.NewReader(strings.NewReader(testTwo))
	r.Comma = '\t'

	dec := NewDecoder(r)

	var example TestEntry
	if err := dec.Decode(&example); err != nil {
		t.Errorf(err.Error())
	}

	expected := TestEntry{
		PublicationTitle:                   " Bill of Rights Journal (via Hein Online)",
		PrintIdentifier:                    "0006-2499",
		OnlineIdentifier:                   "",
		FirstIssueDate:                     "1968",
		FirstVolume:                        "1",
		FirstIssue:                         "",
		LastIssueDate:                      "1996",
		LastVolume:                         "29",
		LastIssue:                          "",
		TitleURL:                           "http://heinonline.org/HOL/Index?index=journals/blorij&collection=journals",
		FirstAuthor:                        "",
		TitleID:                            "227801",
		Embargo:                            "",
		CoverageDepth:                      "Volltext",
		CoverageNotes:                      "",
		PublisherName:                      "via Hein Online",
		OwnAnchor:                          "",
		PackageCollection:                  "HEIN:hein",
		InterlibraryRelevance:              "Keine Fernleihe",
		InterlibraryNationwide:             "",
		InterlibraryElectronicTransmission: "",
		InterlibraryComment:                "",
		AllSerialNumbers:                   "0006-2499",
		ZDBID:                              "2805467-2",
	}
	if !reflect.DeepEqual(example, expected) {
		t.Errorf("Decode: got %#v, want %#v", example, expected)
	}
}

func BenchmarkDecodeKbart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := csv.NewReader(strings.NewReader(testTwo))
		r.Comma = '\t'
		dec := NewDecoder(r)
		var example TestEntry
		dec.Decode(&example)
	}
}

// $ go test -v github.com/miku/span/encoding/csv -bench=.
// === RUN   TestDecode
// --- PASS: TestDecode (0.00s)
// === RUN   TestDecodeRepetitions
// --- PASS: TestDecodeRepetitions (0.00s)
// === RUN   TestDecodeKbart
// --- PASS: TestDecodeKbart (0.00s)
// BenchmarkDecodeKbart-4   	   20000	     69255 ns/op
// PASS
// ok  	github.com/miku/span/encoding/csv	2.039s
