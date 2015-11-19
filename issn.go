package span

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidISSN = errors.New("invalid ISSN")

var (
	replacer = strings.NewReplacer("-", "", " ", "")
	pattern  = regexp.MustCompile("^[0-9]{7}[0-9X]$")
)

type ISSN string

func (s ISSN) Validate() error {
	t := strings.TrimSpace(strings.ToUpper(replacer.Replace(string(s))))
	if len(t) != 8 {
		return ErrInvalidISSN
	}
	if !pattern.Match([]byte(t)) {
		return ErrInvalidISSN
	}
	return nil
}

func (s ISSN) String() string {
	t := strings.TrimSpace(strings.ToUpper(replacer.Replace(string(s))))
	if len(t) != 8 {
		return string(s)
	}
	return t[:4] + "-" + t[4:]
}
