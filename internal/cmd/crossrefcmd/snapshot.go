package crossrefcmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"github.com/miku/span/formats/crossref"
	"github.com/miku/span/parallel"
	json "github.com/segmentio/encoding/json"
)

// Stage1Config holds the tunables for Stage1Extract.
type Stage1Config struct {
	BatchSize         int
	ErrCountThreshold int64
}

// DefaultStage1Config returns the default configuration.
func DefaultStage1Config() Stage1Config {
	return Stage1Config{
		BatchSize:         40000,
		ErrCountThreshold: 1,
	}
}

// WriteFields writes a variable number of values separated by sep to a given
// writer. Returns bytes written and error.
func WriteFields(w io.Writer, sep string, values ...any) (int, error) {
	var ss = make([]string, len(values))
	for i, v := range values {
		switch v.(type) {
		case int, int8, int16, int32, int64:
			ss[i] = fmt.Sprintf("%d", v)
		case uint, uint8, uint16, uint32, uint64:
			ss[i] = fmt.Sprintf("%d", v)
		case float32, float64:
			ss[i] = fmt.Sprintf("%f", v)
		case fmt.Stringer:
			ss[i] = fmt.Sprintf("%s", v)
		default:
			ss[i] = fmt.Sprintf("%v", v)
		}
	}
	s := fmt.Sprintln(strings.Join(ss, sep))
	return io.WriteString(w, s)
}

// Stage1Extract reads raw crossref works messages (one JSON per line) from r
// and writes a tab-separated "lineno<TAB>indexed-date<TAB>DOI" record for each
// non-excluded document to w. Up to cfg.ErrCountThreshold JSON/date errors are
// tolerated (logged and skipped); beyond that the error is returned.
func Stage1Extract(cfg Stage1Config, excludes map[string]struct{}, r io.Reader, w io.Writer) error {
	var numErrs atomic.Int64 // error count across threads
	pp := parallel.NewProcessor(r, w, func(lineno int64, b []byte) ([]byte, error) {
		var (
			// This was a crossref.Document, but we only need a few fields.
			doc struct {
				DOI       string
				Deposited crossref.DateField `json:"deposited"`
				Indexed   crossref.DateField `json:"indexed"`
			}
			buf bytes.Buffer
		)
		if err := json.Unmarshal(b, &doc); err != nil {
			// Encountered with a single document found,
			// {"DOI":"10.15215\/aupress\/9781897425909.026","score":8.143441}
			numErrs.Add(1)
			if n := numErrs.Load(); n > cfg.ErrCountThreshold {
				return nil, err
			} else {
				log.Printf("skipping error (#err: %d <= max: %d): %v", n, cfg.ErrCountThreshold, err)
			}
			return nil, nil
		}
		date, err := doc.Indexed.Date()
		if err != nil {
			// Encountered with a single document found,
			// {"DOI":"10.15215\/aupress\/9781897425909.026","score":8.143441}
			numErrs.Add(1)
			if n := numErrs.Load(); n > cfg.ErrCountThreshold {
				return nil, err
			} else {
				log.Printf("skipping error (#err: %d <= max: %d): %v", n, cfg.ErrCountThreshold, err)
			}
			return nil, nil
		}
		if _, ok := excludes[doc.DOI]; ok {
			return nil, nil
		}
		if _, err := WriteFields(&buf, "\t", lineno+1, date.Format("2006-01-02"), doc.DOI); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	})
	pp.BatchSize = cfg.BatchSize
	return pp.Run()
}

// CopyFile copies the contents from src to dst using io.Copy.  If dst does not
// exist, CopyFile creates it with permissions perm; otherwise CopyFile
// truncates it before writing. From: https://codereview.appspot.com/152180043
func CopyFile(dst, src string, perm os.FileMode) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()
	_, err = io.Copy(out, in)
	return
}
