package crossrefcmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/miku/span/internal/atomicfile"
)

var bNewline = []byte("\n")

var (
	// errRetry marks errors after which a harvest can resume.
	errRetry = errors.New("retry")
	// errIncomplete means the cursor ended before total results were seen.
	errIncomplete = errors.New("cursor exhausted early")
)

// sortFields maps date filters to the record field they refer to. We sort by
// this field, so an interrupted harvest can continue after the last record.
var sortFields = map[string]string{"index": "indexed", "deposit": "deposited", "created": "created"}

// Doer abstracts https://pkg.go.dev/net/http#Client.Do.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Compressor wraps a writer, e.g. with gzip or zstd.
type Compressor func(io.Writer) (io.WriteCloser, error)

// Sync saves messages from crossref.
type Sync struct {
	ApiEndpoint string
	ApiFilter   string
	ApiEmail    string
	Rows        int
	UserAgent   string
	Client      Doer
	Verbose     bool
	MaxResumes  int                     // consecutive failures without progress
	Backoff     func(int) time.Duration // wait before resuming; nil for default
}

// WorksResponse, stripped of the actual messages, as we only need the status
// and mayby total results.
type WorksResponse struct {
	Message struct {
		Items        []json.RawMessage `json:"items"`
		NextCursor   string            `json:"next-cursor"`
		TotalResults int64             `json:"total-results"`
	} `json:"message"`
	Status string `json:"status"`
}

// checkpoint records harvest progress: the partial file is complete up to
// Offset and ends with the records DOIs, which share the sort field
// Timestamp (ms).
type checkpoint struct {
	Offset    int64    `json:"offset"`
	Timestamp int64    `json:"timestamp"`
	DOIs      []string `json:"dois"`
}

// Count returns the number of works in a window.
func (s *Sync) Count(f, u time.Time) (int64, error) {
	wr, err := s.fetch(url.Values{"filter": {s.filter(f.Format(time.DateOnly), u)}, "rows": {"0"}})
	if err != nil {
		return 0, err
	}
	return wr.Message.TotalResults, nil
}

// HarvestWindow writes all works of a window to a compressed file at path.
// The dates are inclusive. Records are requested in order of the filter
// field; each page is appended to path.partial as a separate compressed
// stream, followed by a checkpoint in path.checkpoint. When requests fail,
// e.g. with HTTP 429, or the process is killed, the harvest resumes with a
// fresh cursor after the last record written. Concatenated gzip or zstd
// streams are valid files.
func (s *Sync) HarvestWindow(path string, compress Compressor, f, u time.Time) error {
	field, ok := sortFields[s.ApiFilter]
	if !ok {
		return fmt.Errorf("cannot resume with filter %q, use: index, deposit or created", s.ApiFilter)
	}
	partial, ckPath := path+".partial", path+".checkpoint"
	ck, err := loadCheckpoint(ckPath)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(partial, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if fi, err := file.Stat(); err != nil {
		return err
	} else if fi.Size() < ck.Offset {
		log.Printf("%s shorter than checkpoint, starting over", partial)
		ck = checkpoint{}
	}
	// Drop anything written after the last checkpoint.
	if err := file.Truncate(ck.Offset); err != nil {
		return err
	}
	if _, err := file.Seek(ck.Offset, io.SeekStart); err != nil {
		return err
	}
	if ck.Offset > 0 {
		log.Printf("resuming %s at %s", partial, time.UnixMilli(ck.Timestamp).UTC())
	}
	for fails := 0; ; {
		n, err := s.pass(file, compress, &ck, ckPath, field, f, u)
		if err == nil {
			break
		}
		if !errors.Is(err, errRetry) {
			return err
		}
		if n == 0 && ck.Timestamp > 0 && errors.Is(err, errIncomplete) {
			// A fresh query from the checkpoint yields nothing new.
			log.Printf("accepting %s as complete: %v", partial, err)
			break
		}
		if n > 0 {
			fails = 0
		}
		if fails++; fails > s.MaxResumes {
			return err
		}
		wait := s.backoff(fails)
		log.Printf("%v; resuming in %s [%d/%d]", err, wait, fails, s.MaxResumes)
		time.Sleep(wait)
	}
	if ck.Offset == 0 {
		// An empty window, still a valid compressed file.
		zw, err := compress(file)
		if err != nil {
			return err
		}
		if err := zw.Close(); err != nil {
			return err
		}
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(partial, path); err != nil {
		return err
	}
	if err := os.Remove(ckPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// pass pages through the window with a single cursor, starting at the
// checkpoint, and returns the number of records written.
func (s *Sync) pass(file *os.File, compress Compressor, ck *checkpoint, ckPath, field string, f, u time.Time) (n int, err error) {
	from := f.Format(time.DateOnly)
	if ck.Timestamp > 0 {
		// Filter resolution is one second, so we get a few records again.
		from = time.UnixMilli(ck.Timestamp).UTC().Format("2006-01-02T15:04:05")
	}
	// The API forgets the query, if we only send the cursor.
	vs := url.Values{
		"filter": {s.filter(from, u)},
		"sort":   {field},
		"order":  {"asc"},
		"rows":   {strconv.Itoa(s.Rows)},
		"cursor": {"*"},
	}
	var seen int64
	for {
		wr, err := s.fetch(vs)
		if err != nil {
			return n, fmt.Errorf("%w: %w", errRetry, err)
		}
		items, total := wr.Message.Items, wr.Message.TotalResults
		seen += int64(len(items))
		k, err := appendPage(file, compress, ck, ckPath, field, items)
		n += k
		if err != nil {
			return n, err
		}
		if s.Verbose {
			log.Printf("from: %s, total: %d, seen: %d, written: %d", from, total, seen, n)
		}
		if len(items) == 0 || wr.Message.NextCursor == "" || seen >= total {
			if seen < total {
				return n, fmt.Errorf("%w: %w: seen %d of %d", errRetry, errIncomplete, seen, total)
			}
			return n, nil
		}
		vs.Set("cursor", wr.Message.NextCursor)
	}
}

// appendPage writes the records of a page not yet written as one compressed
// stream, then saves a checkpoint.
func appendPage(file *os.File, compress Compressor, ck *checkpoint, ckPath, field string, items []json.RawMessage) (int, error) {
	var (
		buf  bytes.Buffer
		next = *ck
		n    int
	)
	for _, item := range items {
		doi, ts, err := recordKey(item, field)
		if err != nil {
			return 0, err
		}
		if ts < next.Timestamp || ts == next.Timestamp && slices.Contains(next.DOIs, doi) {
			continue
		}
		if ts > next.Timestamp {
			next.Timestamp, next.DOIs = ts, nil
		}
		next.DOIs = append(next.DOIs, doi)
		buf.Write(cleanEscapedSlashes(item))
		buf.Write(bNewline)
		n++
	}
	if n == 0 {
		return 0, nil
	}
	zw, err := compress(file)
	if err != nil {
		return 0, err
	}
	if _, err := zw.Write(buf.Bytes()); err != nil {
		return 0, err
	}
	if err := zw.Close(); err != nil {
		return 0, err
	}
	if err := file.Sync(); err != nil {
		return 0, err
	}
	if next.Offset, err = file.Seek(0, io.SeekCurrent); err != nil {
		return 0, err
	}
	if err := next.save(ckPath); err != nil {
		return 0, err
	}
	*ck = next
	return n, nil
}

// recordKey returns the DOI of a record and the timestamp (ms) of a date field.
func recordKey(item json.RawMessage, field string) (doi string, ts int64, err error) {
	type date struct {
		Timestamp int64 `json:"timestamp"`
	}
	var doc struct {
		DOI       string `json:"DOI"`
		Indexed   date   `json:"indexed"`
		Deposited date   `json:"deposited"`
		Created   date   `json:"created"`
	}
	if err := json.Unmarshal(item, &doc); err != nil {
		return "", 0, err
	}
	ts = map[string]int64{
		"indexed":   doc.Indexed.Timestamp,
		"deposited": doc.Deposited.Timestamp,
		"created":   doc.Created.Timestamp,
	}[field]
	if ts == 0 {
		return "", 0, fmt.Errorf("record %s: missing %s timestamp", doc.DOI, field)
	}
	return doc.DOI, ts, nil
}

func (s *Sync) filter(from string, u time.Time) string {
	return fmt.Sprintf("from-%s-date:%s,until-%s-date:%s", s.ApiFilter, from, s.ApiFilter, u.Format(time.DateOnly))
}

func (s *Sync) backoff(fails int) time.Duration {
	if s.Backoff != nil {
		return s.Backoff(fails)
	}
	return min(time.Minute<<min(fails-1, 5), 30*time.Minute)
}

// fetch performs a single works API request.
func (s *Sync) fetch(vs url.Values) (*WorksResponse, error) {
	if s.ApiEmail != "" {
		vs.Set("mailto", s.ApiEmail)
	}
	link := s.ApiEndpoint + "?" + vs.Encode()
	if s.Verbose {
		log.Println(link)
	}
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.UserAgent)
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var wr WorksResponse
	if err := json.NewDecoder(resp.Body).Decode(&wr); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if wr.Status != "ok" {
		return nil, fmt.Errorf("crossref api failed: %s", wr.Status)
	}
	return &wr, nil
}

func loadCheckpoint(path string) (checkpoint, error) {
	var ck checkpoint
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return ck, nil
	}
	if err != nil {
		return ck, err
	}
	return ck, json.Unmarshal(b, &ck)
}

func (ck checkpoint) save(path string) error {
	f, err := atomicfile.Create(path, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(ck); err != nil {
		return err
	}
	return f.Commit()
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
