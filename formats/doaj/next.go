package doaj

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/container"
	"github.com/miku/span/formats/finc"
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
		Keywords []string `json:"keywords"`
		Link     []struct {
			Type string `json:"type"`
			Url  string `json:"url"`
		} `json:"link"`
		Month     string `json:"month"`
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

// Date return the document date.
func (doc ArticleV1) Date() (time.Time, error) {
	if doc.CreatedDate != "" {
		return time.Parse("2006-01-02T15:04:05Z", doc.CreatedDate)
	}
	var s string
	if y, err := strconv.Atoi(doc.Bibjson.Year); err == nil {
		s = fmt.Sprintf("%04d-01-01", y)
		if m, err := strconv.Atoi(doc.Bibjson.Month); err == nil {
			if m > 0 && m < 13 {
				s = fmt.Sprintf("%04d-%02d-01", y, m)
			}
		}
	}
	return time.Parse("2006-01-02", s)
}

// DOI returns the DOI or the empty string.
func (doc ArticleV1) DOI() string {
	for _, identifier := range doc.Bibjson.Identifier {
		if identifier.Type == "doi" {
			id := strings.TrimSpace(identifier.Id)
			if !strings.Contains(id, "http") {
				return id
			}
			return DOIPattern.FindString(id)
		}
	}
	return ""
}

// Authors returns a list of authors.
func (doc ArticleV1) Authors() (authors []finc.Author) {
	for _, author := range doc.Bibjson.Author {
		authors = append(authors, finc.Author{Name: html.UnescapeString(author.Name)})
	}
	return authors
}

// ToIntermediateSchema converts a doaj document to intermediate schema. For
// now any record, that has no usable date will be skipped.
func (doc ArticleV1) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error

	output := finc.NewIntermediateSchema()
	output.Date, err = doc.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	output.RawDate = output.Date.Format("2006-01-02")

	if doc.Id == "" {
		return output, span.Skip{Reason: "no identifier in source"}
	}
	id := fmt.Sprintf("ai-%s-%s", SourceIdentifier, doc.Id)
	if len(id) > span.KeyLengthLimit {
		return output, span.Skip{Reason: fmt.Sprintf("id too long: %s", id)}
	}

	output.ArticleTitle = doc.Bibjson.Title
	output.Authors = doc.Authors()
	output.DOI = doc.DOI()
	output.Format = Format
	output.Genre = Genre

	output.ISSN = doc.Bibjson.Journal.Issns
	output.JournalTitle = doc.Bibjson.Journal.Title
	output.MegaCollections = []string{Collection}

	publisher := strings.TrimSpace(doc.Bibjson.Journal.Publisher)
	if publisher != "" {
		output.Publishers = append(output.Publishers, publisher)
	}

	output.RecordID = doc.Id
	output.ID = id
	output.SourceID = SourceIdentifier
	output.Volume = doc.Bibjson.Journal.Volume

	// refs. #8709
	if output.DOI != "" {
		output.URL = append(output.URL, "http://doi.org/"+output.DOI)
	}

	// refs. #8709
	if len(output.URL) == 0 {
		for _, link := range doc.Bibjson.Link {
			output.URL = append(output.URL, link.Url)
		}
	}

	// refs. #6634
	if len(output.URL) == 0 {
		output.URL = append(output.URL, "https://doaj.org/article/"+doc.Id)
	}

	output.StartPage = doc.Bibjson.StartPage
	output.EndPage = doc.Bibjson.EndPage

	if sp, err := strconv.Atoi(doc.Bibjson.StartPage); err == nil {
		if ep, err := strconv.Atoi(doc.Bibjson.EndPage); err == nil {
			output.PageCount = fmt.Sprintf("%d", ep-sp)
			output.Pages = fmt.Sprintf("%d-%d", sp, ep)
		}
	}

	subjects := container.NewStringSet()
	for _, s := range doc.Bibjson.Subject {
		class := LCCPatterns.LookupDefault(s.Code, finc.NotAssigned)
		if class != finc.NotAssigned {
			subjects.Add(class)
		}
	}
	if subjects.Size() == 0 {
		output.Subjects = []string{finc.NotAssigned}
	} else {
		output.Subjects = subjects.SortedValues()
	}

	languages := container.NewStringSet()
	for _, l := range doc.Bibjson.Journal.Language {
		detected := span.LanguageIdentifier(l)
		if detected == "" {
			detected = "und"
		}
		languages.Add(detected)
	}
	output.Languages = languages.Values()

	output.RefType = DefaultRefType
	return output, nil
}
