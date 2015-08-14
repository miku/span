package crossref

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/assetutil"
	"github.com/miku/span/finc"
)

const (
	// Internal bookkeeping.
	SourceID = "49"
)

var (
	errNoDate = errors.New("date is missing")
	errNoURL  = errors.New("URL is missing")
)

var (
	DefaultFormat = "ElectronicArticle"

	// Load assets
	Formats  = assetutil.MustLoadStringMap("assets/crossref/formats.json")
	Genres   = assetutil.MustLoadStringMap("assets/crossref/genres.json")
	RefTypes = assetutil.MustLoadStringMap("assets/crossref/reftypes.json")

	// AuthorReplacer is a special cleaner for author names.
	AuthorReplacer = strings.NewReplacer("#", "", "--", "", "*", "", "|", "", "&NA;", "", "\u0026NA;", "", "\u0026", "")

	// ArticleTitleBlocker will trigger skips, if article title matches exactly.
	ArticleTitleBlocker = []string{"Titelei", "Front Matter"}

	// Future, if a publication date lies beyond it, it gets skipped.
	Future = time.Now().Add(time.Hour * 24 * 365 * 5)
)

// Crossref source.
type Crossref struct{}

func (c Crossref) Iterate(r io.Reader) (<-chan []span.Importer, error) {
	return span.FromJSON(r, func(s string) (span.Importer, error) {
		doc := new(Document)
		err := json.Unmarshal([]byte(s), doc)
		return doc, err
	})
}

// Author is given by family and given name.
type Author struct {
	Family string `json:"family"`
	Given  string `json:"given"`
}

// FamilyCleaned returns a mostly clean family name.
func (author *Author) FamilyCleaned() string {
	return AuthorReplacer.Replace(span.UnescapeTrim(author.Family))
}

// GivenCleaned returns a mostly clean family name.
func (author *Author) GivenCleaned() string {
	return AuthorReplacer.Replace(span.UnescapeTrim(author.Given))
}

// String pretty prints the author.
func (author *Author) String() string {
	var given, family = author.GivenCleaned(), author.FamilyCleaned()
	if given != "" {
		if family != "" {
			return fmt.Sprintf("%s, %s", family, given)
		}
		return given
	}
	return family
}

// DatePart consists of up to three int, representing year, month, day.
type DatePart []int

// DateField contains two representations of one value.
type DateField struct {
	DateParts []DatePart `json:"date-parts"`
	Timestamp int64      `json:"timestamp"`
}

// Document is a example 'works' API response.
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

// PageInfo holds various page related data.
type PageInfo struct {
	RawMessage string
	StartPage  int
	EndPage    int
}

// PageCount returns the number of pages, or zero if this cannot be determined.
func (pi *PageInfo) PageCount() int {
	if pi.StartPage != 0 && pi.EndPage != 0 {
		count := pi.EndPage - pi.StartPage
		if count > 0 {
			return count
		}
	}
	return 0
}

// RecordID is of the form <kind>-<source-id>-<id-base64-unpadded>
// We simple map any primary key of the source (preferably a URL)
// to a safer alphabet. Since the base64 part is not meant to be decoded
// we drop the padding. It is simple enough to recover the original value.
func (doc *Document) RecordID() string {
	enc := fmt.Sprintf("ai-%s-%s", SourceID, base64.URLEncoding.EncodeToString([]byte(doc.URL)))
	return strings.TrimRight(enc, "=")
}

// PageInfo parses a page specfication in a best effort manner into a PageInfo struct.
func (doc *Document) PageInfo() PageInfo {
	pi := PageInfo{RawMessage: doc.Page}
	parts := strings.Split(doc.Page, "-")
	if len(parts) != 2 {
		return pi
	}
	spage, err := strconv.Atoi(parts[0])
	if err != nil {
		return pi
	}
	pi.StartPage = spage

	epage, err := strconv.Atoi(parts[1])
	if err != nil {
		return pi
	}
	pi.EndPage = epage
	return pi
}

// Date returns a time.Date in a best effort manner. Date parts seem to be always
// present in the source document, while timestamp is only present if
// dateparts consist of all three: year, month and day.
// It is an error, if no valid date can be extracted.
func (d *DateField) Date() (t time.Time, err error) {
	if len(d.DateParts) == 0 {
		return t, errNoDate
	}
	parts := d.DateParts[0]
	switch len(parts) {
	case 1:
		t, err = time.Parse("2006-01-02", fmt.Sprintf("%04d-01-01", parts[0]))
		if err != nil {
			return t, err
		}
	case 2:
		t, err = time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-01", parts[0], parts[1]))
		if err != nil {
			return t, err
		}
	case 3:
		t, err = time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-%02d", parts[0], parts[1], parts[2]))
		if err != nil {
			return t, err
		}
	}
	return t, err
}

// CombinedTitle returns a longish title.
func (doc *Document) CombinedTitle() string {
	if len(doc.Title) > 0 {
		if len(doc.Subtitle) > 0 {
			return span.UnescapeTrim(fmt.Sprintf("%s : %s", doc.Title[0], doc.Subtitle[0]))
		}
		return span.UnescapeTrim(doc.Title[0])
	}
	if len(doc.Subtitle) > 0 {
		return span.UnescapeTrim(doc.Subtitle[0])
	}
	return ""
}

// FullTitle returns everything title.
func (doc *Document) FullTitle() string {
	return span.UnescapeTrim(strings.Join(append(doc.Title, doc.Subtitle...), " "))
}

// ShortTitle returns the first main title only.
func (doc *Document) ShortTitle() (s string) {
	if len(doc.Title) > 0 {
		s = span.UnescapeTrim(doc.Title[0])
	}
	return
}

// MemberName resolves the primary name of the member.
func (doc *Document) MemberName() (name string, err error) {
	id, err := doc.parseMemberID()
	if err != nil {
		return
	}
	name, err = LookupMemberName(id)
	return
}

// ParseMemberID extracts the numeric member id.
func (doc *Document) parseMemberID() (id int, err error) {
	fields := strings.Split(doc.Member, "/")
	if len(fields) > 0 {
		id, err = strconv.Atoi(fields[len(fields)-1])
		if err != nil {
			return id, fmt.Errorf("invalid member: %s", doc.Member)
		}
		return id, nil
	}
	return id, fmt.Errorf("invalid member: %s", doc.Member)
}

// ToIntermediateSchema converts a crossref document into IS.
func (doc *Document) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error
	output := finc.NewIntermediateSchema()

	output.Date, err = doc.Issued.Date()

	if err != nil {
		return output, err
	}

	output.RawDate = output.Date.Format("2006-01-02")

	if doc.URL == "" {
		return output, errNoURL
	}

	output.RecordID = doc.RecordID()
	if len(output.RecordID) > span.KeyLengthLimit {
		return output, span.Skip{Reason: fmt.Sprintf("ID_TOO_LONG %s", output.RecordID)}
	}

	if output.Date.After(Future) {
		return output, span.Skip{Reason: fmt.Sprintf("TOO_FUTURISTIC %s", output.RecordID)}
	}

	if doc.Type == "journal-issue" {
		return output, span.Skip{Reason: fmt.Sprintf("JOURNAL_ISSUE %s", output.RecordID)}
	}

	output.ArticleTitle = doc.CombinedTitle()
	if len(output.ArticleTitle) == 0 {
		return output, span.Skip{Reason: fmt.Sprintf("NO_ATITLE %s", output.RecordID)}
	}

	for _, title := range ArticleTitleBlocker {
		if output.ArticleTitle == title {
			return output, span.Skip{Reason: fmt.Sprintf("BLOCKED_ATITLE %s", output.RecordID)}
		}
	}

	output.DOI = doc.DOI
	output.Format = Formats.LookupDefault(doc.Type, DefaultFormat)
	output.Genre = Genres.LookupDefault(doc.Type, "unknown")
	output.ISSN = doc.ISSN
	output.Issue = doc.Issue
	output.Languages = []string{"eng"}
	output.Publishers = append(output.Publishers, doc.Publisher)
	output.RefType = RefTypes.LookupDefault(doc.Type, "GEN")
	output.SourceID = SourceID
	output.Subjects = doc.Subjects
	output.Type = doc.Type
	output.URL = append(output.URL, doc.URL)
	output.Volume = doc.Volume

	if len(doc.ContainerTitle) > 0 {
		output.JournalTitle = span.UnescapeTrim(doc.ContainerTitle[0])
	} else {
		return output, span.Skip{Reason: fmt.Sprintf("NO_JTITLE %s", output.RecordID)}
	}

	if len(doc.Subtitle) > 0 {
		output.ArticleSubtitle = span.UnescapeTrim(doc.Subtitle[0])
	}

	for _, author := range doc.Authors {
		output.Authors = append(output.Authors, finc.Author{
			FirstName: author.GivenCleaned(),
			LastName:  author.FamilyCleaned()})
	}

	// TODO(miku): do we need a config for these things?
	// Maybe a generic filter (in js?) that will gather exclusion rules?
	// if len(output.Authors) == 0 {
	// 	return output, span.Skip{Reason: fmt.Sprintf("NO_AUTHORS %s", output.RecordID)}
	// }

	pi := doc.PageInfo()
	output.StartPage = fmt.Sprintf("%d", pi.StartPage)
	output.EndPage = fmt.Sprintf("%d", pi.EndPage)
	output.Pages = pi.RawMessage
	output.PageCount = fmt.Sprintf("%d", pi.PageCount())

	name, err := doc.MemberName()
	if err == nil {
		output.MegaCollection = fmt.Sprintf("%s (CrossRef)", name)
	} else {
		// TODO(miku): Use publisher as fallback, refs. #4833
		output.MegaCollection = fmt.Sprintf("X-U (CrossRef)")
	}

	return output, nil
}
