package span

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

type Tagged struct {
	Tag   string
	Value string
}

type taggedFlag struct {
	Tagged
}

func (v *taggedFlag) String() string {
	return fmt.Sprintf("%s:%s", v.Tag, v.Value)
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

func main() {
	t := TaggedFlag("x", Tagged{}, "tag path in the form: TAG:PATH")
	flag.Parse()
	log.Println(t)
}
