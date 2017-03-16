package licensing

import "io"

// FilterFunc is a function that can match an Entry.
type FilterFunc func(Entry) bool

// Holdings abstracts from the holdings implementation.
type Holdings interface {
	io.ReaderFrom
	Filter(f FilterFunc) bool
	SerialNumberMap() map[string][]Entry
}
