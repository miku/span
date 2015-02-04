package holdings

type Holding struct {
	EZBID        int           `xml:"ezb_id,attr"`
	Title        string        `xml:"title"`
	Publishers   string        `xml:"publishers"`
	PISSN        []string      `xml:"EZBIssns>p-issn"`
	EISSN        []string      `xml:"EZBIssns>e-issn"`
	Entitlements []Entitlement `xml:"entitlements>entitlement"`
}

type Entitlement struct {
	Status     string `xml:"status,attr"`
	URL        string `xml:"url"`
	Anchor     string `xml:"anchor"`
	FromYear   int    `xml:"begin>year"`
	FromVolume int    `xml:"begin>volume"`
	FromIssue  int    `xml:"begin>issue"`
	FromDelay  string `xml:"begin>delay"`
	ToYear     int    `xml:"end>year"`
	ToVolume   int    `xml:"end>volume"`
	ToIssue    int    `xml:"end>issue"`
	ToDelay    string `xml:"end>delay"`
}
