package sift

// AnyFilter allows anything.
type AnyFilter struct {
	AnyFilter struct{} `json:"any"`
}

// Apply just returns true.
func (AnyFilter) Apply(_ interface{}) bool { return true }

// NoneFilter blocks everything.
type NoneFilter struct {
	None struct{} `json:"none"`
}

// Apply just returns false.
func (NoneFilter) Apply(_ interface{}) bool { return false }
