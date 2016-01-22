package holdings

import (
	"encoding/csv"
	"io"
)

// KBART: Knowledge Bases And Related Tools working group. A single holding
// file entry.
type KBART struct {
	PublicationTitle         string `json:"title"`
	PrintIdentifier          string `json:"print-identifier"`
	OnlineIdentifier         string `json:"online-identifier"`
	FirstIssueDate           string `json:"first-issue-date"`
	FirstVolume              string `json:"first-volume"`
	FirstIssue               string `json:"first-issue"`
	LastIssueDate            string `json:"last-issue-date"`
	LastVolume               string `json:"last-volume"`
	LastIssue                string `json:"last-issue"`
	TitleURL                 string `json:"title-url"`
	FirstAuthor              string `json:"first-author"`
	TitleID                  string `json:"title-id"`
	Embargo                  string `json:"embargo"`
	CoverageDepth            string `json:"coverage-depth"`
	CoverageNotes            string `json:"coverage-notes"`
	PublisherName            string `json:"publisher-name"`
	InterlibraryRelevance    string `json:"relevance"`
	InterlibraryNationwide   string `json:"nationwide"`
	InterlibraryTransmission string `json:"transmission"`
	InterlibraryComment      string `json:"comment"`
	Publisher                string `json:"publisher"`
	Anchor                   string `json:"anchor"`
	ZDBID                    string `json:"zdb-id"`
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
