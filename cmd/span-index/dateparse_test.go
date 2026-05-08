package main

import (
	"testing"
	"time"
)

func TestParseDateRelative(t *testing.T) {
	cases := []struct {
		in   string
		near time.Duration // tolerance from "now"
		back time.Duration // expected offset back from now
	}{
		{"now", time.Second, 0},
		{"1.day.ago", time.Second, 24 * time.Hour},
		{"2.days.ago", time.Second, 48 * time.Hour},
		{"3.hours.ago", time.Second, 3 * time.Hour},
		{"1 week ago", time.Second, 7 * 24 * time.Hour},
	}
	for _, c := range cases {
		got, err := parseDate(c.in)
		if err != nil {
			t.Errorf("parseDate(%q): %v", c.in, err)
			continue
		}
		want := time.Now().UTC().Add(-c.back)
		if d := got.Sub(want); d > c.near || d < -c.near {
			t.Errorf("parseDate(%q) = %v, want ~%v (off by %v)", c.in, got, want, d)
		}
	}
}

func TestParseDateISO(t *testing.T) {
	got, err := parseDate("2026-01-01")
	if err != nil {
		t.Fatal(err)
	}
	if got.Year() != 2026 || got.Month() != 1 || got.Day() != 1 {
		t.Errorf("got %v, want 2026-01-01", got)
	}
}

func TestParseDateBad(t *testing.T) {
	if _, err := parseDate("not-a-date"); err == nil {
		t.Error("expected error for garbage input")
	}
	if _, err := parseDate(""); err == nil {
		t.Error("expected error for empty input")
	}
}
