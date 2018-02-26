package thieme

import (
	"encoding/xml"
	"fmt"

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
						Text string `xml:",chardata"`
						Lang string `xml:"lang,attr"`
						P    struct {
							Text string `xml:",chardata"` // Weitere Beobachtungen üb...
						} `xml:"p"`
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

func (record Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()
	if len(record.Metadata.Article) == 0 {
		return nil, fmt.Errorf("no article found")
	}
	article := record.Metadata.Article[0]

	output.ArticleTitle = article.Front.ArticleMeta.TitleGroup.ArticleTitle.Text
	return output, nil
}
