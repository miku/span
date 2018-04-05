package thieme

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
)

// Record was generated 2018-02-15 13:37:41 by tir on hayiti.
type Record struct {
	XMLName xml.Name `xml:"Record"`
	Text    string   `xml:",chardata"`
	Header  struct {
		Text       string `xml:",chardata"`
		Status     string `xml:"status,attr"`
		Identifier struct {
			Text string `xml:",chardata"` // 10.1055-s-0029-1195170, 1...
		} `xml:"identifier"`
		Datestamp struct {
			Text string `xml:",chardata"` // 2013-03-13T06:02:46Z, 201...
		} `xml:"datestamp"`
		SetSpec struct {
			Text string `xml:",chardata"` // journalarticles, journala...
		} `xml:"setSpec"`
	} `xml:"header"`
	Metadata struct {
		Text    string `xml:",chardata"`
		Article []struct {
			Text                      string `xml:",chardata"`
			Xsi                       string `xml:"xsi,attr"`
			NoNamespaceSchemaLocation string `xml:"noNamespaceSchemaLocation,attr"`
			Lang                      string `xml:"lang,attr"`
			ArticleType               string `xml:"article-type,attr"`
			Front                     struct {
				Text        string `xml:",chardata"`
				JournalMeta struct {
					Text      string `xml:",chardata"`
					JournalID struct {
						Text string `xml:",chardata"`
					} `xml:"journal-id"`
					JournalTitleGroup struct {
						Text         string `xml:",chardata"`
						JournalTitle struct {
							Text string `xml:",chardata"` // Dtsch med Wochenschr, Dts...
						} `xml:"journal-title"`
					} `xml:"journal-title-group"`
					ISSN []struct {
						Text    string `xml:",chardata"` // 0012-0472, 1439-4413, 001...
						PubType string `xml:"pub-type,attr"`
					} `xml:"issn"`
					Publisher struct {
						Text          string `xml:",chardata"`
						PublisherName struct {
							Text string `xml:",chardata"` // Georg Thieme Verlag Stutt...
						} `xml:"publisher-name"`
					} `xml:"publisher"`
				} `xml:"journal-meta"`
				ArticleMeta struct {
					Text      string `xml:",chardata"`
					ArticleID struct {
						Text      string `xml:",chardata"` // 10.1055/s-0029-1195170, 1...
						PubIDType string `xml:"pub-id-type,attr"`
					} `xml:"article-id"`
					ArticleCategories struct {
						Text      string `xml:",chardata"`
						SubjGroup struct {
							Text    string `xml:",chardata"`
							Subject struct {
								Text string `xml:",chardata"` // Feuilleton, Medicinal - B...
							} `xml:"subject"`
						} `xml:"subj-group"`
					} `xml:"article-categories"`
					TitleGroup struct {
						Text         string `xml:",chardata"`
						ArticleTitle struct {
							Text string `xml:",chardata"` // Weitere Beobachtungen üb...
							Lang string `xml:"lang,attr"`
						} `xml:"article-title"`
						TransTitleGroup struct {
							Text       string `xml:",chardata"`
							Lang       string `xml:"lang,attr"`
							TransTitle struct {
								Text string `xml:",chardata"` // Teonanacatl and Psilocybi...
								Lang string `xml:"lang,attr"`
							} `xml:"trans-title"`
						} `xml:"trans-title-group"`
					} `xml:"title-group"`
					ContribGroup struct {
						Text    string `xml:",chardata"`
						Contrib []struct {
							Text string `xml:",chardata"`
							Name struct {
								Text    string `xml:",chardata"`
								Surname struct {
									Text string `xml:",chardata"` // Riess, Freyer, Riess, Auf...
								} `xml:"surname"`
								GivenNames struct {
									Text string `xml:",chardata"` // L., T., L., E., E., v., C...
								} `xml:"given-names"`
								Suffix struct {
									Text string `xml:",chardata"` // Sir, Sir, Sir, Sir, Sir, ...
								} `xml:"suffix"`
							} `xml:"name"`
							Aff struct {
								Text        string `xml:",chardata"`
								Institution struct {
									Text string `xml:",chardata"` // I. Aus der inneren Abthei...
								} `xml:"institution"`
							} `xml:"aff"`
							Collab struct {
								Text string `xml:",chardata"` // for the Eunice Kennedy Sh...
							} `xml:"collab"`
						} `xml:"contrib"`
					} `xml:"contrib-group"`
					PubDate struct {
						Text    string `xml:",chardata"`
						PubType string `xml:"pub-type,attr"`
						Month   struct {
							Text string `xml:",chardata"` // 12, 12, 12, 12, 12, 12, 1...
						} `xml:"month"`
						Year struct {
							Text string `xml:",chardata"` // 1879, 1879, 1879, 1879, 1...
						} `xml:"year"`
						Day struct {
							Text string `xml:",chardata"` // 31, 31, 31, 31, 31, 31, 3...
						} `xml:"day"`
					} `xml:"pub-date"`
					Volume struct {
						Text string `xml:",chardata"` // 5, 5, 5, 5, 5, 5, 5, 5, 5...
					} `xml:"volume"`
					Issue struct {
						Text string `xml:",chardata"` // 52, 52, 52, 52, 52, 52, 5...
					} `xml:"issue"`
					Fpage struct {
						Text string `xml:",chardata"` // 663, 667, 667, 669, 674, ...
					} `xml:"fpage"`
					Lpage struct {
						Text string `xml:",chardata"` // 667, 667, 669, 669, 674, ...
					} `xml:"lpage"`
					Abstract struct {
						Text string `xml:",innerxml"`
						Lang string `xml:"lang,attr"`
					} `xml:"abstract"`
					TransAbstract struct {
						Text string `xml:",chardata"`
						Lang string `xml:"lang,attr"`
						P    struct {
							Text string `xml:",chardata"` // Die zweite von Tschernogu...
						} `xml:"p"`
					} `xml:"trans-abstract"`
					KwdGroup []struct {
						Text string `xml:",chardata"`
						Lang string `xml:"lang,attr"`
						Kwd  []struct {
							Text string `xml:",chardata"` // Intra-operative vascular ...
						} `xml:"kwd"`
					} `xml:"kwd-group"`
					Supplement struct {
						Text string `xml:",chardata"` // S 01, S 01, S 01, S 01, S...
					} `xml:"supplement"`
				} `xml:"article-meta"`
			} `xml:"front"`
		} `xml:"article"`
	} `xml:"metadata"`
	About struct {
		Text string `xml:",chardata"`
	} `xml:"about"`
}

// Date returns the parsed publishing date.
func (record Record) Date() (time.Time, error) {
	if len(record.Metadata.Article) == 0 {
		return time.Time{}, fmt.Errorf("empty record")
	}
	article := record.Metadata.Article[0]
	pd := article.Front.ArticleMeta.PubDate

	if pd.Month.Text == "0" {
		pd.Month.Text = "01"
	}

	if pd.Day.Text == "0" {
		pd.Day.Text = "01"
	}

	if pd.Year.Text != "" && pd.Month.Text != "" && pd.Day.Text != "" {
		s := fmt.Sprintf("%s-%s-%s", leftPad(pd.Year.Text, "0", 4),
			leftPad(pd.Month.Text, "0", 2),
			leftPad(pd.Day.Text, "0", 2))
		return time.Parse("2006-01-02", s)
	}
	if pd.Year.Text != "" && pd.Month.Text != "" {
		s := fmt.Sprintf("%s-%s-01", leftPad(pd.Year.Text, "0", 4), leftPad(pd.Month.Text, "0", 2))
		return time.Parse("2006-01-02", s)
	}
	if pd.Year.Text != "" {
		s := fmt.Sprintf("%s-01-01", leftPad(pd.Year.Text, "0", 4))
		return time.Parse("2006-01-02", s)
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

// ToIntermediateSchema converts a single record.
func (record Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()
	if len(record.Metadata.Article) == 0 {
		return nil, span.Skip{Reason: "no article found"}
	}
	article := record.Metadata.Article[0]

	date, err := record.Date()
	if err != nil {
		return output, span.Skip{Reason: err.Error()}
	}
	output.Date = date
	output.RawDate = date.Format("2006-01-02")

	output.SourceID = SourceID
	output.Format = Format
	output.MegaCollections = []string{Collection}
	output.Genre = Genre
	output.RefType = DefaultRefType

	output.JournalTitle = article.Front.JournalMeta.JournalTitleGroup.JournalTitle.Text
	output.ArticleTitle = article.Front.ArticleMeta.TitleGroup.ArticleTitle.Text
	output.StartPage = article.Front.ArticleMeta.Fpage.Text
	output.EndPage = article.Front.ArticleMeta.Lpage.Text
	output.Volume = article.Front.ArticleMeta.Volume.Text
	output.Issue = article.Front.ArticleMeta.Issue.Text

	output.Abstract = sanitize.HTML(article.Front.ArticleMeta.Abstract.Text)
	output.Publishers = append(output.Publishers, article.Front.JournalMeta.Publisher.PublisherName.Text)

	for _, issn := range article.Front.JournalMeta.ISSN {
		switch issn.PubType {
		case "print":
			output.ISSN = append(output.ISSN, issn.Text)
		case "e-issn":
			output.EISSN = append(output.EISSN, issn.Text)
		default:
			return output, fmt.Errorf("unhandled issn type: %s", issn.PubType)
		}
	}

	if article.Front.ArticleMeta.ArticleID.PubIDType == "doi" {
		output.DOI = article.Front.ArticleMeta.ArticleID.Text
	} else {
		return output, fmt.Errorf("unknown id type: %s", article.Front.ArticleMeta.ArticleID.PubIDType)
	}

	output.RecordID = output.DOI
	output.ID = fmt.Sprintf("ai-60-%s", base64.RawURLEncoding.EncodeToString([]byte(output.DOI)))

	var authors []finc.Author
	for _, contrib := range article.Front.ArticleMeta.ContribGroup.Contrib {
		authors = append(authors, finc.Author{
			FirstName: contrib.Name.GivenNames.Text,
			LastName:  contrib.Name.Surname.Text,
		})
	}
	output.Authors = authors

	subject := strings.TrimSpace(article.Front.ArticleMeta.ArticleCategories.SubjGroup.Subject.Text)
	if subject != "" {
		output.Subjects = append(output.Subjects, subject)
	}

	return output, nil
}
