package licensing

// FilterFunc is a function that can match an Entry.
type FilterFunc func(Entry) bool
