package compact

import (
	"bytes"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	var cases = []struct {
		cfg     Config
		wantErr bool
	}{
		{Config{Strategy: "first"}, false},
		{Config{Strategy: "last"}, false},
		{Config{Strategy: "random"}, false},
		{Config{Strategy: "min", SortField: "x"}, false},
		{Config{Strategy: "max", SortField: "x"}, false},
		{Config{Strategy: "min"}, true},  // missing sort-key
		{Config{Strategy: "max"}, true},  // missing sort-key
		{Config{Strategy: "nope"}, true}, // unknown
	}
	for _, c := range cases {
		err := c.cfg.Validate()
		if (err != nil) != c.wantErr {
			t.Errorf("Validate(%+v) err=%v, wantErr=%v", c.cfg, err, c.wantErr)
		}
	}
}

func TestExtract(t *testing.T) {
	c := Config{KeyField: "id", SortField: "ts"}
	key, sortVal, err := c.extract([]byte(`{"id":"abc","ts":"2020"}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(key) != `"abc"` {
		t.Errorf("key = %q, want \"abc\"", key)
	}
	if sortVal != "2020" {
		t.Errorf("sortVal = %q, want 2020", sortVal)
	}
	if _, _, err := c.extract([]byte(`{"other":1}`)); err == nil {
		t.Error("expected missing key error")
	}
}

func TestRawToSortString(t *testing.T) {
	if got := rawToSortString([]byte(`"hello"`)); got != "hello" {
		t.Errorf("got %q, want hello", got)
	}
	if got := rawToSortString([]byte(`42`)); got != "42" {
		t.Errorf("got %q, want 42", got)
	}
}

func TestRunLastStrategy(t *testing.T) {
	input := strings.Join([]string{
		`{"id":"a","v":1}`,
		`{"id":"a","v":2}`,
		`{"id":"b","v":1}`,
	}, "\n") + "\n"
	var buf bytes.Buffer
	cfg := DefaultConfig() // strategy "last", key "id"
	if err := Run(cfg, strings.NewReader(input), &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %q", len(lines), buf.String())
	}
	// "last" keeps the last record per key; groups are emitted sorted by key.
	if lines[0] != `{"id":"a","v":2}` {
		t.Errorf("line 0 = %q, want {\"id\":\"a\",\"v\":2}", lines[0])
	}
	if lines[1] != `{"id":"b","v":1}` {
		t.Errorf("line 1 = %q, want {\"id\":\"b\",\"v\":1}", lines[1])
	}
}

func TestRunFirstStrategy(t *testing.T) {
	input := strings.Join([]string{
		`{"id":"a","v":1}`,
		`{"id":"a","v":2}`,
	}, "\n") + "\n"
	var buf bytes.Buffer
	cfg := DefaultConfig()
	cfg.Strategy = "first"
	if err := Run(cfg, strings.NewReader(input), &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != `{"id":"a","v":1}` {
		t.Errorf("first: got %q, want {\"id\":\"a\",\"v\":1}", got)
	}
}
