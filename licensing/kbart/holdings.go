// Package kbart implements support for KBART (Knowledge Bases And Related Tools
// working group, http://www.uksg.org/kbart/) holding files
// (http://www.uksg.org/kbart/s5/guidelines/data_format).
//
// > This is a generic format that minimizes the effort involved in receiving and
// loading the data, and reduces the likelihood of errors being introduced during
// exchange. Tab-delimited formats are preferable to comma-separated formats, as
// commas appear regularly within the distributed data and, though they can be
// "commented out", doing so leaves a greater opportunity for error than the use
// of a tab-delimited format. Tab-delimited formats can be easily exported from
// all commonly used spreadsheet programs.
package kbart

import (
	"io"
	"regexp"

	"github.com/miku/span"
	"github.com/miku/span/encoding/tsv"
	"github.com/miku/span/licensing"
)

// Holdings contains a list of entries about licenced or available content. In
// addition to access to all entries, this type exposes a couple of helper
// methods.
type Holdings []licensing.Entry

// ReadFrom create holdings struct from a reader. Expects a tab separated CSV with
// a single header line.
func (h *Holdings) ReadFrom(r io.Reader) (int64, error) {
	var wc span.WriteCounter
	dec := tsv.NewDecoder(io.TeeReader(r, &wc))
	for {
		var entry licensing.Entry
		err := dec.Decode(&entry)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		*h = append(*h, entry)
	}
	return int64(wc.Count()), nil
}

// SerialNumberMap creates a map from ISSN to associated licensing entries. This
// is here for performance mostly, so we can access relevant licensing entry by ISSN.
func (h *Holdings) SerialNumberMap() map[string][]licensing.Entry {
	cache := make(map[string]map[licensing.Entry]bool)
	for _, e := range *h {
		for _, issn := range e.ISSNList() {
			if cache[issn] == nil {
				cache[issn] = make(map[licensing.Entry]bool)
			}
			cache[issn][e] = true
		}
	}
	// Make unique.
	result := make(map[string][]licensing.Entry)
	for issn, entrymap := range cache {
		for k := range entrymap {
			result[issn] = append(result[issn], k)
		}
	}
	return result
}

// TitleMap maps an exact title to a list of entries.
func (h *Holdings) TitleMap() map[string][]licensing.Entry {
	cache := make(map[string]map[licensing.Entry]bool)
	for _, e := range *h {
		if cache[e.PublicationTitle] == nil {
			cache[e.PublicationTitle] = make(map[licensing.Entry]bool)
		}
		cache[e.PublicationTitle][e] = true
	}
	// Make unique.
	result := make(map[string][]licensing.Entry)
	for title, entrymap := range cache {
		for k := range entrymap {
			result[title] = append(result[title], k)
		}
	}
	return result
}

// WisoDatabaseMap derives a structure from the holdings file, that maps WISO
// database names to the associated entries, refs. #9534.
func (h *Holdings) WisoDatabaseMap() map[string][]licensing.Entry {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`https://www.wiso-net.de/toc_list/([A-Z]{3,4})`),
		regexp.MustCompile(`https://www.wiso-net.de/.*dbShortcut=:2:2:([A-Z]{3,4})`),
		regexp.MustCompile(`https://www.wiso-net.de/.*dbShortcut=([A-Z]{3,4})`),
	}
	cache := make(map[string]map[licensing.Entry]bool)
	for _, e := range *h {
		for _, p := range patterns {
			matches := p.FindStringSubmatch(e.TitleURL)
			if len(matches) < 2 {
				continue
			}
			db := matches[1]
			if cache[db] == nil {
				cache[db] = make(map[licensing.Entry]bool)
			}
			cache[db][e] = true
		}
	}
	// Make unique.
	result := make(map[string][]licensing.Entry)
	for issn, entrymap := range cache {
		for k := range entrymap {
			result[issn] = append(result[issn], k)
		}
	}
	return result
}

// Filter finds entries with certain characteristics. This will be slow for KBART
// files with thousands of entries.
func (h *Holdings) Filter(f licensing.FilterFunc) (result []licensing.Entry) {
	cache := make(map[licensing.Entry]bool)
	for _, e := range *h {
		if f(e) {
			cache[e] = true
		}
	}
	for k := range cache {
		result = append(result, k)
	}
	return
}
