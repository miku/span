// Package span implements common functions.
//
//	Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//	                  The Finc Authors, http://finc.info
//	                  Martin Czygan, <martin.czygan@uni-leipzig.de>
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
package span

import (
	"encoding/base64"
	"fmt"
)

// AppVersion of span package. Commandline tools will show this on -v.
// Set at build time via: -ldflags "-X github.com/miku/span.AppVersion=..."
var AppVersion = "0.2.20"

const (
	// KeyLengthLimit was a limit imposed by the memcached protocol, which
	// was used for blob storage until Q1 2017. We switched the key-value
	// store, so this limit is somewhat obsolete.
	KeyLengthLimit = 250
)

// Skip marks records to skip.
type Skip struct {
	Reason string
}

// Error returns the reason for skipping.
func (s Skip) Error() string {
	return fmt.Sprintf("[skip] %s", s.Reason)
}

// GenFincID returns a finc.id string consisting of an arbitraty prefix (e.g.
// "ai"), source id and URL safe record id. No additional checks, sid and rid
// should not be empty.
func GenFincID(sid, rid string) string {
	return fmt.Sprintf("ai-%s-%s", sid,
		base64.RawURLEncoding.EncodeToString([]byte(rid)))
}
