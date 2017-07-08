package finc

import (
	"github.com/miku/span/encoding/formeta"
)

type Formeta struct{}

func (s *Formeta) Export(is IntermediateSchema, _ bool) ([]byte, error) {
	return formeta.Marshal(is)
}
