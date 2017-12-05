package imslp

import (
	"encoding/base64"
	"fmt"

	"github.com/beevik/etree"
	"github.com/miku/span/formats/finc"
)

// SourceIdentifier of IMSLP.
const SourceIdentifier = "15"

// Data is just the raw bytes.
type Data []byte

// UnmarshalText unmarshals textual representation of itself.
func (data *Data) UnmarshalText(text []byte) error {
	*data = append(*data, text...)
	return nil
}

// ToIntermediateSchema converts record to intermediate schema.
func (data *Data) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes([]byte(*data)); err != nil {
		return nil, err
	}
	output := finc.NewIntermediateSchema()
	output.SourceID = SourceIdentifier
	for _, t := range doc.FindElements("//var/recordId") {
		encoded := base64.RawURLEncoding.EncodeToString([]byte(t.Text()))
		output.ID = fmt.Sprintf("ai-%s-%s", SourceIdentifier, encoded)
		output.RecordID = t.Text()
	}
	for _, t := range doc.FindElements("//var[@name='Work Title']/string") {
		output.ArticleTitle = t.Text()
	}
	for _, t := range doc.FindElements("//var[@name='permlink']/string") {
		output.URL = append(output.URL, t.Text())
	}
	for _, t := range doc.FindElements("//var[@name='composer']/string") {
		output.Authors = append(output.Authors, finc.Author{Name: t.Text()})
	}
	return output, nil
}
