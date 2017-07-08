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
package finc

import (
	"strings"

	"github.com/miku/span/assetutil"
)

var (
	// SubjectMapping = assetutil.MustLoadStringSliceMap("assets/finc/subjects.json")
	// LanguageMap   = assetutil.MustLoadStringMap("assets/finc/iso-639-3-language.json")
	AIAccessFacet = "Electronic Resources"

	FormatDe105  = assetutil.MustLoadStringMap("assets/finc/formats/de105.json")
	FormatDe14   = assetutil.MustLoadStringMap("assets/finc/formats/de14.json")
	FormatDe15   = assetutil.MustLoadStringMap("assets/finc/formats/de15.json")
	FormatDe520  = assetutil.MustLoadStringMap("assets/finc/formats/de520.json")
	FormatDe540  = assetutil.MustLoadStringMap("assets/finc/formats/de540.json")
	FormatDeCh1  = assetutil.MustLoadStringMap("assets/finc/formats/dech1.json")
	FormatDed117 = assetutil.MustLoadStringMap("assets/finc/formats/ded117.json")
	FormatDeGla1 = assetutil.MustLoadStringMap("assets/finc/formats/degla1.json")
	FormatDel152 = assetutil.MustLoadStringMap("assets/finc/formats/del152.json")
	FormatDel189 = assetutil.MustLoadStringMap("assets/finc/formats/del189.json")
	FormatDeZi4  = assetutil.MustLoadStringMap("assets/finc/formats/dezi4.json")
	FormatDeZwi2 = assetutil.MustLoadStringMap("assets/finc/formats/dezwi2.json")
	FormatNrw    = assetutil.MustLoadStringMap("assets/finc/formats/nrw.json")
)

// AuthorReplacer is a special cleaner for author names.
var AuthorReplacer = strings.NewReplacer(
	"anonym", "",
	"Anonymous", "",
	"keine Angabe", "",
	"No authorship indicated", "",
	"Not Available, Not Available", "",
	"Author Index", "",
	"AUTHOR Index", "",
	"AUTHOR INDEX", "")
