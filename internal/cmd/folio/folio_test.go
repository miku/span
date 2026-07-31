package folio

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/miku/span/folio"
)

// stubClient is a fake FOLIO client for testing Run without network access.
type stubClient struct {
	authErr    error
	authCalled bool
	gotUser    string
	gotPass    string
	resp       *folio.MetadataCollectionsResponse
	respErr    error
}

func (s *stubClient) Authenticate(username, password string) error {
	s.authCalled = true
	s.gotUser = username
	s.gotPass = password
	return s.authErr
}

func (s *stubClient) MetadataCollections(opts folio.MetadataCollectionsOpts) (*folio.MetadataCollectionsResponse, error) {
	return s.resp, s.respErr
}

func sampleResponse() *folio.MetadataCollectionsResponse {
	return &folio.MetadataCollectionsResponse{
		FincConfigMetadataCollections: []folio.FincConfigMetadataCollection{
			{
				Label:               "Example Collection",
				SolrMegaCollections: []string{"Example MegaCollection"},
			},
		},
	}
}

func TestRunIncompleteCredentials(t *testing.T) {
	c := &stubClient{}
	err := Run(Config{User: "", Password: ""}, c, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for incomplete credentials")
	}
	if c.authCalled {
		t.Error("Authenticate should not be called with incomplete credentials")
	}
}

func TestRunAuthError(t *testing.T) {
	c := &stubClient{authErr: errors.New("boom")}
	err := Run(Config{User: "u", Password: "p"}, c, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected auth error to propagate")
	}
	if c.gotUser != "u" || c.gotPass != "p" {
		t.Errorf("credentials not forwarded: got %q/%q", c.gotUser, c.gotPass)
	}
}

func TestRunTableOutput(t *testing.T) {
	c := &stubClient{resp: sampleResponse()}
	var buf bytes.Buffer
	if err := Run(Config{User: "u", Password: "p"}, c, &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Example Collection") {
		t.Errorf("table output missing label: %q", out)
	}
	if !strings.Contains(out, "Example MegaCollection") {
		t.Errorf("table output missing mega collection: %q", out)
	}
}

func TestRunRawOutput(t *testing.T) {
	c := &stubClient{resp: sampleResponse()}
	var buf bytes.Buffer
	if err := Run(Config{User: "u", Password: "p", Raw: true}, c, &buf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	// Raw output is one JSON object per line.
	if !strings.HasPrefix(out, "{") || !strings.Contains(out, `"label":"Example Collection"`) {
		t.Errorf("raw output not JSON as expected: %q", out)
	}
}
