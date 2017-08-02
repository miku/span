package sift

// Collectioner returns a number of collections.
type Collectioner interface {
	Collections() []string
}

// SerialNumberer returns a list of ISSN as strings.
type SerialNumberer interface {
	SerialNumber() []string
}

// DocumentObjectIdentifier returns a single document object identifier as string.
type DocumentObjectIdentifier interface {
	DOI() string
}

// Subjecter returns a number of subjects.
type Subjecter interface {
	Subjects() []string
}

// PublicationDater returns a publication date string in ISO format: 2017-07-15.
type PublicationDater interface {
	PublicationDate() string
}

// Volumer returns a volume number, preferably without decoration.
type Volumer interface {
	Volume() string
}

// Issuer returns a issue, preferably without decoration.
type Issuer interface {
	Issue() string
}

// Applier returns a boolean, given a value. This abstracts away the process of
// the actual decision making. The implementations will need to type assert
// certain interfaces to access values from the interfaces.
type Applier interface {
	Apply(interface{}) bool
}

// Any allows anything.
type Any struct {
	Any struct{} `json:"any"`
}

// Apply just returns true.
func (a Any) Apply(_ interface{}) bool { return true }

// None blocks everything.
type None struct {
	None struct{} `json:"none"`
}

// Apply just returns false.
func (a None) Apply(_ interface{}) bool { return false }

// Collection allows only values belonging to a given collection.
type Collection struct {
	Fallback    bool     `json:"fallback"`
	Collections []string `json:"collections"`
}

// Apply checks, if a value belongs to a given collection.
func (c Collection) Apply(v interface{}) bool {
	if w, ok := v.(Collectioner); ok {
		for _, a := range w.Collections() {
			for _, b := range c.Collections {
				if a == b {
					return true
				}
			}
		}
	}
	return c.Fallback
}
