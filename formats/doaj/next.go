package doaj

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ArticleV1 represents an API v1 response.
type ArticleV1 struct {
	Bibjson struct {
		Abstract string `json:"abstract"`
		Author   []struct {
			Name string `json:"name"`
		} `json:"author"`
		EndPage    string `json:"end_page"`
		Identifier []struct {
			Id   string `json:"id"`
			Type string `json:"type"`
		} `json:"identifier"`
		Journal struct {
			Country  string   `json:"country"`
			Issns    []string `json:"issns"`
			Language []string `json:"language"`
			License  []struct {
				OpenAccess bool   `json:"open_access"`
				Title      string `json:"title"`
				Type       string `json:"type"`
				Url        string `json:"url"`
			} `json:"license"`
			Number    string `json:"number"`
			Publisher string `json:"publisher"`
			Title     string `json:"title"`
			Volume    string `json:"volume"`
		} `json:"journal"`
		Link []struct {
			ContentType string `json:"content_type"`
			Type        string `json:"type"`
			Url         string `json:"url"`
		} `json:"link"`
		StartPage string `json:"start_page"`
		Subject   []struct {
			Code   string `json:"code"`
			Scheme string `json:"scheme"`
			Term   string `json:"term"`
		} `json:"subject"`
		Title string `json:"title"`
		Year  string `json:"year"`
	} `json:"bibjson"`
	CreatedDate string `json:"created_date"`
	Id          string `json:"id"`
	LastUpdated string `json:"last_updated"`
}

// Date return the document date. Journals entries usually have no date, so
// they will err.
func (doc ArticleV1) Date() (time.Time, error) {
	if doc.CreatedDate != "" {
		return time.Parse("2006-01-02T15:04:05Z", doc.Index.Date)
	}
	var s string
	if y, err := strconv.Atoi(doc.BibJSON.Year); err == nil {
		s = fmt.Sprintf("%04d-01-01", y)
		if m, err := strconv.Atoi(doc.BibJSON.Month); err == nil {
			if m > 0 && m < 13 {
				s = fmt.Sprintf("%04d-%02d-01", y, m)
			}
		}
	}
	return time.Parse("2006-01-02", s)
}

// DOI returns the DOI or the empty string.
func (doc ArticleV1) DOI() string {
	for _, identifier := range doc.BibJSON.Identifier {
		if identifier.Type == "doi" {
			id := strings.TrimSpace(identifier.ID)
			if !strings.Contains(id, "http") {
				return id
			}
			return DOIPattern.FindString(id)
		}
	}
	return ""
}
