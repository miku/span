package exporter

import (
	"strings"

	"github.com/miku/span/assetutil"
)

var (
	SubjectMapping = assetutil.MustLoadStringSliceMap("assets/finc/subjects.json")
	LanguageMap    = assetutil.MustLoadStringMap("assets/finc/iso-639-3-language.json")
	AIAccessFacet  = "Electronic Resources"

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
)

// AuthorReplacer is a special cleaner for author names.
var AuthorReplacer = strings.NewReplacer(
	"- -", "",
	"anonym", "",
	"Anonymous", "",
	"EB", "",
	"keine Angabe", "",
	"mg", "",
	"MM", "",
	"mm", "",
	"No authorship indicated", "",
	"Not Available, Not Available", "",
	"O.V.", "",
	"ps", "",
	"rb", "",
	"et al", "")
