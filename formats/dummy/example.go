// Package dummy is just a minimal example.
//
// $ echo '{"title": "Sample Title"}' | ./span-import -i dummy | jq .
// {
//   "rft.atitle": "Sample Title",
//   "version": "0.9"
// }
package dummy

import "github.com/miku/span/formats/finc"

// Example record, which consists only of a title.
type Example struct {
	Title string `json:"title"`
}

// ToIntermediateSchema implements an example convertion.
func (ex Example) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()
	output.ArticleTitle = ex.Title
	return output, nil
}
