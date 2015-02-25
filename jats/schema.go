package jats

import "encoding/xml"

// Article mirrors a JATS article element. Some elements, such as
// article categories are not implmented yet.
type Article struct {
	XMLName xml.Name `xml:"article"`
	Front   struct {
		JournalMeta struct {
			JournalID struct {
				Type string `xml:"journal-id-type,attr"`
				ID   string `xml:",chardata"`
			} `xml:"journal-id"`
			ISSN []struct {
				Type  string `xml:"pub-type,attr"`
				Value string `xml:",chardata"`
			} `xml:"issn"`
			Publisher struct {
				Name struct {
					Value string `xml:",chardata"`
				} `xml:"publisher-name"`
			} `xml:"publisher"`
		} `xml:"journal-meta"`
		ArticleMeta struct {
			ArticleID []struct {
				Type  string `xml:"pub-id-type,attr"`
				Value string `xml:",chardata"`
			} `xml:"article-id"`
			TitleGroup struct {
				Title struct {
					Value string `xml:",chardata"`
				} `xml:"article-title"`
				Subtitle struct {
					Value string `xml:",chardata"`
				} `xml:"subtitle"`
			} `xml:"title-group"`
			ContribGroup struct {
				Contrib []struct {
					Type string `xml:"contrib-type,attr"`
					Name struct {
						Surname struct {
							Value string `xml:",chardata"`
						} `xml:"surname"`
						GivenNames struct {
							Value string `xml:",chardata"`
						} `xml:"given-names"`
					} `xml:"name"`
				} `xml:"contrib"`
			} `xml:"contrib-group"`
			PubDate struct {
				Type  string `xml:"pub-type,attr"`
				Month struct {
					Value string `xml:",chardata"`
				} `xml:"month"`
				Year struct {
					Value string `xml:",chardata"`
				} `xml:"year"`
				Day struct {
					Value string `xml:",chardata"`
				} `xml:"day"`
			} `xml:"pub-date"`
			Volume struct {
				Value string `xml:",chardata"`
			} `xml:"volume"`
			Issue struct {
				Value string `xml:",chardata"`
			} `xml:"issue"`
			FirstPage struct {
				Value string `xml:",chardata"`
			} `xml:"fpage"`
			LastPage struct {
				Value string `xml:",chardata"`
			} `xml:"lpage"`
		} `xml:"article-meta"`
	} `xml:"front"`
}
