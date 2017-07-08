// TODO.
package elseviernext

import (
	"archive/tar"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/miku/span/s/fincnext"
)

const (
	SourceID   = "85"
	Format     = "ElectronicArticle"
	Collection = "Elsevier Journals"
	Genre      = "article"
	// Default ris.type
	DefaultRefType = "EJOUR"
)

var (
	ErrNoYearFound     = errors.New("no year found")
	ErrTarFileRequired = errors.New("a tar file is required")
)

// Dataset describes journal issues and items, usually inside a dataset.xml.
type Dataset struct {
	xml.Name         `xml:"dataset"`
	SchemaVersion    string `xml:"schema-version,attr"`
	Xsi              string `xml:"xsi,attr"`
	SchemaLocation   string `xml:"schemaLocation,attr"`
	DatasetUniqueIds struct {
		ProfileCode      string `xml:"profile-code"`
		ProfileDatasetId string `xml:"profile-dataset-id"`
		Timestamp        string `xml:"timestamp"`
	} `xml:"dataset-unique-ids"`
	DatasetProperties struct {
		DatasetAction     string `xml:"dataset-action"`
		ProductionProcess string `xml:"production-process"`
	} `xml:"dataset-properties"`
	DatasetContent struct {
		JournalIssue []struct {
			Version struct {
				VersionNumber string `xml:"version-number"`
				Stage         string `xml:"stage"`
			} `xml:"version"`
			JournalIssueUniqueIds struct {
				Pii    string `xml:"pii"`
				Doi    string `xml:"doi"`
				JidAid struct {
					Jid  string `xml:"jid"`
					Issn string `xml:"issn"`
					Aid  string `xml:"aid"`
				} `xml:"jid-aid"`
			} `xml:"journal-issue-unique-ids"`
			JournalIssueProperties struct {
				Jid               string `xml:"jid"`
				Issn              string `xml:"issn"`
				VolumeIssueNumber struct {
					VolFirst string `xml:"vol-first"`
					Suppl    string `xml:"suppl"`
				} `xml:"volume-issue-number"`
				CollectionTitle string `xml:"collection-title"`
			} `xml:"journal-issue-properties"`
			FilesInfo struct {
				Ml []struct {
					Pathname   string `xml:"pathname"`
					Filesize   string `xml:"filesize"`
					Purpose    string `xml:"purpose"`
					DTDVersion string `xml:"dtd-version"`
				} `xml:"ml"`
			} `xml:"files-info"`
		} `xml:"journal-issue"`
		JournalItem []struct {
			CrossMark string `xml:"cross-mark,attr"`
			Version   struct {
				VersionNumber string `xml:"version-number"`
				Stage         string `xml:"stage"`
			} `xml:"version"`
			JournalItemUniqueIds struct {
				Pii    string `xml:"pii"`
				Doi    string `xml:"doi"`
				JidAid struct {
					Jid  string `xml:"jid"`
					Issn string `xml:"issn"`
					Aid  string `xml:"aid"`
				} `xml:"jid-aid"`
			} `xml:"journal-item-unique-ids"`
			JournalItemProperties struct {
				Pit                   string `xml:"pit"`
				ProductionType        string `xml:"production-type"`
				OnlinePublicationDate string `xml:"online-publication-date"`
			} `xml:"journal-item-properties"`
			FilesInfo struct {
				Ml []struct {
					Pathname   string `xml:"pathname"`
					Filesize   string `xml:"filesize"`
					Purpose    string `xml:"purpose"`
					DTDVersion string `xml:"dtd-version"`
				} `xml:"ml"`
			} `xml:"files-info"`
		} `xml:"journal-item"`
	} `xml:"dataset-content"`
}

// Pages is a helper, so we can calculate the total.
type Pages struct {
	FirstPage string `xml:"first-page"`
	LastPage  string `xml:"last-page"`
}

// Total number of pages. Will not do any plausibility checks.
func (p Pages) Total() string {
	f, err := strconv.Atoi(p.FirstPage)
	if err != nil {
		return ""
	}
	t, err := strconv.Atoi(p.LastPage)
	if err != nil {
		return ""
	}
	return strconv.Itoa(t - f)
}

// SerialIssue contains information about an issue, usually inside issue.xml.
type SerialIssue struct {
	xml.Name  `xml:"serial-issue"`
	IssueInfo struct {
		Pii               string `xml:"pii"`
		Jid               string `xml:"jid"`
		Issn              string `xml:"issn"`
		VolumeIssueNumber struct {
			VolFirst string `xml:"vol-first"`
			IssFirst string `xml:"iss-first"`
		} `xml:"volume-issue-number"`
	} `xml:"issue-info"`
	IssueData struct {
		CoverDate struct {
			DateRange struct {
				StartDate string `xml:"start-date"`
			} `xml:"date-range"`
		} `xml:"cover-date"`
		Pages      []Pages `xml:"pages"`
		CoverImage string  `xml:"cover-image"`
	} `xml:"issue-data"`
	IssueBody struct {
		IncludeItem []struct {
			Pii   string `xml:"pii"`
			Doi   string `xml:"doi"`
			Pages string `xml:"pages"`
		} `xml:"include-item"`
		IssueSec []struct {
			SectionTitle string `xml:"section-title"`
			IncludeItem  []struct {
				Pii   string `xml:"pii"`
				Doi   string `xml:"doi"`
				Pages Pages  `xml:"pages"`
			} `xml:"include-item"`
		} `xml:"issue-sec"`
	} `xml:"issue-body"`
}

// Article describes a single article.
type Article struct {
	xml.Name   `xml:"article"`
	Xlink      string `xml:"xlink,attr"`
	Ce         string `xml:"ce,attr"`
	Sb         string `xml:"sb,attr"`
	Docsubtype string `xml:"docsubtype,attr"`
	Version    string `xml:"version,attr"`
	Lang       string `xml:"lang,attr"`
	ItemInfo   struct {
		Jid       string `xml:"jid"`
		Aid       string `xml:"aid"`
		Pii       string `xml:"pii"`
		Doi       string `xml:"doi"`
		Copyright struct {
			Type string `xml:"type,attr"`
			Year string `xml:"year,attr"`
		} `xml:"copyright"`
	} `xml:"item-info"`
	Head struct {
		ArticleFootnote string `xml:"article-footnote"`
		Dochead         struct {
			Id     string `xml:"id,attr"`
			Textfn string `xml:"textfn"`
		} `xml:"dochead"`
		Title       string `xml:"title"`
		AuthorGroup struct {
			Id     string `xml:"id,attr"`
			Author []struct {
				Id        string `xml:"id,attr"`
				Orcid     string `xml:"orcid,attr"`
				GivenName string `xml:"given-name"`
				Surname   string `xml:"surname"`
				Degrees   string `xml:"degrees"`
				CrossRef  struct {
					Refid string `xml:"refid,attr"`
					Id    string `xml:"id,attr"`
					Sup   string `xml:"sup"`
				} `xml:"cross-ref"`
				EAddress struct {
					Type string `xml:"type,attr"`
					Id   string `xml:"id,attr"`
				} `xml:"e-address"`
			} `xml:"author"`
			Affiliation struct {
				Id          string `xml:"id,attr"`
				Textfn      string `xml:"textfn"`
				Affiliation struct {
					Sa           string   `xml:"sa,attr"`
					Organization []string `xml:"organization"`
					AddressLine  string   `xml:"address-line"`
					City         string   `xml:"city"`
					State        string   `xml:"state"`
					PostalCode   string   `xml:"postal-code"`
					Country      string   `xml:"country"`
				} `xml:"affiliation"`
			} `xml:"affiliation"`
			Correspondence struct {
				Id    string `xml:"id,attr"`
				Label string `xml:"label"`
				Text  string `xml:"text"`
			} `xml:"correspondence"`
		} `xml:"author-group"`
		DateReceived struct {
			Day   string `xml:"day,attr"`
			Month string `xml:"month,attr"`
			Year  string `xml:"year,attr"`
		} `xml:"date-received"`
		DateRevised struct {
			Day   string `xml:"day,attr"`
			Month string `xml:"month,attr"`
			Year  string `xml:"year,attr"`
		} `xml:"date-revised"`
		DateAccepted struct {
			Day   string `xml:"day,attr"`
			Month string `xml:"month,attr"`
			Year  string `xml:"year,attr"`
		} `xml:"date-accepted"`
		Abstract []struct {
			Text         string `xml:",innerxml"`
			Lang         string `xml:"lang,attr"`
			Id           string `xml:"id,attr"`
			View         string `xml:"view,attr"`
			Class        string `xml:"class,attr"`
			SectionTitle string `xml:"section-title"`
			AbstractSec  []struct {
				Role         string `xml:"role,attr"`
				Id           string `xml:"id,attr"`
				View         string `xml:"view,attr"`
				SectionTitle string `xml:"section-title"`
				SimplePara   struct {
					Id   string `xml:"id,attr"`
					View string `xml:"view,attr"`
					List struct {
						Id       string `xml:"id,attr"`
						ListItem []struct {
							Id    string `xml:"id,attr"`
							Label string `xml:"label"`
							Para  struct {
								Id   string `xml:"id,attr"`
								View string `xml:"view,attr"`
							} `xml:"para"`
						} `xml:"list-item"`
					} `xml:"list"`
				} `xml:"simple-para"`
			} `xml:"abstract-sec"`
		} `xml:"abstract"`
		Keywords struct {
			Lang         string `xml:"lang,attr"`
			Id           string `xml:"id,attr"`
			View         string `xml:"view,attr"`
			Class        string `xml:"class,attr"`
			SectionTitle string `xml:"section-title"`
			Keyword      []struct {
				Id   string `xml:"id,attr"`
				Text string `xml:"text"`
			} `xml:"keyword"`
		} `xml:"keywords"`
	} `xml:"head"`
	// in case of a "simple-article", carry this field as well
	SimpleHead struct {
		Title       string `xml:"title"`
		AuthorGroup struct {
			Id     string `xml:"id,attr"`
			Author []struct {
				Id        string `xml:"id,attr"`
				GivenName string `xml:"given-name"`
				Surname   string `xml:"surname"`
				EAddress  struct {
					Id   string `xml:"id,attr"`
					Type string `xml:"type,attr"`
				} `xml:"e-address"`
			} `xml:"author"`
			Affiliation struct {
				Id          string `xml:"id,attr"`
				Textfn      string `xml:"textfn"`
				Affiliation struct {
					Sa           string   `xml:"sa,attr"`
					Organization []string `xml:"organization"`
					AddressLine  string   `xml:"address-line"`
					City         string   `xml:"city"`
					PostalCode   string   `xml:"postal-code"`
					Country      string   `xml:"country"`
				} `xml:"affiliation"`
			} `xml:"affiliation"`
		} `xml:"author-group"`
	} `xml:"simple-head"`
}

// Date returns the date of the article. Currently use the date-received attribute.
func (article Article) Date() (time.Time, error) {
	dr := article.Head.DateReceived
	var year, month, day = dr.Year, dr.Month, dr.Day

	year = strings.TrimLeft(year, "0")
	month = strings.TrimLeft(month, "0")
	day = strings.TrimLeft(day, "0")

	if year == "" {
		if article.ItemInfo.Copyright.Year != "" {
			year = article.ItemInfo.Copyright.Year
		} else {
			// require at least a year
			return time.Time{}, ErrNoYearFound
		}
	}
	if month == "" {
		month = "1"
	}
	if day == "" {
		day = "1"
	}
	return time.Parse("2006-1-2", fmt.Sprintf("%s-%s-%s", year, month, day))
}

func (article Article) Title() string {
	if article.Head.Title != "" {
		return article.Head.Title
	}
	return article.SimpleHead.Title
}

func (article Article) Authors() []fincnext.Author {
	var authors []fincnext.Author
	for _, author := range article.Head.AuthorGroup.Author {
		authors = append(authors, fincnext.Author{
			FirstName: author.GivenName,
			LastName:  author.Surname,
			Name:      fmt.Sprintf("%s %s", author.GivenName, author.Surname),
		})
	}
	if len(authors) == 0 {
		// assume it is a simple article and try to get authors from there
		for _, author := range article.SimpleHead.AuthorGroup.Author {
			authors = append(authors, fincnext.Author{
				FirstName: author.GivenName,
				LastName:  author.Surname,
				Name:      fmt.Sprintf("%s %s", author.GivenName, author.Surname),
			})
		}
	}
	return authors
}

// Shipment is a tar export, looks like SAXC0000000000046A.tar. The tar is not
// extracted but loaded into memory at once. Issues and articles are stored in a
// map, each keyed on the PII.
type Shipment struct {
	// the full path to the file
	origin string
	// parsed dataset.xml from the tar file
	dataset Dataset
	// issues, keyed by the PII
	issues map[string]SerialIssue
	// articles, keyed by the PII
	articles map[string]Article
}

// String describes the shipment briefly.
func (s Shipment) String() string {
	return fmt.Sprintf("<Shipment origin=%s, issues=%d, articles=%d>",
		s.origin,
		len(s.dataset.DatasetContent.JournalIssue),
		len(s.dataset.DatasetContent.JournalItem))
}

// NewShipment creates a new bag of data from a given tarfile.
func NewShipment(r io.Reader) (Shipment, error) {

	var shipment = Shipment{
		origin:   fmt.Sprintf("%T", r),
		issues:   make(map[string]SerialIssue),
		articles: make(map[string]Article),
	}

	if ff, ok := r.(*os.File); ok {
		shipment.origin = ff.Name()
	}

	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return shipment, err
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, tr); err != nil {
			return shipment, err
		}

		dec := xml.NewDecoder(&buf)
		dec.Strict = false

		switch {
		case strings.HasSuffix(header.Name, "main.xml"):
			var article Article
			if err := dec.Decode(&article); err != nil {
				return shipment, err
			}
			shipment.articles[article.ItemInfo.Pii] = article
		case strings.HasSuffix(header.Name, "issue.xml"):
			var si SerialIssue
			if err := dec.Decode(&si); err != nil {
				return shipment, err
			}
			shipment.issues[si.IssueInfo.Pii] = si
		case strings.HasSuffix(header.Name, "dataset.xml"):
			var ds Dataset
			if err := dec.Decode(&ds); err != nil {
				return shipment, err
			}
			shipment.dataset = ds
		}
	}
	return shipment, nil
}

// BatchConvert converts all items for a shipment into importable objects.
func (s Shipment) BatchConvert() ([]fincnext.IntermediateSchema, error) {
	var outputs []fincnext.IntermediateSchema

	for _, ji := range s.dataset.DatasetContent.JournalIssue {
		pii := ji.JournalIssueUniqueIds.Pii
		si, ok := s.issues[pii]
		if !ok {
			log.Println(fmt.Sprintf("skipping, issue referenced %s, but not cached", pii))
			continue
		}
		for _, sec := range si.IssueBody.IssueSec {
			for _, ii := range sec.IncludeItem {
				output := fincnext.NewIntermediateSchema()

				article, ok := s.articles[ii.Pii]

				if !ok {
					log.Println(fmt.Sprintf("skipping, article referenced %s, but not cached", ii.Pii))
					continue
				}

				output.Authors = article.Authors()
				output.DOI = article.ItemInfo.Doi
				output.Format = Format
				output.Genre = Genre
				output.ISSN = []string{si.IssueInfo.Issn}
				output.Issue = si.IssueInfo.VolumeIssueNumber.IssFirst
				output.Languages = []string{"eng"}
				output.MegaCollection = Collection
				output.RecordID = fmt.Sprintf("ai-%s-%s", SourceID, base64.RawURLEncoding.EncodeToString([]byte(article.ItemInfo.Doi)))
				output.RefType = DefaultRefType
				output.SourceID = SourceID
				output.Volume = si.IssueInfo.VolumeIssueNumber.VolFirst

				output.ArticleTitle = article.Title()
				output.JournalTitle = ji.JournalIssueProperties.CollectionTitle

				output.StartPage = ii.Pages.FirstPage
				output.EndPage = ii.Pages.LastPage
				output.Pages = ii.Pages.Total()

				output.URL = []string{
					fmt.Sprintf("http://doi.org/%s", article.ItemInfo.Doi),
				}

				date, err := article.Date()
				if err != nil {
					log.Printf("%+v: %s", article.Head, err)
					continue
				}

				output.Date = date
				output.RawDate = date.Format("2006-01-02")

				var buf bytes.Buffer
				for _, abs := range article.Head.Abstract {
					buf.WriteString(sanitize.HTML(abs.Text))
				}
				output.Abstract = buf.String()

				outputs = append(outputs, *output)
			}
		}
	}
	return outputs, nil
}
