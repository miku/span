package crossrefcmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

var bNewline = []byte("\n")

// Doer abstracts https://pkg.go.dev/net/http#Client.Do.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Sync saves messages from crossref.
type Sync struct {
	ApiEndpoint string
	ApiFilter   string
	ApiEmail    string
	Rows        int
	UserAgent   string
	Client      Doer
	Verbose     bool
	Mode        string
	MaxRetries  int
}

// WorksResponse, stripped of the actual messages, as we only need the status
// and mayby total results.
type WorksResponse struct {
	Message struct {
		Facets struct {
		} `json:"facets"`
		Items        []json.RawMessage `json:"items"`
		ItemsPerPage int64             `json:"items-per-page"`
		NextCursor   string            `json:"next-cursor"` // iterate
		Query        struct {
			SearchTerms any   `json:"search-terms"`
			StartIndex  int64 `json:"start-index"`
		} `json:"query"`
		TotalResults int64 `json:"total-results"` // want to estimate total results (and verify download)
	} `json:"message"`
	MessageType    string `json:"message-type"`
	MessageVersion string `json:"message-version"`
	Status         string `json:"status"`
}

// WriteWindow writes a slice of data from the API to a writer. The dates in
// filters should always be of the form YYYY-MM-DD, YYYY-MM or YYYY. The date
// filters are inclusive
// (https://api.crossref.org/swagger-ui/index.html#/operations/Works/get_works).
func (s *Sync) WriteWindow(w io.Writer, f, u time.Time) error {
	filter := fmt.Sprintf("from-%s-date:%s,until-%s-date:%s",
		s.ApiFilter, f.Format("2006-01-02"), s.ApiFilter, u.Format("2006-01-02"))
	vs := url.Values{}
	vs.Add("filter", filter)
	vs.Add("cursor", "*")
	vs.Add("rows", fmt.Sprintf("%d", s.Rows))
	if s.ApiEmail != "" {
		vs.Add("mailto", s.ApiEmail)
	}
	var (
		seen int64
		i    int
	)
OUTER:
	for {
		link := fmt.Sprintf("%s?%s", s.ApiEndpoint, vs.Encode())
		if s.Verbose {
			log.Println(link)
		}
		req, err := http.NewRequest("GET", link, nil)
		if err != nil {
			return err
		}
		req.Header.Add("User-Agent", s.UserAgent)
		resp, err := s.Client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		var wr WorksResponse
		if err := json.NewDecoder(resp.Body).Decode(&wr); err != nil {
			if i < s.MaxRetries {
				i++
				log.Printf("decode: %v", err)
				log.Printf("[%d] retrying", i)
				continue
			} else {
				// total: 10493829, seen: 3120000 (29.73%)
				// 2021/12/14 18:15:08 decode: unexpected EOF
				return fmt.Errorf("decode: %w", err)
			}
		}
		if wr.Status != "ok" {
			return fmt.Errorf("crossref api failed: %s", wr.Status)
		}
		seen = seen + int64(len(wr.Message.Items))
		if s.Verbose {
			var pct float64
			if wr.Message.TotalResults == 0 {
				pct = 0.0
			} else {
				pct = 100 * (float64(seen) / float64(wr.Message.TotalResults))
			}
			log.Printf("status: %s, total: %d, seen: %d (%0.2f%%), cursor: %s",
				wr.Status, wr.Message.TotalResults, seen, pct, wr.Message.NextCursor)
		}
		switch s.Mode {
		case "t", "tabs":
			if _, err := fmt.Fprintf(w, "%s\t%d\t%d\n",
				f.Format("2006-01-02"), seen, wr.Message.TotalResults); err != nil {
				return err
			}
			break OUTER
		case "s", "sync":
			for _, item := range wr.Message.Items {
				cleanItem := cleanEscapedSlashes(item)
				cleanItem = append(cleanItem, bNewline...)
				if _, err := w.Write(cleanItem); err != nil {
					return err
				}
			}
			if seen >= wr.Message.TotalResults {
				if s.Verbose {
					log.Printf("done, seen: %d, total: %d", seen, wr.Message.TotalResults)
				}
				return nil
			}
			cursor := wr.Message.NextCursor
			if cursor == "" {
				return nil
			}
			vs = url.Values{}
			vs.Add("cursor", cursor)
			if s.ApiEmail != "" {
				vs.Add("mailto", s.ApiEmail)
			}
		default:
			return fmt.Errorf("use tabs (t) or sync (s) mode")
		}
		// status: ok, total: 55818, seen: 47818 (85.67%)
		// We had repeated requests, with seemingly a new cursor, but no new
		// messages and seen < total; we assume, we have got all we could and
		// move on. Note: this may be a temporary glitch; rather retry.
		if len(wr.Message.Items) == 0 {
			if wr.Message.TotalResults-seen < int64(0.1*float64(wr.Message.TotalResults)) {
				log.Printf("assuming ok to skip - seen: %d, total: %d", seen, wr.Message.TotalResults)
				break
			} else {
				return fmt.Errorf("no more messages, consider restart; total: %d, seen: %d", wr.Message.TotalResults, seen)
			}
		}
		i = 0
	}
	return nil
}

// cleanEscapedSlashes is a convenience; API returns escaped forward slashes,
// which is valid JSON but makes grepping harder.
func cleanEscapedSlashes(raw json.RawMessage) json.RawMessage {
	if !json.Valid(raw) {
		return raw
	}
	fixed := bytes.ReplaceAll(raw, []byte(`\/`), []byte(`/`))
	if !json.Valid(fixed) {
		return raw
	}
	return fixed
}
