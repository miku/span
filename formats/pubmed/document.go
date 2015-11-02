package pubmed

// Document according to http://www.ncbi.nlm.nih.gov/books/NBK3828/
// DTD http://www.ncbi.nlm.nih.gov/corehtml/query/static/PubMed.dtd
type Document struct {
	ArticleSet struct {
		Article struct {
			Journal struct {
				PublisherName string `xml:"PublisherName"`
				JournalTitle  string `xml:"JournalTitle"`
				ISSN          string `xml:"Issn"`
				EISSN         string `xml:"E-Issn"`
				Volume        string `xml:"Volume"`
				Issue         string `xml:"Issue"`
				PubDate       struct {
					Year  string `xml:"Year"`
					Month string `xml:"Month"`
					Day   string `xml:"Day"`
				} `xml:"PubDate"`
			} `xml:"Journal"`
			AuthorList struct {
				Author []struct {
					FirstName string `xml:"FirstName"`
					LastName  string `xml:"LastName"`
				} `xml:"Author"`
			} `xml:"AuthorList"`
			ArticleIdList []struct {
				ArticleId struct {
					OpenAccess string `xml:"OpenAccess,attr"`
					IdType     string `xml:"IdType,attr"`
					Id         string `xml:",chardata"`
				}
			} `xml:"ArticleIdList"`
			ArticleType        string   `xml:"ArticleType"`
			ArticleTitle       string   `xml:"ArticleTitle"`
			VernacularTitle    string   `xml:"VernacularTitle"`
			FirstPage          string   `xml:"FirstPage"`
			LastPage           string   `xml:"LastPage"`
			VernacularLanguage string   `xml:"VernacularLanguage"`
			Language           string   `xml:"Language"`
			Subject            []string `xml:"subject"`
			Links              []string `xml:"Links>Link"`
			History            []struct {
				PubDate struct {
					Status string `xml:"PubStatus,attr"`
					Year   string `xml:"Year"`
					Month  string `xml:"Month"`
					Day    string `xml:"Day"`
				} `xml:"PubDate"`
			} `xml:"History"`
			Abstract           string `xml:"Abstract"`
			VernacularAbstract string `xml:"VernacularAbstract"`
			Format             struct {
				HTML string `xml:"html,attr"`
				PDF  string `xml:"pdf,attr"`
			} `xml:"format"`
			CopyrightInformation string `xml:"CopyrightInformation"`
		} `xml:"Article"`
	} `xml:"ArticleSet"`
}
