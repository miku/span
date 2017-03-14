package licensing

import (
	"reflect"
	"testing"
)

func TestISSNList(t *testing.T) {
	var cases = []struct {
		entry  Entry
		result []string
	}{
		{
			Entry{PrintIdentifier: "1234-5678"},
			[]string{"1234-5678"},
		},
		{
			Entry{OnlineIdentifier: "1234-5678"},
			[]string{"1234-5678"},
		},
		{
			Entry{PrintIdentifier: "22", OnlineIdentifier: "1234-5678"},
			[]string{"1234-5678"},
		},
		{
			Entry{PrintIdentifier: "22", OwnAnchor: "1234-5678; Hello 2222-222X"},
			[]string{"1234-5678", "2222-222X"},
		},
		{
			Entry{
				OnlineIdentifier: "4444-4444",
				PrintIdentifier:  "3333-3333",
				OwnAnchor:        "1234-5678; Hello 2222-222X",
			},
			[]string{"1234-5678", "2222-222X", "3333-3333", "4444-4444"},
		},
	}

	for _, c := range cases {
		result := c.entry.ISSNList()
		if !reflect.DeepEqual(result, c.result) {
			t.Errorf("ISSNList: got %v, want %v", result, c.result)
		}
	}
}

func TestCoversDate(t *testing.T) {
	var cases = []struct {
		entry Entry
		value string
		ok    bool
		err   error
	}{
		{
			entry: Entry{
				FirstIssueDate: "1992-06-23",
				LastIssueDate:  "1998-04-10",
			},
			value: "1997",
			ok:    true,
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "",
				LastIssueDate:  "1998-04-10",
			},
			value: "1997",
			ok:    true,
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "",
				LastIssueDate:  "1998-04-10",
			},
			value: "1998-04",
			ok:    true,
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "",
				LastIssueDate:  "",
			},
			value: "1998-04",
			ok:    true,
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "2000",
				LastIssueDate:  "",
			},
			value: "1999-12-31",
			ok:    false,
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "2000",
				LastIssueDate:  "",
			},
			value: "",
			ok:    false,
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "",
				LastIssueDate:  "",
			},
			value: "",
			ok:    false,
			err:   nil,
		},
		{
			entry: Entry{
				FirstIssueDate: "1992-02-10",
				LastIssueDate:  "1992-09-10",
			},
			value: "1992",
			ok:    false,
			err:   nil,
		},
	}

	for _, c := range cases {
		ok, err := c.entry.CoversDate(c.value)
		if err != c.err {
			t.Errorf("CoversDate: got %v, want %v", err, c.err)
		}
		if ok != c.ok {
			t.Errorf("CoversDate: got %v, want %v", ok, c.ok)
		}
	}
}

func BenchmarkCoversDate(b *testing.B) {
	entry := Entry{
		FirstIssueDate: "1992-06-23",
		LastIssueDate:  "1998-04-10",
	}
	v := "1992"
	for i := 0; i < b.N; i++ {
		entry.CoversDate(v)
	}
}
