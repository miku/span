package solrutil

import "fmt"

// AllowedKeys returns an error if facets values contain non-zero values that
// are not explicitly allowed. Used for reviews.
func (f FacetMap) AllowedKeys(allowed ...string) error {
	s := make(map[string]bool)
	for _, v := range allowed {
		s[v] = true
	}
	for k, count := range f {
		if !s[k] && count > 0 {
			return fmt.Errorf("facet value not allowed: %s (%d)", k, count)
		}
	}
	return nil
}

// EqualSizeNonZero checks if frequencies of the given keys are the same and
// non-zero. Used for reviews.
func (f FacetMap) EqualSizeNonZero(keys ...string) error {
	var prev int
	for i, k := range keys {
		size, ok := f[k]
		if !ok {
			return fmt.Errorf("facet key not found: %s", k)
		}
		if i > 0 {
			if prev != size {
				return fmt.Errorf("facet counts differ: %d vs %d", prev, size)
			}
		}
		prev = size
	}
	return nil
}

// AllowedKeys checks for a query and facet field, whether the values contain
// only allowed values. Used for reviews.
func (index Index) AllowedKeys(query, field string, values ...string) error {
	facets, err := index.facets(query, field)
	if err != nil {
		return err
	}
	err = facets.AllowedKeys(values...)
	if err != nil {
		return fmt.Errorf("%s [%s]: %s", query, field, err)
	}
	return nil
}

// EqualSizeNonZero checks, if given facet field values have the same size.
// Used for reviews.
func (index Index) EqualSizeNonZero(query, field string, values ...string) error {
	facets, err := index.facets(query, field)
	if err != nil {
		return err
	}
	err = facets.EqualSizeNonZero(values...)
	if err != nil {
		return fmt.Errorf("%s [%s]: %s", query, field, err)
	}
	return nil
}

// EqualSizeTotal checks, if given facet field values have the same size as the
// total number of records. Used for reviews.
func (index Index) EqualSizeTotal(query, field string, values ...string) error {
	r, err := index.FacetQuery(query, field)
	if err != nil {
		return err
	}
	total := r.Response.NumFound
	facets, err := r.Facets()
	if err != nil {
		return err
	}
	err = facets.EqualSizeNonZero(values...)
	if err != nil {
		return fmt.Errorf("%s [%s]: %s", query, field, err)
	}
	if len(values) > 0 {
		if int64(facets[values[0]]) != total {
			return fmt.Errorf("%s [%s]: size mismatch, got %d, want %d",
				query, field, facets[values[0]], total)
		}
	}
	return nil
}

// MinRatioPct fails, if the number of records matching a value undercuts a
// given ratio of all records matching the query. The ratio ranges from 0 to
// 100. Used for reviews.
func (index Index) MinRatioPct(query, field, value string, minRatioPct float64) error {
	r, err := index.FacetQuery(query, field)
	if err != nil {
		return err
	}
	total := r.Response.NumFound
	facets, err := r.Facets()
	if err != nil {
		return err
	}
	size, ok := facets[value]
	if !ok {
		return fmt.Errorf("field not found: %s", field)
	}
	ratio := (float64(size) / float64(total)) * 100
	if ratio < minRatioPct {
		return fmt.Errorf("%s [%s=%s]: ratio undercut, got %0.2f%%, want %0.2f%%",
			query, field, value, ratio, minRatioPct)
	}
	return nil
}

// MinCount fails, if the number of records matching a value undercuts a given
// size. Used for reviews.
func (index Index) MinCount(query, field, value string, minCount int) error {
	facets, err := index.facets(query, field)
	if err != nil {
		return err
	}
	size, ok := facets[value]
	if !ok {
		return fmt.Errorf("field not found: %s", field)
	}
	if size < minCount {
		return fmt.Errorf("%s [%s=%s]: undercut, got %d, want at least %d",
			query, field, value, size, minCount)
	}
	return nil
}
