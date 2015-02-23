package jats

import "encoding/xml"

type ISSN struct {
	ID   string `xml:",chardata"`
	Type string `xml:"pub-type,attr"`
}

type AbbreviatedTitle struct {
	Title string `xml:",chardata"`
	Type  string `xml:"abbrev-type,attr"`
}

type JournalTitleGroup struct {
	AbbreviatedTitle AbbreviatedTitle `xml:"abbrev-journal-title"`
}

type JournalID struct {
	ID   string `xml:",chardata"`
	Type string `xml:"journal-id-type,attr"`
}

type JournalMeta struct {
	IDList     []JournalID       `xml:"journal-id"`
	ISSN       []ISSN            `xml:"issn"`
	TitleGroup JournalTitleGroup `xml:"journal-title-group"`
}

type ArticleID struct {
	ID   string `xml:",chardata"`
	Type string `xml:"pub-id-type,attr"`
}

type ArticleMeta struct {
	IDList   []ArticleID `xml:"article-id"`
	Title    string      `xml:"title-group>article-title>text"`
	Subtitle string      `xml:"title-group>subtitle>text"`
}

type Meta struct {
	Journal JournalMeta `xml:"journal-meta"`
	Article ArticleMeta `xml:"article-meta"`
}

type Body struct {
}

type Document struct {
	XMLName xml.Name `xml:"article"`
	Meta    Meta     `xml:"front"`
	Body    Body     `xml:"body"`
}
