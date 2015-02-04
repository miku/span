package holdings

import (
	"fmt"
	"time"
)

type Item struct {
	Title string     `xml:"title"`
	ISSN  string     `xml:"issn"`
	Covs  []Coverage `xml:"coverage"`
}

type Coverage struct {
	FromYear         int    `xml:"from>year"`
	FromVolume       int    `xml:"from>volume"`
	FromIssue        int    `xml:"from>issue"`
	ToYear           int    `xml:"to>year"`
	ToVolume         int    `xml:"to>volume"`
	ToIssue          int    `xml:"to>issue"`
	Comment          string `xml:"comment"`
	DaysNotAvailable int    `xml:"embargo>days_not_available"`
}

func (c *Coverage) String() string {
	return fmt.Sprintf("%d/%d/%d-%d/%d/%d", c.FromYear, c.FromVolume, c.FromIssue, c.ToYear, c.ToVolume, c.ToIssue)
}

func (c *Coverage) EffectiveCoverage() *Coverage {
	effective := Coverage{FromYear: c.FromYear, FromVolume: c.FromVolume, FromIssue: c.FromIssue}

	var wall time.Duration
	wall, _ = time.ParseDuration(fmt.Sprintf("-%dh", 24*c.DaysNotAvailable))
	now := time.Now().Add(wall)

	if c.ToYear == 0 {
		effective.ToYear = now.Year()
	} else {
		effective.ToYear = c.ToYear
	}

	effective.ToVolume = c.ToVolume
	effective.ToIssue = c.ToIssue
	effective.Comment = c.Comment
	return &effective
}
