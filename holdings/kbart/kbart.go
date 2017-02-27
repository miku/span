package kbart

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/container"
	"github.com/miku/span/holdings"
)

var (
	ErrIncompleteLine     = errors.New("incomplete KBART line")
	ErrIncompleteEmbargo  = errors.New("incomplete embargo")
	ErrInvalidEmbargo     = errors.New("invalid embargo")
	ErrMissingIdentifiers = errors.New("missing identifiers")
)

var (
	// delayPattern fixes allowed embargo strings.
	delayPattern = regexp.MustCompile(`([P|R])([0-9]+)([Y|M|D])`)
)

// embargo is a string representing a delay, e.g. P1Y, R10M.
type embargo string

// entry represents the various columns.
type columns struct {
	PublicationTitle         string
	PrintIdentifier          string
	OnlineIdentifier         string
	FirstIssueDate           string
	FirstVolume              string
	FirstIssue               string
	LastIssueDate            string
	LastVolume               string
	LastIssue                string
	TitleURL                 string
	FirstAuthor              string
	TitleID                  string
	Embargo                  embargo
	CoverageDepth            string
	CoverageNotes            string
	PublisherName            string
	InterlibraryRelevance    string
	InterlibraryNationwide   string
	InterlibraryTransmission string
	InterlibraryComment      string
	Publisher                string
	// Anchor might contain ISSNs, too
	Anchor string
	ZDBID  string
}

// Convert string like P12M, P1M, R10Y into a time.Duration.
func (e embargo) AsDuration() (time.Duration, error) {
	var d time.Duration

	emb := strings.TrimSpace(string(e))
	if len(emb) == 0 {
		return d, nil
	}

	var parts = delayPattern.FindStringSubmatch(emb)

	if len(parts) == 0 {
		return d, nil
	}

	if len(parts) < 4 {
		return d, ErrIncompleteEmbargo
	}

	i, err := strconv.Atoi(parts[2])
	if err != nil {
		return d, ErrInvalidEmbargo
	}

	switch parts[3] {
	case "D":
		return time.Duration(-i) * 24 * time.Hour, nil
	case "M":
		return time.Duration(-i) * 24 * time.Hour * 30, nil
	case "Y":
		return time.Duration(-i) * 24 * time.Hour * 365, nil
	default:
		return d, ErrInvalidEmbargo
	}
}

// DisallowEarlier returns true if dates *before* the boundary should be
// disallowed.
func (e embargo) DisallowEarlier() bool {
	return strings.HasPrefix(strings.TrimSpace(string(e)), "R")
}

// Reader reads tab-separated KBART. The encoding/csv package did not like
// that particular format so we use a simple bufio.Reader for now.
type Reader struct {
	r          *bufio.Reader
	currentRow int

	SkipFirstRow           bool
	SkipMissingIdentifiers bool
	SkipIncompleteLines    bool
	SkipInvalidEmbargo     bool
}

// NewReader creates a new KBART reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		SkipFirstRow:           true,
		SkipMissingIdentifiers: true,
		// some files have weird headers, refs. #7004
		SkipIncompleteLines: true,
		r:                   bufio.NewReader(r),
	}
}

// ReadEntries returns a list of entries.
func (r *Reader) ReadEntries() (holdings.Entries, error) {
	entries := make(holdings.Entries)

	for {
		cols, entry, err := r.Read()

		if err == io.EOF {
			break
		}

		switch err {
		case nil: // pass
		case ErrMissingIdentifiers:
			if r.SkipMissingIdentifiers {
				log.Println("skipping line with missing identifiers")
				continue
			} else {
				return entries, err
			}
		case ErrIncompleteLine:
			if r.SkipIncompleteLines {
				log.Println("skipping incomplete line")
				continue
			} else {
				return entries, err
			}
		case ErrInvalidEmbargo:
			if r.SkipInvalidEmbargo {
				log.Println("skipping invalid embargo")
				continue
			} else {
				return entries, err
			}
		default:
			return entries, err
		}

		pi := strings.TrimSpace(cols.PrintIdentifier)
		oi := strings.TrimSpace(cols.OnlineIdentifier)

		// Slight ISSN restoration (e.g. http://www.jstor.org/kbart/collections/as).
		if len(pi) == 8 {
			pi = fmt.Sprintf("%s-%s", pi[:4], pi[4:])
		}

		if len(oi) == 8 {
			oi = fmt.Sprintf("%s-%s", oi[:4], oi[4:])
		}

		// Collect all identifiers.
		identifiers := container.NewStringSet()
		if pi != "" {
			identifiers.Add(pi)
		}
		if oi != "" {
			identifiers.Add(oi)
		}

		// Extract ISSN from anchor field.
		for _, issn := range span.ISSNPattern.FindAllString(cols.Anchor, -1) {
			identifiers.Add(issn)
		}

		if identifiers.Size() == 0 {
			if !r.SkipMissingIdentifiers {
				return entries, ErrMissingIdentifiers
			}
		}

		for _, id := range identifiers.Values() {
			entries[id] = append(entries[id], holdings.License(entry))
		}
	}

	return entries, nil
}

// Read reads a single line. Returns all columns, a parsed entry and error.
func (r *Reader) Read() (columns, holdings.Entry, error) {
	var entry holdings.Entry
	var cols columns

	if r.SkipFirstRow && r.currentRow == 0 {
		if _, err := r.r.ReadString('\n'); err != nil {
			return cols, entry, err
		}
	}
	r.currentRow++

	var line string
	var err error

	for {
		line, err = r.r.ReadString('\n')
		if strings.TrimSpace(line) != "" {
			break
		}
		if err == io.EOF {
			return cols, entry, io.EOF
		}
	}

	record := strings.Split(line, "\t")

	if err == io.EOF {
		return cols, entry, io.EOF
	}
	if err != nil {
		return cols, entry, err
	}
	// usually 23 columns, but might be only 19, refs. #7004
	if len(record) < 19 {
		log.Printf("warning: KBART line has (about) %d columns, expected 19 or 23", len(record))
	}

	// if less than 13, bail out, refs. #7004
	if len(record) < 13 {
		return cols, entry, ErrIncompleteLine
	}

	cols = columns{
		PrintIdentifier:  record[1],
		OnlineIdentifier: record[2],
		FirstIssueDate:   record[3],
		FirstVolume:      record[4],
		FirstIssue:       record[5],
		LastIssueDate:    record[6],
		LastVolume:       record[7],
		LastIssue:        record[8],
		Embargo:          embargo(record[12]),
	}

	if len(record) > 21 {
		cols.Anchor = record[21]
	}

	emb, err := cols.Embargo.AsDuration()
	if err != nil {
		return cols, entry, err
	}

	entry = holdings.Entry{
		Begin: holdings.Signature{
			Date:   cols.FirstIssueDate,
			Volume: cols.FirstVolume,
			Issue:  cols.FirstIssue,
		},
		End: holdings.Signature{
			Date:   cols.LastIssueDate,
			Volume: cols.LastVolume,
			Issue:  cols.LastIssue,
		},
		Embargo:                emb,
		EmbargoDisallowEarlier: cols.Embargo.DisallowEarlier(),
	}

	return cols, entry, nil
}
