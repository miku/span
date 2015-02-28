package holdings

// Item is the main google scholar holdings container
type Item struct {
	Title string     `xml:"title"`
	ISSN  string     `xml:"issn"`
	Covs  []Coverage `xml:"coverage"`
}

// Coverage contains coverage information for an item
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
