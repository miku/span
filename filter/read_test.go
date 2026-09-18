package filter

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestReadLines(t *testing.T) {
	var cases = []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a\n\n  b \r\n", []string{"a", "b"}},
		{"a\nb", []string{"a"}}, // a last line without newline is dropped
	}
	for _, c := range cases {
		got, err := readLines(strings.NewReader(c.in))
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("readLines(%q): got %q, want %q", c.in, got, c.want)
		}
	}
}

func TestUnzipOrPlain(t *testing.T) {
	zipped, err := os.ReadFile("testdata/z.zip")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ in, want string }{
		{string(zipped), "a\nb\n"},
		{"plain\n", "plain\n"},
		{"", ""},
	} {
		got, err := unzipOrPlain([]byte(c.in))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
	}
}

func TestReadLinkOrFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("1234-5678\n2345-6789\n"))
	}))
	defer srv.Close()
	file := filepath.Join(t.TempDir(), "list.txt")
	if err := os.WriteFile(file, []byte("0000-0000\n"), 0644); err != nil {
		t.Fatal(err)
	}
	var cases = []struct {
		about      string
		link, file string
		want       []string
		err        bool
	}{
		{"nothing", "", "", nil, false},
		{"link", srv.URL + "/list", "", []string{"1234-5678", "2345-6789"}, false},
		{"link wins", srv.URL + "/list", file, []string{"1234-5678", "2345-6789"}, false},
		{"file", "", file, []string{"0000-0000"}, false},
		{"file url", "file://" + file, "", []string{"0000-0000"}, false},
		{"not found", srv.URL + "/missing", "", nil, true},
	}
	for _, c := range cases {
		t.Run(c.about, func(t *testing.T) {
			got, err := readLinkOrFile(c.link, c.file)
			if (err != nil) != c.err {
				t.Fatalf("got err %v, want error: %v", err, c.err)
			}
			if !slices.Equal(got, c.want) {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestHoldingsCachePutLink(t *testing.T) {
	zipped, err := os.ReadFile("testdata/z.zip")
	if err != nil {
		t.Fatal(err)
	}
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		switch r.URL.Path {
		case "/kbart.txt":
			w.Write([]byte("publication_title\tprint_identifier\nA\t1234-5678\n"))
		case "/kbart.zip":
			w.Write(zipped) // not KBART, but must parse without error
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := make(HoldingsCache)
	for range 2 {
		if err := c.putLink(srv.URL + "/kbart.txt"); err != nil {
			t.Fatal(err)
		}
	}
	if requests != 1 {
		t.Errorf("got %d requests, want 1, a cached link must not be fetched again", requests)
	}
	if _, ok := c[srv.URL+"/kbart.txt"].SerialNumberMap["1234-5678"]; !ok {
		t.Errorf("expected ISSN in cache, got %v", c)
	}
	if err := c.putLink(srv.URL + "/kbart.zip"); err != nil {
		t.Fatal(err)
	}
	if err := c.putLink(srv.URL + "/missing"); err == nil {
		t.Error("expected error for missing link")
	}
}
