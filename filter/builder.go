package filter

import (
	"strings"

	"github.com/miku/span/container"
	"github.com/segmentio/encoding/json"
)

// NewTagger creates an empty Tagger for programmatic assembly.
func NewTagger() *Tagger {
	return &Tagger{FilterMap: make(map[string]Tree)}
}

// Add associates a label (ISIL) with a filter and returns the tagger for
// chaining.
func (t *Tagger) Add(label string, f Filter) *Tagger {
	t.FilterMap[label] = Tree{Root: f}
	return t
}

// MarshalJSON serializes the tagger to its canonical JSON form.
func (t *Tagger) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.FilterMap)
}

// MarshalJSON serializes the tree by delegating to its root filter.
func (t Tree) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Root)
}

// NewTree wraps a filter as a tree.
func NewTree(f Filter) Tree {
	return Tree{Root: f}
}

// --- Leaf filter constructors ---

// Any creates a filter that accepts all records.
func Any() *AnyFilter {
	return &AnyFilter{}
}

// Source creates a filter matching records with any of the given source IDs.
func Source(ids ...string) *SourceFilter {
	return &SourceFilter{Values: ids}
}

// Collection creates a filter matching records in any of the given
// collections.
func Collection(names ...string) *CollectionFilter {
	return &CollectionFilter{Values: container.NewStringSet(names...)}
}

// Subject creates a filter matching records with any of the given subjects.
func Subject(subjects ...string) *SubjectFilter {
	return &SubjectFilter{Values: container.NewStringSet(subjects...)}
}

// Pkg creates a filter matching records in any of the given packages.
func Pkg(packages ...string) *PackageFilter {
	return &PackageFilter{Values: container.NewStringSet(packages...)}
}

// DOI creates a filter matching records with any of the given DOIs.
func DOI(dois ...string) *DOIFilter {
	return &DOIFilter{Values: dois}
}

// ISSN creates a filter matching records with any of the given ISSNs.
func ISSN(issns ...string) *ISSNFilter {
	return &ISSNFilter{Values: container.NewStringSet(issns...)}
}

// ISBN creates a filter matching records with any of the given ISBNs.
func ISBN(isbns ...string) *ISBNFilter {
	return &ISBNFilter{Values: container.NewStringSet(isbns...)}
}

// Holdings creates a holdings filter from file paths and/or URLs. The actual
// KBART data is not loaded; this constructor is meant for building a tree that
// will be serialized to JSON configuration.
func Holdings(names ...string) *HoldingsFilter {
	return &HoldingsFilter{Names: names}
}

// HoldingsWithOpts creates a holdings filter with additional options.
func HoldingsWithOpts(verbose, compareByTitle bool, names ...string) *HoldingsFilter {
	return &HoldingsFilter{
		Names:          names,
		Verbose:        verbose,
		CompareByTitle: compareByTitle,
	}
}

// --- Logic filter constructors ---

// Or creates a filter that matches if any of the given filters match.
func Or(filters ...Filter) *OrFilter {
	return &OrFilter{Filters: filters}
}

// And creates a filter that matches only if all given filters match.
func And(filters ...Filter) *AndFilter {
	return &AndFilter{Filters: filters}
}

// Not creates a filter that inverts another filter.
func Not(f Filter) *NotFilter {
	return &NotFilter{Filter: f}
}

// --- MarshalJSON methods ---

func (f *AnyFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]struct{}{"any": {}})
}

func (f *SourceFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string][]string{"source": f.Values})
}

func (f *CollectionFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string][]string{"collection": f.Values.Values()})
}

func (f *SubjectFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string][]string{"subject": f.Values.Values()})
}

func (f *PackageFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string][]string{"package": f.Values.Values()})
}

func (f *DOIFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		DOI struct {
			List []string `json:"list"`
		} `json:"doi"`
	}{
		DOI: struct {
			List []string `json:"list"`
		}{List: f.Values},
	})
}

func (f *ISSNFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ISSN struct {
			List []string `json:"list"`
		} `json:"issn"`
	}{
		ISSN: struct {
			List []string `json:"list"`
		}{List: f.Values.Values()},
	})
}

func (f *ISBNFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ISBN struct {
			List []string `json:"list"`
		} `json:"isbn"`
	}{
		ISBN: struct {
			List []string `json:"list"`
		}{List: f.Values.Values()},
	})
}

func (f *OrFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string][]Filter{"or": f.Filters})
}

func (f *AndFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string][]Filter{"and": f.Filters})
}

func (f *NotFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]Filter{"not": f.Filter})
}

func (f *HoldingsFilter) MarshalJSON() ([]byte, error) {
	var files, urls []string
	for _, name := range f.Names {
		if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
			urls = append(urls, name)
		} else {
			files = append(files, name)
		}
	}
	return json.Marshal(struct {
		Holdings struct {
			Files          []string `json:"files,omitempty"`
			URLs           []string `json:"urls,omitempty"`
			Verbose        bool     `json:"verbose,omitempty"`
			CompareByTitle bool     `json:"compare-by-title,omitempty"`
		} `json:"holdings"`
	}{
		Holdings: struct {
			Files          []string `json:"files,omitempty"`
			URLs           []string `json:"urls,omitempty"`
			Verbose        bool     `json:"verbose,omitempty"`
			CompareByTitle bool     `json:"compare-by-title,omitempty"`
		}{
			Files:          files,
			URLs:           urls,
			Verbose:        f.Verbose,
			CompareByTitle: f.CompareByTitle,
		},
	})
}
