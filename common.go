// Package span implements common functions.
//
//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                    The Finc Authors, http://finc.info
//                    Martin Czygan, <martin.czygan@uni-leipzig.de>
//
// This file is part of some open source application.
//
// Some open source application is free software: you can redistribute
// it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation, either
// version 3 of the License, or (at your option) any later version.
//
// Some open source application is distributed in the hope that it will
// be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
//
// @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
//
package span

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
)

const (
	// AppVersion of span package. Commandline tools will show this on -v.
	AppVersion = "0.1.245"
	// KeyLengthLimit was a limit imposed by the memcached protocol, which
	// was used for blob storage until Q1 2017. We switched the key-value
	// store, so this limit is somewhat obsolete.
	KeyLengthLimit = 250
)

// ISSNPattern is a regular expression matching standard ISSN.
var ISSNPattern = regexp.MustCompile(`[0-9]{4,4}-[0-9]{3,3}[0-9X]`)

// Skip marks records to skip.
type Skip struct {
	Reason string
}

// Error returns the reason for skipping.
func (s Skip) Error() string {
	return fmt.Sprintf("SKIP %s", s.Reason)
}

// UnescapeTrim unescapes HTML character references and trims the space of a given string.
func UnescapeTrim(s string) string {
	return strings.TrimSpace(html.UnescapeString(s))
}

// LoadSet reads the content of from a reader and creates a set from each line.
func LoadSet(r io.Reader, m map[string]struct{}) error {
	br := bufio.NewReader(r)
	for {
		v, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		m[strings.TrimSpace(v)] = struct{}{}
	}
	return nil
}

// UserHomeDir returns the home directory of the user.
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
