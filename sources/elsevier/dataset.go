package elsevier

type Dataset struct {
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
					Aid  string `xml:"aid"`
				} `xml:"jid-aid"`
			} `xml:"journal-item-unique-ids"`
			JournalItemProperties struct {
				PIT                   string `xml:"pit"`
				ProductionType        string `xml:"production-type"`
				OnlinePublicationDate string `xml:"online-publication-date"`
			} `xml:"journal-item-properties"`
			FilesInfo struct {
				ML struct {
					Pathname   string `xml:"pathname"`
					Filesize   string `xml:"filesize"`
					Purpose    string `xml:"purpose"`
					DTDVersion string `xml:"dtd-version"`
					Weight     string `xml:"weight"`
				} `xml:"ml"`
			} `xml:"files-info"`
		} `xml:"journal-item"`
	} `xml:"dataset-content"`
}
