package span

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidISSN = errors.New("invalid ISSN")

var (
	replacer = strings.NewReplacer("-", "")
	pattern  = regexp.MustCompile("^[0-9]{7}[0-9X]$")
)

type ISSN struct {
	issn string
}

func NewISSN(s string) (ISSN, error) {
	t := strings.TrimSpace(strings.ToUpper(replacer.Replace(s)))
	if len(t) != 8 {
		return ISSN{}, ErrInvalidISSN
	}
	if !pattern.Match([]byte(t)) {
		return ISSN{}, ErrInvalidISSN
	}
	return ISSN{t[:4] + "-" + t[4:]}, nil
}

func (s ISSN) String() string {
	return s.issn
}
