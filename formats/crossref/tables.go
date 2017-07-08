package crossref

// FormatMap maps crossref type to format.
var FormatMap = map[string]string{
	"book":                "eBook",
	"book-chapter":        "ElectronicBookPart",
	"book-part":           "ElectronicBookPart",
	"book-section":        "ElectronicBookPart",
	"book-series":         "ElectronicSerial",
	"component":           "Unknown",
	"dataset":             "ElectronicResourceRemoteAccess",
	"dissertation":        "ElectronicThesis",
	"journal":             "ElectronicJournal",
	"journal-article":     "ElectronicArticle",
	"journal-issue":       "ElectronicJournal",
	"monograph":           "eBook",
	"proceedings":         "ElectronicProceeding",
	"proceedings-article": "ElectronicProceeding",
	"reference-book":      "eBook",
	"reference-entry":     "ElectronicResourceRemoteAccess",
	"report":              "ElectronicArticle",
	"report-series":       "eBook",
}

// GenreMap crossref types to Open URL genres.
var GenreMap = map[string]string{
	"book":                "book",
	"book-chapter":        "bookitem",
	"book-part":           "bookitem",
	"book-section":        "bookitem",
	"book-series":         "bookitem",
	"component":           "document",
	"dataset":             "document",
	"dissertation":        "book",
	"journal":             "unknown",
	"journal-article":     "article",
	"journal-issue":       "issue",
	"monograph":           "book",
	"proceedings":         "proceeding",
	"proceedings-article": "proceeding",
	"reference-book":      "book",
	"reference-entry":     "document",
	"report":              "report",
	"report-series":       "report",
}

// RefTypeMap maps crossref types to reference type.
var RefTypeMap = map[string]string{
	"book":                "EBOOK",
	"book-chapter":        "ECHAP",
	"book-part":           "ECHAP",
	"book-section":        "ECHAP",
	"book-series":         "SER",
	"component":           "GEN",
	"dataset":             "DATA",
	"dissertation":        "THES",
	"journal":             "EJOUR",
	"journal-article":     "EJOUR",
	"journal-issue":       "EJOUR",
	"monograph":           "EBOOK",
	"proceedings":         "CONF",
	"proceedings-article": "CONF",
	"reference-book":      "EBOOK",
	"reference-entry":     "GEN",
	"report":              "RPRT",
	"report-series":       "SER",
}
