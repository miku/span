package elsevier

import "encoding/xml"

type Dataset struct {
	XMLName          xml.Name `xml:"dataset"`
	DatasetUniqueIDs struct {
		ProfileCode      string `xml:"profile-code"`
		ProfileDatasetId string `xml:"profile-dataset-id"`
		Timestamp        string `xml:"timestamp"`
	} `xml:"dataset-unique-ids"`
	DatasetProperties struct {
		DatasetAction     string `xml:"dataset-action"`
		ProductionProcess string `xml:"production-process"`
	}
	DatasetContent struct {
		// Journal issues
		JournalIssue []struct {
			Version struct {
				VersionNumber string `xml:"version-number"`
				Stage         string `xml:"H300"`
			} `xml:"version"`
			JournalIssueUniqueIDs struct {
				PII string `xml:"pii"`
			} `xml:"journal-issue-unique-ids"`
			JournalIssueProperties struct {
				JID               string `xml:"jid"`
				ISSN              string `xml:"issn"`
				VolumeIssueNumber struct {
					VolFirst string `xml:"vol-first"`
					Suppl    string `xml:"suppl"`
				} `xml:"volume-issue-number"`
				CollectionTitle string `xml:"collection-title"`
			} `xml:"journal-issue-properties"`
			FilesInfo struct {
				ML struct {
					Pathname   string `xml:"pathname"`
					Filesize   string `xml:"filesize"`
					Purpose    string `xml:"purpose"`
					DTDVersion string `xml:"dtd-version"`
				} `xml:"ml"`
			} `xml:"files-info"`
		} `xml:"journal-issue"`

		// Journal items
		JournalItem []struct {
			Version struct {
				VersionNumber string `xml:"version-number"`
				Stage         string `xml:"H300"`
			} `xml:"version"`
			JournalItemUniqueIDs struct {
				PII    string `xml:"pii"`
				DOI    string `xml:"doi"`
				JIDAid struct {
					PII  string `xml:"jid"`
					ISSN string `xml:"issn"`
					AID  string `xml:"aid"`
				} `xml:"jid-aid"`
			} `xml:"journal-item-unique-ids"`
			JournalItemProperties struct {
				PIT                   string `xml:"pit"`
				ProductionType        string `xml:"production-type"`
				OnlinePublicationDate string `xml:"online-publication-date"`
				SponsoredAccess       struct {
					Type string `xml:"type"`
				} `xml:"sponsored-access"`
				FundingBodyId string `xml:"funding-body-id"`
			} `xml:"journal-item-properties"`
			FilesInfo struct {
				ML struct {
					Pathname   string `xml:"pathname"`
					Filesize   string `xml:"filesize"`
					Purpose    string `xml:"purpose"`
					DTDVersion string `xml:"dtd-version"`
					Weight     string `xml:"weight"`
					Assets     []struct {
						Pathname string `xml:"pathname"`
						Filesize string `xml:"filesize"`
						Type     string `xml:"type"`
					} `xml:"assets"`
				} `xml:"ml"`
			} `xml:"files-info"`
		} `xml:"journal-item"`
	} `xml:"dataset-content"`
}

type SerialIssue struct {
	XMLName   xml.Name `xml:"serial-issue"`
	IssueInfo struct {
		PII               string
		JID               string
		VolumeIssueNumber struct {
			VolFirst string `xml:"vol-first"`
			IssFirst string `xml:"iss-first"`
			IssLast  string `xml:"iss-last"`
		} `xml:"volume-issue-number"`
	} `xml:"issue-info"`
	IssueData struct {
		CoverDate struct {
			StartDate string `xml:"start-date"`
			EndDate   string `xml:"end-date"`
		} `xml:"cover-date"`
		Pages struct {
			FirstPage string `xml:"first-page"`
			LastPage  string `xml:"last-page"`
		} `xml:"pages"`
		CoverImage struct {
			Figure struct {
				Link struct {
					Locator string `xml:"locator,attr"`
				} `xml:"link"`
			} `xml:"figure"`
		} `xml:"cover-image"`
		TitleEditorsGroup struct {
			Title   string `xml:"title"`
			Editors struct {
				AuthorGroup []struct {
					Author struct {
						GivenName string `xml:"given-name"`
						Surname   string `xml:"surname"`
					} `xml:"author"`
				} `xml:"author-group"`
			} `xml:"editors"`
		}
	} `xml:"issue-data"`

	IssueBody struct {
		IncludeItems []struct {
			PII   string `xml:"pii"`
			DOI   string `xml:"doi"`
			Pages struct {
				FirstPage string `xml:"first-page"`
				LastPage  string `xml:"last-page"`
			} `xml:"pages"`
		} `xml:"include-item"`
		IssueSections []struct {
			SectionTitle string `xml:"section-title"`
			IncludeItem  struct {
				PII   string `xml:"pii"`
				DOI   string `xml:"doi"`
				Pages struct {
					FirstPage string `xml:"first-page"`
					LastPage  string `xml:"last-page"`
				} `xml:"pages"`
			}
		} `xml:"issue-sec"`
	} `xml:"issue-body"`
}

type SimpleArticle struct {
	XMLName  xml.Name `xml:"simple-article"`
	ItemInfo struct {
		JID       string `xml:"jid"`
		AID       string `xml:"aid"`
		PII       string `xml:"ce:pii"`
		DOI       string `xml:"ce:doi"`
		Copyright struct {
			Type string `xml:"type,attr"`
			Year string `xml:"year,attr"`
		} `xml:"ce:copyright"`
	} `xml:"ce:item-info"`

	SimpleHead struct {
		DocHead struct {
			TextFn string `xml:"ce:textfn"`
		} `xml:"ce:dochead"`
		Title       string `xml:"ce:title"`
		AuthorGroup []struct {
			Author struct {
				ID                string `xml:"id,attr"`
				GivenName         string `xml:"ce:given-name"`
				Surname           string `xml:"ce:surname"`
				ElectronicAddress struct {
					Type  string `xml:"type,attr"`
					Value string `xml:",chardata"`
				} `xml:"ce:e-address"`
				Affiliation struct {
					TextFn string `xml:"ce:textfn"`
				} `xml:"ce:affilitaion"`
				Crossref struct {
					RefID string `xml:"refid,attr"`
					Sup   struct {
						Loc string `xml:"loc,attr"`
					}
				} `xml:"ce:cross-ref"`
				Correspondence struct {
					ID    string `xml:"id,attr"`
					Label string `xml:"ce:label"`
					Text  string `xml:"ce:text"`
				} `xml:"ce:correspondence"`
			} `xml:"ce:author"`
		} `xml:"ce:author-group"`
	} `xml:"ce:simple-head"`
}
