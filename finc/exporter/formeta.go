package exporter

import (
	"fmt"

	"github.com/miku/span/finc"
)

type Formeta struct{}

func (s *Formeta) Export(is finc.IntermediateSchema, _ bool) ([]byte, error) {
	return []byte{}, fmt.Errorf("not yet implemented")
}
