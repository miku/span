package doaj

type Response struct {
	ID     string   `json:"_id"`
	Index  string   `json:"_index"`
	Source Document `json:"_source"`
	Type   string   `json:"_type"`
}

type Document struct {
	BibJson BibJSON `json:"bibjson"`
	Created string  `json:"created_date"`
	ID      string  `json:"id"`
	Index   Index   `json:"index"`
	Updated string  `json:"last_updated"`
	// make Response.Type available here
	Type string
}

type Index struct {
	Classification []string `json:"classification"`
	Country        string   `json:"country"`
	Date           string   `json:"date"`
	ISSN           []string `json:"issn"`
	Language       []string `json:"language"`
	License        []string `json:"license"`
	Publishers     []string `json:"publisher"`
	SchemaCode     []string `json:"schema_code"`
	SchemaSubjects []string `json:"schema_subjects"`
	Subjects       []string `json:"subject"`
}

type License struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}

type Journal struct {
	Country   string    `json:"country"`
	Language  []string  `json:"language"`
	License   []License `json:"license"`
	Number    string    `json:"number"`
	Publisher string    `json:"publisher"`
	Title     string    `json:"title"`
	Volume    string    `json:"volume"`
}

type Author struct {
	Name string `json:"name"`
}

type Subject struct {
	Code   string `json:"code"`
	Scheme string `json:"scheme"`
	Term   string `json:"term"`
}

type Link struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Identifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type BibJSON struct {
	Abstract   string       `json:"abstract"`
	Author     []Author     `json:"author"`
	EndPage    string       `json:"end_page"`
	Identifier []Identifier `json:"identifier"`
	Journal    Journal      `json:"journal"`
	Link       []Link       `json:"link"`
	Month      string       `json:"month"`
	StartPage  string       `json:"start_page"`
	Subject    []Subject    `json:"subject"`
	Title      string       `json:"title"`
	Year       string       `json:"year"`
}
