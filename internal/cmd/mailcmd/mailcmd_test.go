package mailcmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/miku/span/mail"
)

func TestValidate(t *testing.T) {
	var cases = []struct {
		cfg  Config
		want string // substring of error, "" for no error
	}{
		{Config{}, "-f/--sender"},
		{Config{From: "a@x"}, "-s/--subject"},
		{Config{From: "a@x", Subject: "s"}, "-t/--recipient"},
		{Config{From: "a@x", Subject: "s", To: []string{"b@x"}}, ""},
	}
	for i, c := range cases {
		err := c.cfg.Validate()
		if c.want == "" {
			if err != nil {
				t.Errorf("case %d: unexpected error %v", i, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("case %d: got %v, want substring %q", i, err, c.want)
		}
	}
}

func TestRunWritesMessage(t *testing.T) {
	cfg := Config{
		From:    "a@x",
		To:      []string{"b@x"},
		Subject: "hello",
		Body:    "world",
	}
	var buf bytes.Buffer
	if err := Run(cfg, &buf, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Subject: hello") {
		t.Errorf("missing subject in %q", out)
	}
	if !strings.Contains(out, "world") {
		t.Errorf("missing body in %q", out)
	}
}

func TestRunSends(t *testing.T) {
	cfg := Config{From: "a@x", To: []string{"b@x"}, Subject: "s", Body: "b"}
	var got *mail.Message
	send := func(m *mail.Message) error { got = m; return nil }
	if err := Run(cfg, nil, send); err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Subject != "s" {
		t.Errorf("sender not called correctly: %+v", got)
	}
}

func TestRunValidateFails(t *testing.T) {
	if err := Run(Config{}, &bytes.Buffer{}, nil); err == nil {
		t.Error("expected validation error")
	}
}
