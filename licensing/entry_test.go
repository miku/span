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
