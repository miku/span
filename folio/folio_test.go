package folio

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasic(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("body"))
	}))
	defer func() {
		ts.Close()
	}()
	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatalf("req: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if resp.StatusCode != 202 {
		t.Fatalf("got %v, want %v", resp.StatusCode, 202)
	}
}
