package localdata

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	var cases = []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "id source doi",
			input: `{"finc.id":"ai-1","finc.source_id":"49","doi":"10.1/x"}` + "\n",
			want:  "ai-1,49,10.1/x\n",
		},
		{
			name:  "with labels",
			input: `{"finc.id":"ai-2","finc.source_id":"50","doi":"10.2/y","x.labels":["DE-15","DE-14"]}` + "\n",
			want:  "ai-2,50,10.2/y,DE-15,DE-14\n",
		},
		{
			name:  "missing fields",
			input: `{"finc.id":"ai-3"}` + "\n",
			want:  "ai-3,,\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := Config{BatchSize: 1}
			if err := Run(cfg, strings.NewReader(c.input), &buf); err != nil {
				t.Fatalf("Run: %v", err)
			}
			if got := buf.String(); got != c.want {
				t.Errorf("Run() = %q, want %q", got, c.want)
			}
		})
	}
}
