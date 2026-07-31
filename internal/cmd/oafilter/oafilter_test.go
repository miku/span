package oafilter

import (
	"testing"

	"github.com/miku/span/formats/finc"
)

// fakeApplier returns a fixed decision.
type fakeApplier struct{ result bool }

func (f fakeApplier) Apply(is finc.IntermediateSchema) bool { return f.result }

func TestSetOpenAccess(t *testing.T) {
	var cases = []struct {
		name           string
		sourceID       string
		megaColls      []string
		applier        oaApplier
		lookup         FreeContentLookup
		excludeSids    map[string]bool
		openAccessSids map[string]bool
		want           bool
	}{
		{
			name:           "oa sid forces true",
			sourceID:       "49",
			applier:        fakeApplier{false},
			openAccessSids: map[string]bool{"49": true},
			want:           true,
		},
		{
			name:        "excluded sid stays false",
			sourceID:    "49",
			applier:     fakeApplier{true},
			excludeSids: map[string]bool{"49": true},
			want:        false,
		},
		{
			name:     "filter apply sets true",
			sourceID: "50",
			applier:  fakeApplier{true},
			want:     true,
		},
		{
			name:     "filter apply false stays false",
			sourceID: "50",
			applier:  fakeApplier{false},
			want:     false,
		},
		{
			name:      "free content lookup true overrides",
			sourceID:  "50",
			megaColls: []string{"Foo"},
			applier:   fakeApplier{false},
			lookup:    FreeContentLookup{"50:Foo": true},
			want:      true,
		},
		{
			name:      "free content lookup false overrides filter true",
			sourceID:  "50",
			megaColls: []string{"Foo"},
			applier:   fakeApplier{true},
			lookup:    FreeContentLookup{"50:Foo": false},
			want:      false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			is := &finc.IntermediateSchema{
				SourceID:        c.sourceID,
				MegaCollections: c.megaColls,
			}
			exclude := c.excludeSids
			if exclude == nil {
				exclude = map[string]bool{}
			}
			oa := c.openAccessSids
			if oa == nil {
				oa = map[string]bool{}
			}
			setOpenAccess(is, c.applier, c.lookup, exclude, oa)
			if is.OpenAccess != c.want {
				t.Errorf("OpenAccess = %v, want %v", is.OpenAccess, c.want)
			}
		})
	}
}

func TestToSet(t *testing.T) {
	s := toSet([]string{"a", "b", "a"})
	if len(s) != 2 || !s["a"] || !s["b"] {
		t.Errorf("toSet = %v", s)
	}
}
