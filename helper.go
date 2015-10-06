//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                 by The Finc Authors, http://finc.info
//                 by Martin Czygan, <martin.czygan@uni-leipzig.de>
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
	"html"
	"io"
	"log"
	"strings"

	"github.com/rainycape/cld2"
	"golang.org/x/text/language"
)

// UnescapeTrim unescapes HTML character references and trims the space of a given string.
func UnescapeTrim(s string) string {
	return strings.TrimSpace(html.UnescapeString(s))
}

// ByteSink is a fan in writer for a byte channel.
// A newline is appended after each object.
func ByteSink(w io.Writer, out chan []byte, done chan bool) {
	f := bufio.NewWriter(w)
	for b := range out {
		if _, err := f.Write(b[:]); err != nil {
			log.Fatal(err)
		}
		if _, err := f.Write([]byte("\n")); err != nil {
			log.Fatal(err)
		}
	}
	if err := f.Flush(); err != nil {
		log.Fatal(err)
	}
	done <- true
}

// DetectLang3 returns the best guess 3-letter language code for a given text.
func DetectLang3(text string) (string, error) {
	c := cld2.Detect(text)
	b, err := language.ParseBase(c)
	if err != nil {
		return "", err
	}
	return b.ISO3(), nil
}
