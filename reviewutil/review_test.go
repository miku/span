package reviewutil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateTicket(t *testing.T) {
	// Dummy server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We only test this issue, method, headers, ...
		if r.URL.String() == "/issues/1234.json" {
			if r.Method != "PUT" {
				t.Errorf("expected PUT for ticket update, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Content-Type: want application/json, got %s", r.Header.Get("Content-Type"))
			}
			if r.Header.Get("X-Redmine-API-Key") != "xyz" {
				t.Errorf("X-Redmine-API-Key: want xyz got %s", r.Header.Get("X-Redmine-API-Key"))
			}
		}
	}))
	defer ts.Close()

	redmine := Redmine{BaseURL: ts.URL, Token: "xyz"}
	if err := redmine.UpdateTicket("1234", "Some Message"); err != nil {
		t.Errorf("UpdateTicket got %v, want nil", err)
	}
}
