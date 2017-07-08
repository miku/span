package thieme

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/assetutil"
	"github.com/miku/span/formats/finc"
)

const (
	SourceID       = "60"
	Format         = "ElectronicArticle"
	Collection     = "Thieme E-Journals"
	Genre          = "article"
	DefaultRefType = "EJOUR"
)

func leftPad(s string, padStr string, overallLen int) string {
	padCountInt := 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

var LanguageMap = assetutil.MustLoadStringMap("assets/doaj/language-iso-639-3.json")

type Document struct {
	xml.Name `xml:"Article"`
	Journal  struct {
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
		Authors []struct {
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
}

func (doc Document) DOI() (string, error) {
	for _, id := range doc.ArticleIdList {
		if id.ArticleId.IdType == "doi" {
			return id.ArticleId.Id, nil
		}
	}
	return "", fmt.Errorf("no DOI found")
}

func (doc Document) RecordID() (string, error) {
	doi, err := doc.DOI()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ai-60-%s", base64.RawURLEncoding.EncodeToString([]byte(doi))), nil
}

func (doc Document) Date() (time.Time, error) {
	pd := doc.Journal.PubDate

	if pd.Month == "0" {
		pd.Month = "01"
	}

	if pd.Day == "0" {
		pd.Day = "01"
	}

	if pd.Year != "" && pd.Month != "" && pd.Day != "" {
		s := fmt.Sprintf("%s-%s-%s", leftPad(pd.Year, "0", 4), leftPad(pd.Month, "0", 2), leftPad(pd.Day, "0", 2))
		return time.Parse("2006-01-02", s)
	}
	if pd.Year != "" && pd.Month != "" {
		s := fmt.Sprintf("%s-%s-01", leftPad(pd.Year, "0", 4), leftPad(pd.Month, "0", 2))
		return time.Parse("2006-01-02", s)
	}
	if pd.Year != "" {
		s := fmt.Sprintf("%s-01-01", leftPad(pd.Year, "0", 4))
		return time.Parse("2006-01-02", s)
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func (doc Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	id, err := doc.RecordID()
	if err != nil {
		return output, err
	}
	output.RecordID = id
	output.SourceID = SourceID
	output.MegaCollection = Collection
	output.Genre = Genre
	output.Format = Format

	doi, err := doc.DOI()
	if err != nil {
		return output, err
	}
	output.DOI = doi

	date, err := doc.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	output.Date = date
	output.RawDate = date.Format("2006-01-02")

	journal := doc.Journal
	if journal.PublisherName != "" {
		output.Publishers = append(output.Publishers, journal.PublisherName)
	}

	if journal.JournalTitle == "" {
		return output, span.Skip{Reason: fmt.Sprintf("NO_JTITLE %s", output.RecordID)}
	}

	output.JournalTitle = journal.JournalTitle
	output.ISSN = append(output.ISSN, journal.ISSN)
	output.EISSN = append(output.EISSN, journal.EISSN)
	output.Volume = journal.Volume
	output.Issue = journal.Issue

	output.ArticleTitle = doc.ArticleTitle
	if output.ArticleTitle == "" {
		output.ArticleTitle = doc.VernacularTitle
	}

	output.Abstract = doc.Abstract
	if output.Abstract == "" {
		output.Abstract = doc.VernacularAbstract
	}

	output.URL = append(output.URL, doc.Links...)

	var subjects []string
	for _, s := range doc.Subject {
		if len(strings.TrimSpace(s)) > 0 {
			subjects = append(subjects, s)
		}
	}
	output.Subjects = subjects

	if doc.Language != "" {
		output.Languages = append(output.Languages, LanguageMap.LookupDefault(strings.ToUpper(doc.Language), "und"))
	} else {
		if doc.VernacularLanguage != "" {
			output.Languages = append(output.Languages, LanguageMap.LookupDefault(strings.ToUpper(doc.VernacularLanguage), "und"))
		}
	}

	var authors []finc.Author
	for _, author := range doc.AuthorList.Authors {
		authors = append(authors, finc.Author{FirstName: author.FirstName, LastName: author.LastName})
	}
	output.Authors = authors
	output.RefType = DefaultRefType

	return output, nil
}
