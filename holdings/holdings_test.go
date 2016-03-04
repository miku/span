package holdings

import "testing"

func BenchmarkEntryCoversFull(b *testing.B) {
	entry := Entry{
		Begin: Signature{Date: "2009", Volume: "10", Issue: "123"},
		End:   Signature{Date: "2011", Volume: "12", Issue: "234"}}
	s := Signature{Date: "2009", Volume: "11", Issue: "124"}
	for i := 0; i < b.N; i++ {
		_ = entry.Covers(s)
	}
}

func BenchmarkEntryCoversSimple(b *testing.B) {
	entry := Entry{
		Begin: Signature{Date: "2009", Volume: "", Issue: ""},
		End:   Signature{Date: "2011", Volume: "", Issue: ""}}
	s := Signature{Date: "1999", Volume: "", Issue: "0"}
	for i := 0; i < b.N; i++ {
		_ = entry.Covers(s)
	}
}

func BenchmarkEntryCoversEmpty(b *testing.B) {
	entry := Entry{
		Begin: Signature{Date: "", Volume: "", Issue: ""},
		End:   Signature{Date: "", Volume: "", Issue: ""}}
	s := Signature{Date: "", Volume: "", Issue: ""}
	for i := 0; i < b.N; i++ {
		_ = entry.Covers(s)
	}
}

func TestEntryCovers(t *testing.T) {
	var tests = []struct {
		description string
		entry       Entry
		s           Signature
		err         error
	}{
		{
			description: "if nothing is defined, we report missing values",
			entry: Entry{
				Begin: Signature{Date: "", Volume: "", Issue: ""},
				End:   Signature{Date: "", Volume: "", Issue: ""}},
			s:   Signature{Date: "", Volume: "", Issue: ""},
			err: ErrMissingValues,
		},
		{
			description: "if the record has no date, it will report an error, since we cannot determine coverage",
			entry: Entry{
				Begin: Signature{Date: "2010", Volume: "", Issue: ""},
				End:   Signature{Date: "2011", Volume: "", Issue: ""}},
			s:   Signature{Date: "", Volume: "", Issue: ""},
			err: ErrMissingValues,
		},
		{
			description: "partial holding time spans are ok",
			entry: Entry{
				Begin: Signature{Date: "2011", Volume: "", Issue: ""},
				End:   Signature{Date: "", Volume: "", Issue: ""}},
			s:   Signature{Date: "2011", Volume: "", Issue: ""},
			err: nil,
		},
		{
			description: "partial holding time spans are ok",
			entry: Entry{
				Begin: Signature{Date: "", Volume: "", Issue: ""},
				End:   Signature{Date: "2011", Volume: "", Issue: ""}},
			s:   Signature{Date: "2011", Volume: "", Issue: ""},
			err: nil,
		},
		{
			description: "partial holding time spans are ok",
			entry: Entry{
				Begin: Signature{Date: "2011", Volume: "", Issue: ""},
				End:   Signature{Date: "", Volume: "", Issue: ""}},
			s:   Signature{Date: "2012", Volume: "", Issue: ""},
			err: nil,
		},
		{
			description: "partial holding time spans are ok",
			entry: Entry{
				Begin: Signature{Date: "", Volume: "", Issue: ""},
				End:   Signature{Date: "2011", Volume: "", Issue: ""}},
			s:   Signature{Date: "2012", Volume: "", Issue: ""},
			err: ErrAfterCoverageInterval,
		},
		{
			description: "partial holding time spans are ok",
			entry: Entry{
				Begin: Signature{Date: "", Volume: "", Issue: ""},
				End:   Signature{Date: "2011", Volume: "", Issue: ""}},
			s:   Signature{Date: "2012", Volume: "", Issue: ""},
			err: ErrAfterCoverageInterval,
		},
		{
			description: "begin, end and record date are defined, record is too late",
			entry: Entry{
				Begin: Signature{Date: "2010", Volume: "", Issue: ""},
				End:   Signature{Date: "2011", Volume: "", Issue: ""}},
			s:   Signature{Date: "2012", Volume: "", Issue: ""},
			err: ErrAfterCoverageInterval,
		},
		{
			description: "begin, end and record date are defined, record is too early",
			entry: Entry{
				Begin: Signature{Date: "2010", Volume: "", Issue: ""},
				End:   Signature{Date: "2011", Volume: "", Issue: ""}},
			s:   Signature{Date: "2009", Volume: "", Issue: ""},
			err: ErrBeforeCoverageInterval,
		},
		{
			description: "begin, end and record date are defined, record is covered",
			entry: Entry{
				Begin: Signature{Date: "2010", Volume: "", Issue: ""},
				End:   Signature{Date: "2011", Volume: "", Issue: ""}},
			s:   Signature{Date: "2010", Volume: "", Issue: ""},
			err: nil,
		},
		{
			description: "report error, if entries define no date",
			entry: Entry{
				Begin: Signature{Date: "", Volume: "1", Issue: ""},
				End:   Signature{Date: "", Volume: "2", Issue: ""}},
			s:   Signature{Date: "2010", Volume: "", Issue: ""},
			err: ErrMissingValues,
		},
		{
			description: "pass, if date matches, but record has no volume information",
			entry: Entry{
				Begin: Signature{Date: "2009", Volume: "1", Issue: ""},
				End:   Signature{Date: "2011", Volume: "2", Issue: ""}},
			s:   Signature{Date: "2010", Volume: "", Issue: ""},
			err: nil,
		},
		{
			description: "pass, if date is covered, volume is covered",
			entry: Entry{
				Begin: Signature{Date: "2009", Volume: "1", Issue: ""},
				End:   Signature{Date: "2011", Volume: "2", Issue: ""}},
			s:   Signature{Date: "2010", Volume: "1", Issue: ""},
			err: nil,
		},
		{
			description: "pass, if date is covered, volume too late",
			entry: Entry{
				Begin: Signature{Date: "2009", Volume: "1", Issue: ""},
				End:   Signature{Date: "2011", Volume: "2", Issue: ""}},
			s:   Signature{Date: "2010", Volume: "3", Issue: ""},
			err: ErrAfterCoverageInterval,
		},
		{
			description: "pass, if date is covered, volume too early",
			entry: Entry{
				Begin: Signature{Date: "2009", Volume: "10", Issue: ""},
				End:   Signature{Date: "2011", Volume: "12", Issue: ""}},
			s:   Signature{Date: "2009", Volume: "9", Issue: ""},
			err: ErrBeforeCoverageInterval,
		},
		{
			description: "deny, even though signature volume is weird",
			entry: Entry{
				Begin: Signature{Date: "2009", Volume: "10", Issue: ""},
				End:   Signature{Date: "2011", Volume: "12", Issue: ""}},
			s:   Signature{Date: "2009", Volume: "Vol. 9", Issue: ""},
			err: ErrBeforeCoverageInterval,
		},
		{
			description: "pass, even though signature volume is weird",
			entry: Entry{
				Begin: Signature{Date: "2009", Volume: "10", Issue: ""},
				End:   Signature{Date: "2011", Volume: "12", Issue: ""}},
			s:   Signature{Date: "2009", Volume: "Vol. 11", Issue: ""},
			err: nil,
		},
		{
			description: "pass, even though signature volume is weirder",
			entry: Entry{
				Begin: Signature{Date: "2009", Volume: "10", Issue: ""},
				End:   Signature{Date: "2011", Volume: "12", Issue: ""}},
			s:   Signature{Date: "2009", Volume: "11. Ausgabe vom 100", Issue: ""},
			err: nil,
		},
		{
			description: "fail, even though signature volume is correct, but too weird",
			entry: Entry{
				Begin: Signature{Date: "2009", Volume: "10", Issue: ""},
				End:   Signature{Date: "2011", Volume: "12", Issue: ""}},
			s:   Signature{Date: "2009", Volume: "100 Total Vol 6", Issue: ""},
			err: ErrAfterCoverageInterval,
		},
	}

	for _, test := range tests {
		err := test.entry.Covers(test.s)
		if err != test.err {
			t.Errorf("Covers got %v, want %v, description: %s", err, test.err, test.description)
		}
	}
}
