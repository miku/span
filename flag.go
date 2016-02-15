package span

import (
	"flag"
	"fmt"
	"strings"
)

type Tagged struct {
	Tag   string
	Value string
}

func (t *Tagged) String() string {
	return fmt.Sprintf("%s:%s", t.Tag, t.Value)
}

type taggedFlag struct {
	Tagged
}

func (v *taggedFlag) String() string {
	return v.Tagged.String()
}

func (v *taggedFlag) Set(s string) error {
	var parts = strings.Split(s, ":")
	if len(parts) != 2 {
		return fmt.Errorf("format must be [TAG]:[PATH]")
	}
	v.Tagged.Tag, v.Tagged.Value = parts[0], parts[1]
	return nil
}

func TaggedFlag(name string, value Tagged, usage string) *Tagged {
	f := taggedFlag{value}
	flag.CommandLine.Var(&f, name, usage)
	return &f.Tagged
}

// TagSlice collects a number of tags.
type TagSlice []Tagged

func (s *TagSlice) String() string {
	var ss []string
	for _, tag := range *s {
		ss = append(ss, tag.String())
	}
	return strings.Join(ss, ", ")
}

func (s *TagSlice) Set(value string) error {
	tag := taggedFlag{}
	if err := tag.Set(value); err != nil {
		return err
	}
	*s = append(*s, tag.Tagged)
	return nil
}
