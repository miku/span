package span

import (
	"fmt"
	"io"

	"github.com/miku/span/finc"
)

const (
	// AppVersion of span package. Commandline tools will show this on -v.
	AppVersion = "0.1.41"
	// KeyLengthLimit is a limit imposed by memcached protocol, which is used
	// for blob storage as of June 2015. If we change the key value store,
	// this limit might become obsolete.
	KeyLengthLimit = 250
)

type Skip struct {
	Reason string
}

func (s Skip) Error() string {
	return fmt.Sprintf("[skip] %s", s.Reason)
}

// Batcher groups strings together for batched processing.
// It is more effective to send one batch over a channel than many strings.
type Batcher struct {
	Items []interface{}
	Apply func(interface{}) (Importer, error)
}

// Importer objects can be converted into an intermediate schema.
type Importer interface {
	ToIntermediateSchema() (*finc.IntermediateSchema, error)
}

// Source can emit records given a reader. What is actually returned is decided
// by the source, e.g. it may return Importer or Batcher object.
// Dealing with the various types is responsibility of the call site.
// Channel will block on slow consumers and will not drop objects.
type Source interface {
	Iterate(io.Reader) (<-chan interface{}, error)
}
