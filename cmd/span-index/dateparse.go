package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

// parseDate accepts ISO timestamps and a small set of git-like relative forms.
//
// Accepted relative forms (case-insensitive):
//
//	now, today, yesterday
//	N.unit.ago             units: hour(s), day(s), week(s), month(s), year(s)
//	N units ago            (spaces work too: "2 weeks ago")
//
// Anything else is delegated to araddon/dateparse, which understands ISO 8601
// and most common date layouts. The returned time is in UTC.
func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	now := time.Now().UTC()
	switch strings.ToLower(s) {
	case "now":
		return now, nil
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	case "yesterday":
		return time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC), nil
	}
	if t, ok := parseRelative(s, now); ok {
		return t, nil
	}
	t, err := dateparse.ParseAny(s)
	if err != nil {
		return time.Time{}, fmt.Errorf("unrecognised date %q: %w", s, err)
	}
	return t.UTC(), nil
}

// parseRelative recognises N.unit.ago and "N units ago" forms. Returns ok=false
// when the input is shaped differently so the caller can fall back.
func parseRelative(s string, now time.Time) (time.Time, bool) {
	lower := strings.ToLower(s)
	parts := strings.FieldsFunc(lower, func(r rune) bool {
		return r == '.' || r == ' '
	})
	if len(parts) != 3 || parts[2] != "ago" {
		return time.Time{}, false
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil || n < 0 {
		return time.Time{}, false
	}
	unit := strings.TrimSuffix(parts[1], "s")
	switch unit {
	case "second":
		return now.Add(-time.Duration(n) * time.Second), true
	case "minute":
		return now.Add(-time.Duration(n) * time.Minute), true
	case "hour":
		return now.Add(-time.Duration(n) * time.Hour), true
	case "day":
		return now.AddDate(0, 0, -n), true
	case "week":
		return now.AddDate(0, 0, -7*n), true
	case "month":
		return now.AddDate(0, -n, 0), true
	case "year":
		return now.AddDate(-n, 0, 0), true
	}
	return time.Time{}, false
}

// formatSolrDate renders t as a Solr-friendly UTC timestamp.
func formatSolrDate(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}
