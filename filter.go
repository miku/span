package span

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/miku/span/container"
	"github.com/miku/span/finc"
	"github.com/miku/span/holdings"
)

// Filter wraps the decision, whether a given record should be attached or not.
type Filter interface {
	Apply(finc.IntermediateSchema) bool
}

// Any always returns true.
type Any struct{}

// Apply filter.
func (f Any) Apply(is finc.IntermediateSchema) bool { return true }

// SourceFilter allows to attach ISIL on records of a given source.
type SourceFilter struct {
	SourceID string
}

// Apply filter.
func (f SourceFilter) Apply(is finc.IntermediateSchema) bool {
	return is.SourceID == f.SourceID
}

// MarshalJSON provides custom serialization.
func (f SourceFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.SourceID)
}

// HoldingFilter decides ISIL-attachment by looking at licensing information from OVID files.
type HoldingFilter struct {
	Table holdings.Licenses
}

// NewHoldingFilter loads the holdings information for a single institution.
func NewHoldingFilter(r io.Reader) (HoldingFilter, error) {
	licenses, errs := holdings.ParseHoldings(r)
	if len(errs) > 0 {
		var msgs []string
		for i, e := range errs {
			msgs = append(msgs, fmt.Sprintf("[%d] %s", i, e.Error()))
		}
		return HoldingFilter{Table: licenses}, fmt.Errorf("one or more errors\n%s", strings.Join(msgs, "\n"))
	}
	return HoldingFilter{Table: licenses}, nil
}

// MarshalJSON provides custom serialization.
func (f HoldingFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Table)
}

// CoveredAndValid checks coverage and moving wall.
func (f HoldingFilter) CoveredAndValid(signature, issn string) bool {
	licenses, ok := f.Table[issn]
	if !ok {
		return false
	}
	now := time.Now()
	for _, license := range licenses {
		if !license.Covers(signature) {
			continue
		}
		if now.After(license.Wall()) {
			return true
		}
	}
	return false
}

// HoldingFilter compares the (year, volume, issue) of the
// record with license information, including possible moving walls.
func (f HoldingFilter) Apply(is finc.IntermediateSchema) bool {
	signature := holdings.CombineDatum(fmt.Sprintf("%d", is.Date.Year()), is.Volume, is.Issue, "")
	for _, issn := range append(is.ISSN, is.EISSN...) {
		if f.CoveredAndValid(signature, issn) {
			return true
		}
	}
	return false
}

// ListFilter will include records, whose ISSN is contained in a given set.
type ListFilter struct {
	Set *container.StringSet
}

// NewAttachByList reads one record per line from reader.
func NewListFilter(r io.Reader) (ListFilter, error) {
	br := bufio.NewReader(r)
	f := ListFilter{Set: container.NewStringSet()}
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return f, err
		}
		f.Set.Add(strings.TrimSpace(line))
	}
	return f, nil
}

// MarshalJSON provides custom serialization.
func (f ListFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Set.Values())
}

// Apply filter.
func (f ListFilter) Apply(is finc.IntermediateSchema) bool {
	for _, issn := range append(is.ISSN, is.EISSN...) {
		if f.Set.Contains(issn) {
			return true
		}
	}
	return false
}

// ISILAttacher maps an ISIL to one or more Filters. If any of these filters
// return true, the ISIL should be attached (Therefore order does not matter).
type ISILTagger map[string][]Filter

// Tags returns all ISILs that can be attached to a given intermediate schema record.
func (t ISILTagger) Tags(is finc.IntermediateSchema) []string {
	isils := container.NewStringSet()
	for isil, filters := range t {
		for _, f := range filters {
			if f.Apply(is) {
				isils.Add(isil)
			}
		}
	}
	return isils.Values()
}
