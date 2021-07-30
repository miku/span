package dblp

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
)

// Article was generated 2021-07-30 13:08:22 by tir on optiplex.
type Article struct {
	XMLName  xml.Name `xml:"article"`
	Text     string   `xml:",chardata"`
	Mdate    string   `xml:"mdate,attr"`
	Key      string   `xml:"key,attr"`
	Publtype string   `xml:"publtype,attr"`
	Cdate    string   `xml:"cdate,attr"`
	Author   []struct {
		Text  string `xml:",chardata"` // Paul Kocher, Daniel Genki...
		Orcid string `xml:"orcid,attr"`
		Aux   string `xml:"aux,attr"`
	} `xml:"author"`
	Title struct {
		Text   string   `xml:",chardata"` // Spectre Attacks: Exploiti...
		Bibtex string   `xml:"bibtex,attr"`
		Tt     []string `xml:"tt"` // PROP, IIPBF, MATLAB, Repi...
	} `xml:"title"`
	Journal string `xml:"journal"` // meltdownattack.com, meltd...
	Year    string `xml:"year"`    // 2018, 2018, 1994, 1993, 1...
	Ee      []struct {
		Text string `xml:",chardata"` // https://spectreattack.com...
		Type string `xml:"type,attr"`
	} `xml:"ee"`
	Volume string `xml:"volume"` // TR-0263-08-94-165, TR-022...
	Month  string `xml:"month"`  // August, March, December, ...
	URL    string `xml:"url"`    // db/journals/gtelab/index....
	Note   []struct {
		Text string `xml:",chardata"` // This report is also avail...
		Type string `xml:"type,attr"`
	} `xml:"note"`
	Cdrom     string `xml:"cdrom"`     // SQL/X3H2-90-412.pdf, SQL/...
	Publisher string `xml:"publisher"` // IBM Germany Science Cente...
	Editor    []struct {
		Text  string `xml:",chardata"` // Thomas Wetter, Rolf Engel...
		Orcid string `xml:"orcid,attr"`
	} `xml:"editor"`
	Pages  string `xml:"pages"`  // 2947-2962, 1-27, 409-421,...
	Number string `xml:"number"` // 4, 1, 5, 3, 1, 1, 4, 1, 3...
	Cite   []struct {
		Text  string `xml:",chardata"` // journals/nca/HuangCL14a, ...
		Label string `xml:"label,attr"`
	} `xml:"cite"`
	Crossref  string `xml:"crossref"`  // journals/tcci/2011-3, jou...
	Publnr    string `xml:"publnr"`    // TR11-015, 1806.06017, 200...
	Booktitle string `xml:"booktitle"` // Logical Methods in Comput...
}

// ToIntermediateSchema converts DBLP article to intermediate schema.
func (article *Article) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	var err error
	output := finc.NewIntermediateSchema()
	for _, item := range article.Author {
		output.Authors = append(output.Authors, finc.Author{Name: item.Text})
	}
	output.Date, err = time.Parse("2006", article.Year)
	if err != nil {
		return nil, span.Skip{Reason: fmt.Sprintf("unparsable date: %s", article.Year)}
	}
	output.RawDate = output.Date.Format("2006-01-02")
	output.ArticleTitle = article.Title.Text
	output.Volume = article.Volume
	if article.Publisher != "" {
		output.Publishers = []string{article.Publisher}
	}
	output.Issue = article.Number
	return output, nil
}
