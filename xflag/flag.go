// Package xflag add an additional flag type Array for repeated string flags.
//
//   var f xflag.Array
//   flag.Var(&f, "r", "some repeatable flag")
//
//   flag.Parse()
//   for _, v := range f { ... }
//
package xflag

import "strings"

// ArrayFlags allows to store lists of flag values.
type Array []string

// String representation.
func (f *Array) String() string {
	return strings.Join(*f, ", ")
}

// Set appends a value.
func (f *Array) Set(value string) error {
	*f = append(*f, value)
	return nil
}
