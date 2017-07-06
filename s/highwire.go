package s

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/miku/span/finc"
)

// isoMap maps two letter codes to three letter codes, TODO(miku): keep these
// lists around in files.
var isoMap = map[string]string{
	"aa": "aar",
	"ab": "abk",
	"af": "afr",
	"ak": "aka",
	"am": "amh",
	"ar": "ara",
	"an": "arg",
	"as": "asm",
	"av": "ava",
	"ae": "ave",
	"ay": "aym",
	"az": "aze",
	"ba": "bak",
	"bm": "bam",
	"be": "bel",
	"bn": "ben",
	"bi": "bis",
	"bo": "bod",
	"bs": "bos",
	"br": "bre",
	"bg": "bul",
	"ca": "cat",
	"cs": "ces",
	"ch": "cha",
	"ce": "che",
	"cu": "chu",
	"cv": "chv",
	"kw": "cor",
	"co": "cos",
	"cr": "cre",
	"cy": "cym",
	"da": "dan",
	"de": "deu",
	"dv": "div",
	"dz": "dzo",
	"el": "ell",
	"en": "eng",
	"eo": "epo",
	"et": "est",
	"eu": "eus",
	"ee": "ewe",
	"fo": "fao",
	"fa": "fas",
	"fj": "fij",
	"fi": "fin",
	"fr": "fra",
	"fy": "fry",
	"ff": "ful",
	"gd": "gla",
	"ga": "gle",
	"gl": "glg",
	"gv": "glv",
	"gn": "grn",
	"gu": "guj",
	"ht": "hat",
	"ha": "hau",
	"sh": "hbs",
	"he": "heb",
	"hz": "her",
	"hi": "hin",
	"ho": "hmo",
	"hr": "hrv",
	"hu": "hun",
	"hy": "hye",
	"ig": "ibo",
	"io": "ido",
	"ii": "iii",
	"iu": "iku",
	"ie": "ile",
	"ia": "ina",
	"id": "ind",
	"ik": "ipk",
	"is": "isl",
	"it": "ita",
	"jv": "jav",
	"ja": "jpn",
	"kl": "kal",
	"kn": "kan",
	"ks": "kas",
	"ka": "kat",
	"kr": "kau",
	"kk": "kaz",
	"km": "khm",
	"ki": "kik",
	"rw": "kin",
	"ky": "kir",
	"kv": "kom",
	"kg": "kon",
	"ko": "kor",
	"kj": "kua",
	"ku": "kur",
	"lo": "lao",
	"la": "lat",
	"lv": "lav",
	"li": "lim",
	"ln": "lin",
	"lt": "lit",
	"lb": "ltz",
	"lu": "lub",
	"lg": "lug",
	"mh": "mah",
	"ml": "mal",
	"mr": "mar",
	"mk": "mkd",
	"mg": "mlg",
	"mt": "mlt",
	"mn": "mon",
	"mi": "mri",
	"ms": "msa",
	"my": "mya",
	"na": "nau",
	"nv": "nav",
	"nr": "nbl",
	"nd": "nde",
	"ng": "ndo",
	"ne": "nep",
	"nl": "nld",
	"nn": "nno",
	"nb": "nob",
	"no": "nor",
	"ny": "nya",
	"oc": "oci",
	"oj": "oji",
	"or": "ori",
	"om": "orm",
	"os": "oss",
	"pa": "pan",
	"pi": "pli",
	"pl": "pol",
	"pt": "por",
	"ps": "pus",
	"qu": "que",
	"rm": "roh",
	"ro": "ron",
	"rn": "run",
	"ru": "rus",
	"sg": "sag",
	"sa": "san",
	"si": "sin",
	"sk": "slk",
	"sl": "slv",
	"se": "sme",
	"sm": "smo",
	"sn": "sna",
	"sd": "snd",
	"so": "som",
	"st": "sot",
	"es": "spa",
	"sq": "sqi",
	"sc": "srd",
	"sr": "srp",
	"ss": "ssw",
	"su": "sun",
	"sw": "swa",
	"sv": "swe",
	"ty": "tah",
	"ta": "tam",
	"tt": "tat",
	"te": "tel",
	"tg": "tgk",
	"tl": "tgl",
	"th": "tha",
	"ti": "tir",
	"to": "ton",
	"tn": "tsn",
	"ts": "tso",
	"tk": "tuk",
	"tr": "tur",
	"tw": "twi",
	"ug": "uig",
	"uk": "ukr",
	"ur": "urd",
	"uz": "uzb",
	"ve": "ven",
	"vi": "vie",
	"vo": "vol",
	"wa": "wln",
	"wo": "wol",
	"xh": "xho",
	"yi": "yid",
	"yo": "yor",
	"za": "zha",
	"zh": "zho",
	"zu": "zul",
}

// SourceIdentifier for internal bookkeeping.
const (
	SourceIdentifier = "200"
	Format           = "ElectronicArticle"
	Genre            = "article"
	DefaultRefType   = "EJOUR"
)

// Record is a sketch for highwire XML.
type Record struct {
	XMLName xml.Name `xml:"Record"`
	Header  struct {
		Status     string   `xml:"status,attr"`
		Identifier string   `xml:"identifier"`
		Datestamp  string   `xml:"datestamp"`
		SetSpec    []string `xml:"setSpec"`
	} `xml:"header"`
	Metadata struct {
		DC struct {
			Title       []string `xml:"title"`
			Creator     []string `xml:"creator"`
			Subject     []string `xml:"subject"`
			Publisher   []string `xml:"publisher"`
			Date        []string `xml:"date"`
			Type        []string `xml:"type"`
			Format      []string `xml:"format"`
			Identifier  []string `xml:"identifier"`
			Language    []string `xml:"language"`
			Rights      []string `xml:"rights"`
			Description []string `xml:"description"`
		} `xml:"dc"`
	} `xml:"metadata"`
	About string `xml:"about"`
}

// ToIntermediateSchema sketch.
func (r Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	encodedIdentifier := base64.RawURLEncoding.EncodeToString([]byte(r.Header.Identifier))
	output.RecordID = fmt.Sprintf("ai-%s-%s", SourceIdentifier, encodedIdentifier)
	output.SourceID = SourceIdentifier
	output.Genre = Genre
	output.Format = Format
	output.RefType = DefaultRefType

	if len(r.Metadata.DC.Publisher) > 0 {
		output.MegaCollection = fmt.Sprintf("%s (HighWire)", r.Metadata.DC.Publisher[0])
	}

	if len(r.Metadata.DC.Title) > 0 {
		output.ArticleTitle = r.Metadata.DC.Title[0]
	}
	for _, v := range r.Metadata.DC.Creator {
		v = strings.TrimSpace(v)
		v = strings.TrimRight(v, ",")
		output.Authors = append(output.Authors, finc.Author{Name: v})
	}
	for _, v := range r.Metadata.DC.Identifier {
		if strings.HasPrefix(v, "http") {
			output.URL = append(output.URL, v)
		}
		if strings.HasPrefix(v, "http://dx.doi.org/") {
			output.DOI = strings.Replace(v, "http://dx.doi.org/", "", -1)
		}
	}
	output.Abstract = strings.Join(r.Metadata.DC.Description, "\n")

	for _, v := range r.Metadata.DC.Publisher {
		output.Publishers = append(output.Publishers, v)
	}
	for _, v := range r.Metadata.DC.Language {
		v = strings.TrimSpace(v)
		if len(v) == 2 {
			tlc, ok := isoMap[v]
			if !ok {
				return output, fmt.Errorf("unknown iso code: %v", v)
			}
			output.Languages = append(output.Languages, tlc)
		} else {
			return output, fmt.Errorf("not a three letter code: %v", v)
		}
	}
	for _, v := range r.Metadata.DC.Subject {
		output.Subjects = append(output.Subjects, v)
	}
	if len(r.Metadata.DC.Date) > 0 && len(r.Metadata.DC.Date[0]) >= 10 {
		// 1993-01-01 00:00:00.0
		t, err := time.Parse("2006-01-02", r.Metadata.DC.Date[0][:10])
		if err != nil {
			return output, err
		}
		output.Date = t
		output.RawDate = t.Format("2006-01-02")
	} else {
		return output, fmt.Errorf("could not parse date: %v", r.Metadata.DC.Date)
	}

	return output, nil
}
