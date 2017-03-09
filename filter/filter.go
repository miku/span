// Flexible ISIL attachments with expression trees[1], serialized as JSON. The
// top-level key is the label, that is to be given to a record. Here, this label
// is an ISIL. Each ISIL specifies a tree of filters. Intermediate nodes can be
// "or" and "and" filters, leaf nodes contain custom filter, that are matched
// against records.
//
// A filter implements filter.Filter with a two methods: Apply and optionally
// UnmarshalJSON, so we can load the configuration from a file. Each filter can
// define arbitrary options, e.g. a HoldingsFilter can be loaded from a single
// file or a list of urls.
//
// [1] https://en.wikipedia.org/wiki/Binary_expression_tree#Boolean_expressions
//
// The simplest filter is one, that says yes to all records:
//
//     "DE-X": { "any": {} }
//
// On the command line:
//
//     $ span-tag -c '{"DE-X": {"any": {}}}' in.ldj > out.ldj
//
// Another, more complex example:
//
//     "DE-14": {
//       "or": [
//         {
//           "and": [
//             {
//               "source": [
//                 "55"
//               ]
//             },
//             {
//               "holdings": {
//                 "urls": [
//                   "http://www.jstor.org/kbart/collections/asii",
//                   "http://www.jstor.org/kbart/collections/as"
//                 ]
//               }
//             }
//           ]
//         },
//         {
//           "and": [
//             {
//               "source": [
//                 "49"
//               ]
//             },
//             {
//               "holdings": {
//                 "urls": [
//                   "https://example.com/docs/KBART_DE14",
//                   "https://example.com/docs/KBART_FREEJOURNALS"
//                 ]
//               }
//             },
//       ...
//     }
//
// Additional filters must be registered in `unmarshalFilter`.
//
// ----
//
// TODO(miku): Add a small cache to reuse holding file filters, used in multiple places.
// TODO(miku): Simplify some too long Unmarshaler, create filterutil if necessary.
//
// Faster evaluation? http://theory.stanford.edu/~sergei/papers/sigmod10-index.pdf

package filter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/miku/span"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
	"github.com/miku/span/holdings"
	"github.com/miku/span/holdings/generic"
)

// Filter returns go or no for a given record.
type Filter interface {
	Apply(finc.IntermediateSchema) bool
}

// AnyFilter validates any record.
type AnyFilter struct {
	Any struct{} `json:"any"`
}

// Apply will just return true.
func (f *AnyFilter) Apply(finc.IntermediateSchema) bool { return true }

// CollectionFilter validates all records matching one of the given collections.
type CollectionFilter struct {
	values container.StringSet
}

// Apply filters collections.
func (f *CollectionFilter) Apply(is finc.IntermediateSchema) bool {
	return f.values.Contains(is.MegaCollection)
}

// UnmarshalJSON turns a config fragment into a ISSN filter.
func (f *CollectionFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Collections []string `json:"collection"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.values = *container.NewStringSet(s.Collections...)
	return nil
}

// ISSNFilter allows records with a certain ISSN.
type ISSNFilter struct {
	values container.StringSet
}

// Apply applies ISSN filter on intermediate
// schema, no distinction between ISSN and EISSN.
func (f *ISSNFilter) Apply(is finc.IntermediateSchema) bool {
	for _, issn := range append(is.ISSN, is.EISSN...) {
		if f.values.Contains(issn) {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *ISSNFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		ISSN struct {
			Values []string `json:"list"`
			File   string   `json:"file"`
			Link   string   `json:"url"`
		} `json:"issn"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.values = *container.NewStringSet()

	if s.ISSN.Link != "" {
		slink := span.SavedLink{Link: s.ISSN.Link}
		filename, err := slink.Save()
		if err != nil {
			return err
		}
		defer slink.Remove()
		s.ISSN.File = filename
	}

	if s.ISSN.File != "" {
		file, err := os.Open(s.ISSN.File)
		if err != nil {
			return err
		}
		defer file.Close()
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// valid ISSN can contain x, normalize to uppercase
			line = strings.ToUpper(line)

			// sniff ISSNs
			issns := container.NewStringSet()
			for _, s := range span.ISSNPattern.FindAllString(line, -1) {
				issns.Add(s)
			}

			if issns.Size() == 0 {
				log.Printf("warning: no ISSNs found on line: %s", line)
			}
			for _, issn := range issns.Values() {
				f.values.Add(issn)
			}
		}
	}
	for _, v := range s.ISSN.Values {
		f.values.Add(v)
	}
	log.Printf("loaded %d ISSN from list", f.values.Size())
	return nil
}

// SourceFilter allows all records with the given source id or ids.
type SourceFilter struct {
	values []string
}

// Apply filter.
func (f *SourceFilter) Apply(is finc.IntermediateSchema) bool {
	for _, v := range f.values {
		if v == is.SourceID {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *SourceFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Sources []string `json:"source"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.values = s.Sources
	return nil
}

// PackageFilter allows all records of one of the given package name.
type PackageFilter struct {
	values container.StringSet
}

// Apply filters packages.
func (f *PackageFilter) Apply(is finc.IntermediateSchema) bool {
	for _, pkg := range is.Packages {
		if f.values.Contains(pkg) {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *PackageFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Packages []string `json:"package"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	f.values = *container.NewStringSet(s.Packages...)
	return nil
}

// DOIFilter allows records with a given DOI. Use in conjuction with "not" to
// create blacklists.
type DOIFilter struct {
	values []string
}

// Apply filters packages.
func (f *DOIFilter) Apply(is finc.IntermediateSchema) bool {
	for _, v := range f.values {
		if v == is.DOI {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a filter.
func (f *DOIFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		DOI struct {
			Values []string `json:"list"`
			File   string   `json:"file"`
		} `json:"doi"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	var values []string
	if s.DOI.File != "" {

		log.Println(s.DOI.File)

		file, err := os.Open(s.DOI.File)
		if err != nil {
			return err
		}
		defer file.Close()
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			values = append(values, line)
		}
	}
	for _, v := range s.DOI.Values {
		values = append(values, v)
	}
	f.values = values
	return nil
}

// HoldingsFilter filters a record against a holding file. The holding file
// might be in KBART, Ovid or Google format.
type HoldingsFilter struct {
	entries holdings.Entries
	Verbose bool
}

// Apply tests validity against holding file. TODO(miku): holdings file
// indentifiers can be ISSNs, ISBNs or DOIs.
func (f *HoldingsFilter) Apply(is finc.IntermediateSchema) bool {
	signature := holdings.Signature{
		Date:   is.Date.Format("2006-01-02"),
		Volume: is.Volume,
		Issue:  is.Issue,
	}

	var err error

	for _, issn := range append(is.ISSN, is.EISSN...) {
		for _, license := range f.entries.Licenses(issn) {
			if err = license.Covers(signature); err == nil {
				return true
			}
			if !f.Verbose {
				continue
			}
			b, merr := json.MarshalIndent(map[string]interface{}{
				"document": is,
				"err":      err.Error(),
				"issn":     issn,
				"license":  license}, "", "    ")
			if merr == nil {
				log.Printf(string(b))
			} else {
				log.Printf("cannot even serialize document: %s", merr)
			}
		}
	}
	return false
}

// UnmarshalJSON unwraps a JSON into a HoldingsFilter. Can use holding file from
// file or a list of URLs, if both are given, only the file is used.
// TODO(miku): Allow multiple files as well.
func (f *HoldingsFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Holdings struct {
			Filename string   `json:"file"`
			Links    []string `json:"urls"`
			Verbose  bool     `json:"verbose"`
		} `json:"holdings"`
	}

	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}

	f.Verbose = s.Holdings.Verbose

	// concatenated downloaded and possible extracted links
	concatenated, err := ioutil.TempFile("", "span-")
	defer os.Remove(concatenated.Name()) // clean up

	for _, link := range s.Holdings.Links {
		if _, err := io.Copy(concatenated, &span.ZipOrPlainLinkReader{Link: link}); err != nil {
			return err
		}
	}

	if err := concatenated.Close(); err != nil {
		return err
	}

	filename := concatenated.Name()

	if s.Holdings.Filename != "" {
		filename = s.Holdings.Filename
	}

	if filename == "" {
		return fmt.Errorf("holdings filter: either file or urls must be given")
	}

	file, err := generic.New(filename)
	if err != nil {
		return err
	}

	f.entries, err = file.ReadEntries()
	return err
}

// OrFilter returns true, if there is at least one filter matching.
type OrFilter struct {
	filters []Filter
}

// Apply returns true, if any of the filters returns true. Short circuited.
func (f *OrFilter) Apply(is finc.IntermediateSchema) bool {
	for _, f := range f.filters {
		if f.Apply(is) {
			return true
		}
	}
	return false
}

// UnmarshalJSON turns a config fragment into a or filter.
func (f *OrFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Filters []json.RawMessage `json:"or"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	filters, err := unmarshalFilterList(s.Filters)
	if err != nil {
		return err
	}
	f.filters = filters
	return nil
}

// unmarshalFilter takes the name of a filter and a raw JSON message and
// unmarshals the appropriate filter. All filters must be registered here.
// Unknown filters cause an error.
func unmarshalFilter(name string, raw json.RawMessage) (Filter, error) {
	switch name {
	// add more filters here
	case "any":
		var filter AnyFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "doi":
		var filter DOIFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "issn":
		var filter ISSNFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "package":
		var filter PackageFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "holdings":
		var filter HoldingsFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "collection":
		var filter CollectionFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "source":
		var filter SourceFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "or":
		var filter OrFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "and":
		var filter AndFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	case "not":
		var filter NotFilter
		if err := json.Unmarshal(raw, &filter); err != nil {
			return nil, err
		}
		return &filter, nil
	default:
		return nil, fmt.Errorf("unknown filter: %s", name)
	}
}

// firstKey returns the top level key of an object, given as a raw JSON message.
// It peeks into the fragment. An empty document will cause an error, as
// multiple top level keys.
func firstKey(raw json.RawMessage) (string, error) {
	var peeker = make(map[string]interface{})
	if err := json.Unmarshal(raw, &peeker); err != nil {
		return "", err
	}
	var keys []string
	for k, _ := range peeker {
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return "", fmt.Errorf("no key found")
	}
	if len(keys) > 1 {
		return "", fmt.Errorf("cannot decide which top-level key to use: %s", keys)
	}
	return keys[0], nil
}

// unmarshalFilterList returns filters from a list of JSON fragments. Unknown
// filter names will cause an error.
func unmarshalFilterList(r []json.RawMessage) ([]Filter, error) {
	var filters []Filter
	for _, raw := range r {
		name, err := firstKey(raw)
		if err != nil {
			return filters, err
		}
		filter, err := unmarshalFilter(name, raw)
		if err != nil {
			return filters, err
		}
		filters = append(filters, filter)
	}
	return filters, nil
}

// AndFilter returns true, only if all filters return true.
type AndFilter struct {
	filters []Filter
}

// Apply returns false if any of the filters returns false. Short circuited.
func (f *AndFilter) Apply(is finc.IntermediateSchema) bool {
	for _, f := range f.filters {
		if !f.Apply(is) {
			return false
		}
	}
	return true
}

// UnmarshalJSON turns a config fragment into an or filter.
func (f *AndFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Filters []json.RawMessage `json:"and"`
	}
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	var err error
	f.filters, err = unmarshalFilterList(s.Filters)
	return err
}

// NotFilter inverts another filter.
type NotFilter struct {
	filter Filter
}

// NotFilter inverts.
func (f *NotFilter) Apply(is finc.IntermediateSchema) bool {
	return !f.filter.Apply(is)
}

// UnmarshalJSON turns a config fragment into a not filter.
func (f *NotFilter) UnmarshalJSON(p []byte) error {
	var s struct {
		Filter json.RawMessage `json:"not"`
	}
	var err error
	if err := json.Unmarshal(p, &s); err != nil {
		return err
	}
	// TODO(miku): not should only work with a single filter, what is the
	// meaning of "not": [..., ..., ...] ...
	filters, err := unmarshalFilterList([]json.RawMessage{s.Filter})
	if err != nil {
		return err
	}
	if len(filters) == 0 {
		return fmt.Errorf("no filter to invert")
	}
	f.filter = filters[0]
	return nil
}

// FilterTree allows polymorphic filters.
type FilterTree struct {
	root Filter
}

// UnmarshalJSON will decide which filter is the top level one, by peeking into
// the file.
func (f *FilterTree) UnmarshalJSON(p []byte) error {
	name, err := firstKey(p)
	if err != nil {
		return err
	}
	filter, err := unmarshalFilter(name, p)
	if err != nil {
		return err
	}
	f.root = filter
	return nil
}

// Apply just applies the root filter.
func (f *FilterTree) Apply(is finc.IntermediateSchema) bool {
	return f.root.Apply(is)
}

// Tagger is takes a list of tags (ISILs) and annotates and intermediate schema
// accroding to a number of filter, defined per label. The tagger can be loaded
// directly from JSON.
type Tagger struct {
	filtermap map[string]FilterTree
}

// Tag takes an intermediate schema and returns a labeled version of that
// schema.
func (t *Tagger) Tag(is finc.IntermediateSchema) finc.IntermediateSchema {
	var tags []string
	for tag, filter := range t.filtermap {
		if filter.Apply(is) {
			tags = append(tags, tag)
		}
	}
	is.Labels = tags
	return is
}

// UnmarshalJSON unmarshals a complete filter config from serialized JSON.
func (t *Tagger) UnmarshalJSON(p []byte) error {
	var fm = make(map[string]FilterTree)
	if err := json.Unmarshal(p, &fm); err != nil {
		return err
	}
	t.filtermap = fm
	return nil
}
