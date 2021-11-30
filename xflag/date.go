package xflag

import (
	"time"

	"github.com/araddon/dateparse"
)

// Date can be used to parse command line args into dates.
type Date struct {
	time.Time
}

// String returns a formatted date.
func (d *Date) String() string {
	return d.Format("2006-01-02")
}

// Set parses a value into a date, relatively flexible due to
// araddon/dateparse, 2014-04-26 will work, but oct. 7, 1970, too.
func (d *Date) Set(value string) error {
	t, err := dateparse.ParseStrict(value)
	if err != nil {
		return err
	}
	*d = Date{t}
	return nil
}
