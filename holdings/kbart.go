package holdings

import (
	"encoding/csv"
	"io"
)

// KBART: Knowledge Bases And Related Tools working group. A single holding
// file entry.
type KBART struct {
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
	Embargo                  string
	CoverageDepth            string
	CoverageNotes            string
	PublisherName            string
	InterlibraryRelevance    string
	InterlibraryNationwide   string
	InterlibraryTransmission string
	InterlibraryComment      string
	Publisher                string
	Anchor                   string
	ZDBID                    string
}

// KBARTEntries are a list of holding file entries.
type KBARTEntries []KBART

// NewKBARTEntries creates a new list of kbart entries from a reader.
func NewKBARTEntries(r io.Reader) (KBARTEntries, error) {
	var k KBARTEntries
	reader := csv.NewReader(r)
	reader.Comma = '\t'

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, err
		}
		k = append(k, KBART{
			PublicationTitle:         record[0],
			PrintIdentifier:          record[1],
			OnlineIdentifier:         record[2],
			FirstIssueDate:           record[3],
			FirstVolume:              record[4],
			FirstIssue:               record[5],
			LastIssueDate:            record[6],
			LastVolume:               record[7],
			LastIssue:                record[8],
			TitleURL:                 record[9],
			FirstAuthor:              record[10],
			TitleID:                  record[11],
			Embargo:                  record[12],
			CoverageDepth:            record[13],
			CoverageNotes:            record[14],
			PublisherName:            record[15],
			Anchor:                   record[16],
			InterlibraryRelevance:    record[17],
			InterlibraryNationwide:   record[18],
			InterlibraryTransmission: record[19],
			InterlibraryComment:      record[20],
			Publisher:                record[21],
			ZDBID:                    record[23]})
	}
	return k, nil
}
