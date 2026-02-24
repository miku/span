// Package filter implements flexible ISIL attachments with expression trees[1],
// serialized as JSON. The top-level key is the label, that is to be given to a
// record. Here, this label is an ISIL. Each ISIL can specify a tree of filters.
// Intermediate nodes can be "or", "and" or "not" filters, leaf nodes contain
// filters, that are matched against records (like "collection", "source" or "issn").
//
// A filter needs to implement Apply. If the filter takes configuration
// options, it needs to implement UnmarshalJSON as well. Each filter can define
// arbitrary options, for example a HoldingsFilter can load KBART data from a single
// file or a list of urls.
//
// [1] https://en.wikipedia.org/wiki/Binary_expression_tree#Boolean_expressions
//
//
// The simplest filter is one, that says *yes* to all records:
//
//     {"DE-X": {"any": {}}}
//
// On the command line:
//
//     $ span-tag -c '{"DE-X": {"any": {}}}' < input.ldj > output.ldj
//
//
// Another slightly more complex example: Here, the ISIL "DE-14" is attached to a
// record, if the following conditions are met: There are two alternatives, each
// consisting of a conjuntion. The first says: IF "the record is from source id 55"
// AND IF "the record can be validated against one of the holding files given by
// their url", THEN "attach DE-14". The second says: IF "the record is from source
// id 49" AND "it validates against any one of the holding files given by their
// urls" AND "the record belongs to any one of the given collections", THEN
// "attach DE-14".
//
// {
//   "DE-14": {
//     "or": [
//       {
//         "and": [
//           {
//             "source": [
//               "55"
//             ]
//           },
//           {
//             "holdings": {
//               "urls": [
//                 "http://www.jstor.org/kbart/collections/asii",
//                 "http://www.jstor.org/kbart/collections/as"
//               ]
//             }
//           }
//         ]
//       },
//       {
//         "and": [
//           {
//             "source": [
//               "49"
//             ]
//           },
//           {
//             "holdings": {
//               "urls": [
//                 "https://example.com/KBART_DE14",
//                 "https://example.com/KBART_FREEJOURNALS"
//               ]
//             }
//           },
//           {
//             "collection": [
//               "Turkish Family Physicans Association (CrossRef)",
//               "Helminthological Society (CrossRef)",
//               "International Association of Physical Chemists (IAPC) (CrossRef)",
//               "The Society for Antibacterial and Antifungal Agents, Japan (CrossRef)",
//               "Fundacao CECIERJ (CrossRef)"
//             ]
//           }
//         ]
//       }
//     ]
//   }
// }
//
// If is relatively easy to add a new filter. Imagine we want to build a filter that only allows records
// that have the word "awesome" in their title.
//
// We first define a new type:
//
//     type AwesomeFilter struct{}
//
// We then implement the Apply method:
//
//     func (f *AwesomeFilter) Apply(is finc.IntermediateSchema) bool {
//         return strings.Contains(strings.ToLower(is.ArticleTitle), "awesome")
//     }
//
// That is all. We need to register the filter, so we can use it in the configuration file.
// Add an entry to filterRegistry in filter.go:
//
//     var filterRegistry = map[string]func() Filter{
//         ...
//         "awesome": func() Filter { return &AwesomeFilter{} },
//         ...
//     }
//
// We can then use the filter in the JSON configuration:
//
//     {"DE-X": {"awesome": {}}}
//
//
// Further readings: http://theory.stanford.edu/~sergei/papers/sigmod10-index.pdf
package filter
