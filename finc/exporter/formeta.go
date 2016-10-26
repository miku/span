package exporter

import (
	"github.com/miku/span/encoding/formeta"
	"github.com/miku/span/finc"
)

type Formeta struct{}

func (s *Formeta) Export(is finc.IntermediateSchema, _ bool) ([]byte, error) {
	return formeta.Marshal(is)
}
