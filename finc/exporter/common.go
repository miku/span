package exporter

import "github.com/miku/span/assetutil"

var (
	SubjectMapping = assetutil.MustLoadStringSliceMap("assets/finc/subjects.json")
	LanguageMap    = assetutil.MustLoadStringMap("assets/finc/iso-639-3-language.json")
	AIAccessFacet  = "Electronic Resources"

	FormatDe15   = assetutil.MustLoadStringMap("assets/finc/formats/de15.json")
	FormatDeGla1 = assetutil.MustLoadStringMap("assets/finc/formats/degla1.json")
	FormatDe105  = assetutil.MustLoadStringMap("assets/finc/formats/de105.json")
)
