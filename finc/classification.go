package finc

import "regexp"

const NOT_ASSIGNED = "not assigned"

var LCCFincMap = map[string]string{
	"^[AZ]+.*":           "Allgemeines",
	"^B[0-9CDHJ]+.*":     "Philosophie",
	"^BF+.*":             "Psychologie",
	"^B[LMPQRSTVX]+.*":   "Theologie und Religionswissenschaften",
	"^C[0-9BDEJNRST]+.*": "Geschichte",
	"^CC+.*":             "Klassische Archäologie",
	"^[DEF]+.*":          "Geschichte",
	"^G[0-9ABCEF]+.*":    "Geographie",
	"^G[NRT]+.*":         "Ethnologie (Volks- und Völkerkunde)",
	"^GV+.*":             "Sport",
	"^H[0-9AMNQSTVX]+.*": "Soziologie",
	"^H[BCDEFGJ]+.*":     "Wirtschaftswissenschaften",
	"^J+.*":              "Politologie",
	"^K+.*":              "Rechtswissenschaft",
	"^L+.*":              "Pädagogik",
	"^M+.*":              "Musikwissenschaft",
	"^N+.*":              "Kunst und Kunstgeschichte",
	"^P[0-9BHJKLMNZ]+.*": "Allgemeine und vergleichende Sprach- und Literaturwissenschaft, Indogermanistik, Außereuropäische Sprachen und Literaturen",
	"^PA+.*":             "Klassische Philologie",
	"^P[CQ]+.*":          "Romanistik",
	"^P[DFT]+.*":         "Germanistik, Niederlandistik, Skandinavistik",
	"^P[ERS]+.*":         "Anglistik, Amerikanistik",
	"^PG+.*":             "Slawistik",
	"^Q[0-9]+.*":         "Allgemeine Naturwissenschaft",
	"^QA+.*":             "Mathematik",
	"^Q[BC]+.*":          "Physik",
	"^QD+.*":             "Chemie und Pharmazie",
	"^QE+.*":             "Geologie und Paläontologie",
	"^Q[HKLMPR]+.*":      "Biologie",
	"^R+.*":              "Medizin",
	"^S+.*":              "Land- und Forstwirtschaft, Gartenbau, Fischereiwirtschaft, Hauswirtschaft",
	"^T+.*":              "Technik",
	"^[UV]+.*":           "Militärwissenschaft",
}

var DDCFincMap = map[string]string{
	"^0[1-35-9][0-9].*":               "Allgemeines",
	"((^2[0-9]{2})|(^1[37][0-9])).*":  "Theologie und Religionswissenschaft",
	"^((1[0-46-9])|21)[0-9].*":        "Philosophie",
	"^1[35][0-9].*":                   "Psychologie",
	"^37[0-9].*":                      "Pädagogik",
	"^(4[019][0-9])|(8[09][0-9]).*":   "Allgemeine und vergleichende Sprach- und Literaturwissenschaft, Indogermanistik, Außereuropäische Sprachen und Literaturen",
	"^(43[0-9])|(83[0-9]).*":          "Germanistik, Niederlandistik, Skandinavistik",
	"^(42[0-9])|(8[12][0-9]).*":       "Anglistik, Amerikanistik",
	"^(4[4-6][0-9])|(8[4-6][0-9]).*":  "Romanistik",
	"^39[0-9].*":                      "Ethnologie (Volks- und Völkerkunde)",
	"^93[0-9].*":                      "Klassische Archäologie",
	"^7[0234-7][0-9].*":               "Kunst und Kunstgeschichte",
	"^78[0-9].*":                      "Musikwissenschaft",
	"^32[0-9].*":                      "Politologie",
	"^3[0-367][0-9].*":                "Soziologie",
	"^35[0-9].*":                      "Militärwissenschaft",
	"^((9[012-9])|(1[89])|27)[0-9].*": "Geschichte",
	"^3[45][0-9].*":                   "Rechtswissenschaft",
	"^(3[38][0-9])|(65[0-9]).*":       "Wirtschaftswissenschaften",
	"^(91|55)[0-9].*":                 "Geographie",
	"^(51|16|31)[0-9].*":              "Mathematik",
	"^(00|77)[0-9].*":                 "Informatik",
	"^50[0-9].*":                      "Allgemeine Naturwissenschaft",
	"^5[56][0-9].*":                   "Geologie und Paläontologie",
	"^5[23][0-9].*":                   "Physik",
	"^(54[0-9])|(66[0-9]).*":          "Chemie und Pharmazie",
	"^5[7-9][0-9].*":                  "Biologie",
	"^(61|57)[0-9].*":                 "Medizin",
	"^(6[34][0-9])|(7[134][0-9]).*":   "Land- und Forstwirtschaft, Gartenbau, Fischereiwirtschaft, Hauswirtschaft",
	"^((6[0-789])|72)[0-9].*":         "Technik",
	"^79[0-9].*":                      "Sport",
	"^((4[78])|8[78])[0-9].*":         "Klassische Philologie",
	"^[48]8[0-9].*":                   "Neugriechische Philologie",
	"^[48]7[0-9].*":                   "Neulateinische Philologie",
	"^04[0-9].*":                      NOT_ASSIGNED,
}

// RegexpResolve takes a string and a map with regular expressions
// and returns the value of the map, where the key regexp matches.
func RegexpResolve(s string, m map[string]string) string {
	for p, c := range m {
		rx := regexp.MustCompile(p)
		if rx.MatchString(s) {
			return c
		}
	}
	return NOT_ASSIGNED
}
