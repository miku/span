package reviewutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/miku/span"
	"github.com/sethgrid/pester"
	yaml "gopkg.in/yaml.v2"
)

// ReviewConfig contains various index review cases and general configuration.
type ReviewConfig struct {
	SolrServer        string     `yaml:"solr"`
	Ticket            string     `yaml:"ticket"`
	ZeroResultsPolicy string     `yaml:"zero-results-policy"`
	AllowedKeys       [][]string `yaml:"allowed-keys"`
	AllRecords        [][]string `yaml:"all-records"`
	MinRatio          [][]string `yaml:"min-ratio"`
	MinCount          [][]string `yaml:"min-count"`
}

// ReadFrom can populate a config from a YAML stream.
func (rc *ReviewConfig) ReadFrom(r io.Reader) (n int64, err error) {
	if rc == nil {
		rc = &ReviewConfig{}
	}
	readerCounter := span.NewReaderCounter(r)
	if err = yaml.NewDecoder(readerCounter).Decode(rc); err != nil {
		return 0, err
	}
	return readerCounter.Count(), nil
}

type Redmine struct {
	BaseURL string
	Token   string
}

// TicketLink returns link to issue for updates.
func (r *Redmine) TicketLink(ticket string) string {
	return fmt.Sprintf("%s/issues/%s.json", r.BaseURL, ticket)
}

// UpdateTicket updates ticket given ticket number and message.
// http://www.redmine.org/projects/redmine/wiki/Rest_Issues#Updating-an-issue
func (r *Redmine) UpdateTicket(ticket, message string) error {
	// Prepare payload.
	body, err := json.Marshal(map[string]interface{}{
		"issue": map[string]interface{}{
			"notes": message,
		},
	})
	if err != nil {
		return err
	}
	link := r.TicketLink(ticket)
	log.Printf("prepare to PUT to %s", link)
	req, err := http.NewRequest("PUT", link, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Redmine-API-Key", r.Token)
	resp, err := pester.Do(req)
	if err != nil {
		return fmt.Errorf("could not update ticket: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error, ticket update resulted in a %d", resp.StatusCode)
	}
	// We expect a zero byte response from redmine for success.
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	switch len(b) {
	case 0:
		log.Printf("got empty response from redmine [%d]", resp.StatusCode)
	default:
		log.Printf("redmine response [%d] (%db): %s", resp.StatusCode, len(b), string(b))
	}
	log.Printf("updated %s/issues/%s", r.BaseURL, ticket)
	return nil
}
