package folio

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bl-users/login" {
			t.Fatalf("unexpected URL: %s", r.URL.Path)
		}
		b, err := httputil.DumpRequest(r, true)
		if err != nil {
			t.Fatalf("cannot dump request")
		}
		t.Logf("server got request: %v", string(b))
		w.Header().Add("X-OKAPI-TOKEN", "12345678")
	}))
	t.Logf("test server: %v", ts.URL)
	defer func() {
		ts.Close()
	}()
	api := &API{
		Base: ts.URL,
	}
	if api.Token != "" {
		t.Fatalf("expected empty token, got %v", api.Token)
	}
	err := api.Authenticate("admin", "admin")
	if err != nil {
		t.Fatalf("failed with: %v", err)
	}
}
