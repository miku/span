// Package xflag add an additional flag type Array for repeated string flags.
//
//   var f xflag.Array
//   flag.Var(&f, "r", "some repeatable flag")
//
//   flag.Parse()                // $ command -r a -r b -r c
//   for _, v := range f { ... } // []string{"a", "b", "c"}
//
package xflag

import (
	"fmt"
	"strings"
)

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

// UserPassword allows to pass in user:password in flags.
type UserPassword struct {
	User     string
	Password string
}

func (u *UserPassword) String() string {
	return fmt.Sprintf("%s:%s", u.User, u.Password)
}

func (u *UserPassword) Set(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return fmt.Errorf("user:password required")
	}
	u.User = parts[0]
	u.Password = parts[1]
	return nil
}
