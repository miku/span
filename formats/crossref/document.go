package crossref

// Author is given by family and given name.
type Author struct {
	Family string `json:"family"`
	Given  string `json:"given"`
}

// DatePart consists of up to three int, representing year, month, day.
type DatePart []int

// DateField contains two representations of one value.
type DateField struct {
	DateParts []DatePart `json:"date-parts"`
	Timestamp int64      `json:"timestamp"`
}

// Document is a example 'works' API response message.
// http://api.crossref.org/works/10.1002/14651858.cd000179
type Document struct {
	Authors        []Author  `json:"author"`
	ContainerTitle []string  `json:"container-title"`
	Deposited      DateField `json:"deposited"`
	DOI            string    `json:"DOI"`
	Indexed        DateField `json:"indexed"`
	ISSN           []string  `json:"ISSN"`
	Issue          string    `json:"issue"`
	Issued         DateField `json:"issued"`
	Member         string    `json:"member"`
	Page           string    `json:"page"`
	Prefix         string    `json:"prefix"`
	Publisher      string    `json:"publisher"`
	ReferenceCount int       `json:"reference-count"`
	Score          float64   `json:"score"`
	Source         string    `json:"source"`
	Subjects       []string  `json:"subject"`
	Subtitle       []string  `json:"subtitle"`
	Title          []string  `json:"title"`
	Type           string    `json:"type"`
	URL            string    `json:"URL"`
	Volume         string    `json:"volume"`
}
