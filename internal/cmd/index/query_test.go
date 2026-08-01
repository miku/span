package index

import "testing"

func TestBuildQuery(t *testing.T) {
	cases := []struct {
		name string
		qf   *queryFlags
		want string
	}{
		{"empty", &queryFlags{}, "*:*"},
		{"raw q overrides", &queryFlags{q: "id:abc", sid: "53"}, "id:abc"},
		{"sid", &queryFlags{sid: "53"}, `source_id:"53"`},
		{
			"until absolute",
			&queryFlags{until: "2026-01-01"},
			"last_indexed:[* TO 2026-01-01T00:00:00Z]",
		},
		{
			"since absolute",
			&queryFlags{since: "2026-01-01"},
			"last_indexed:[2026-01-01T00:00:00Z TO *]",
		},
		{
			"sid and until",
			&queryFlags{sid: "53", until: "2026-01-01"},
			`source_id:"53" AND last_indexed:[* TO 2026-01-01T00:00:00Z]`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildQuery(tt.qf)
			if err != nil {
				t.Fatalf("buildQuery: %v", err)
			}
			if got != tt.want {
				t.Errorf("buildQuery() = %q, want %q", got, tt.want)
			}
		})
	}
}
