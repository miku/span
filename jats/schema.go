package jats

type ISSN struct {
	ID   string `xml:",chardata"`
	Type string `xml:"pub-type,attr"`
}

type JournalMeta struct {
	ID         string   `xml:"journal-id"`
	Type       string   `xml:"journal-id-type,attr"`
	ISSN       []ISSN   `xml:"issn"`
	Title      string   `xml:"journal-title-group>abbrev-journal-title,chardata"`
	TitleType  string   `xml:"journal-title-group>abbrev-journal-title>abbrev-type,attr"`
	Publishers []string `xml:"publisher>publisher-name,chardata"`
}

type ArticleID struct {
	ID   string `xml:"article-id,chardata"`
	Type string `xml:"pib-id-type,attr"`
}

type ArticleMeta struct {
	IDList   []ArticleID `xml:"article-id"`
	Title    string      `xml:"title-group>article-title,chardata"`
	Subtitle string      `xml:"title-group>subtitle,chardata"`
}

type Meta struct {
	Journal JournalMeta `xml:"journal-meta"`
	Article ArticleMeta `xml:"article-meta"`
}

type Body struct {
}

type Document struct {
	Meta Meta `xml:"article>front"`
	Body Body `xml:"article>body"`
}
