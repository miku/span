package zvdd

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/miku/span/formats/finc"
)

var simpleDatePattern = regexp.MustCompile("[12][0-9][0-9][0-9]")

// Mods found embedded in Mets.
type Mods struct {
	XMLName   xml.Name `xml:"mods"`
	TitleInfo struct {
		Title    string `xml:"title"`
		SubTitle string `xml:"subTitle"`
	} `xml:"titleInfo"`
	Type  string `xml:"typeOfResource"`
	Genre struct {
		Authority string `xml:"authority,attr"`
		Value     string `xml:",chardata"`
	} `xml:"genre"`
	Origin []struct {
		Place struct {
			Term struct {
				Type  string `xml:"type,attr"`
				Value string `xml:",chardata"`
			} `xml:"placeTerm"`
		} `xml:"place"`
		DateIssued []struct {
			Date     string `xml:",chardata"`
			Encoding string `xml:"encoding,attr"`
			KeyDate  string `xml:"keyDate,attr"`
		} `xml:"dateIssued"`
		Issuance  string   `xml:"issuance"`
		Publisher []string `xml:"publisher"`
		Edition   []string `xml:"edition"`
	} `xml:"originInfo"`
	Language []struct {
		Authority string `xml:"authority,attr"`
		Type      string `xml:"type,attr"`
		Value     string `xml:",chardata"`
	} `xml:"language>languageTerm"`
	Physical struct {
		Extent string `xml:"extent"`
		Note   string `xml:"note"`
	} `xml:"physicalDescription"`
	Note []struct {
		Type  string `xml:"type,attr"`
		Value string `xml:",chardata"`
	} `xml:"note"`
	Identifier []struct {
		Type  string `xml:"type,attr"`
		Value string `xml:",chardata"`
	} `xml:"identifier"`
	Location struct {
		PhysicalLocation struct {
			Authority string `xml:"authority,attr"`
			Value     string `xml:",chardata"`
		} `xml:"physicalLocation"`
		HoldingsSimple struct {
			CopyInformation struct {
				SubLocation  string `xml:"subLocation"`
				ShelfLocator string `xml:"shelfLocator"`
			}
		} `xml:"holdingsSimple"`
		URL []string `xml:"URL"`
	} `xml:"location"`
	Extension struct {
		Value string `xml:",innerxml"`
	} `xml:"extension"`
	RecordInfo []struct {
		RecordCreationDate struct {
			Encoding string `xml:"encoding,attr"`
			Value    string `xml:",chardata"`
		} `xml:"recordCreationDate"`
		RecordIdentifier struct {
			Source string `xml:"source,attr"`
			Value  string `xml:",chardata"`
		} `xml:"recordIdentifier"`
		DescriptionStandard string `xml:"descriptionStandard"`
	}
	Classification []struct {
		Authority string `xml:"authority"`
		Value     string `xml:",chardata"`
	} `xml:"classification"`
}

// Mets fields, relevant for conversion.
type Mets struct {
	XMLName xml.Name `xml:"mets"`
	Header  struct {
		XMLName xml.Name `xml:"metsHdr"`
		Agents  []struct {
			Role      string `xml:"ROLE,attr"`
			Type      string `xml:"TYPE,attr"`
			OtherType string `xml:"OTHERTYPE,attr"`
			Name      string `xml:"name"`
		} `xml:"agent"`
	}
	DmdSection struct {
		// Wrap Wrap
		Wrap struct {
			MimeType  string `xml:"MIMETYPE,attr"`
			Type      string `xml:"MDTYPE,attr"`
			OtherType string `xml:"OTHERMDTYPE,attr"`
			Data      struct {
				Mods Mods `xml:"mods"`
				// Value string `xml:",innerxml"`
				// Mods found embedded in Mets.
			} `xml:"xmlData"`
		} `xml:"mdWrap"`
	} `xml:"dmdSec"`
	AmdSection struct {
		ID     string `xml:"ID,attr"`
		Rights struct {
			ID   string `xml:"ID,attr"`
			Wrap struct {
				MimeType  string `xml:"MIMETYPE,attr"`
				Type      string `xml:"MDTYPE,attr"`
				OtherType string `xml:"OTHERMDTYPE,attr"`
				Data      struct {
					Value string `xml:",innerxml"`
				} `xml:"xmlData"`
			} `xml:"mdWrap"`
		} `xml:"rightsMD"`
		Provenience struct {
			ID   string `xml:"ID,attr"`
			Wrap struct {
				MimeType  string `xml:"MIMETYPE,attr"`
				Type      string `xml:"MDTYPE,attr"`
				OtherType string `xml:"OTHERMDTYPE,attr"`
				Data      struct {
					Value string `xml:",innerxml"`
				} `xml:"xmlData"`
			} `xml:"mdWrap"`
		} `xml:"digiprovMD"`
	} `xml:"amdSec"`
	FileSection struct {
		FileGroup struct {
			Use   string `xml:"USE,attr"`
			Files []struct {
				MimeType     string `xml:"MIMETYPE,attr"`
				Checksum     string `xml:"CHECKSUM,attr"`
				Created      string `xml:"CREATED,attr"`
				ChecksumType string `xml:"CHECKSUMTYPE,attr"`
				Size         string `xml:"SIZE,attr"`
				ID           string `xml:"ID,attr"`
				FLocat       struct {
					Ref  string `xml:"href,attr"`
					Type string `xml:"LOCTYPE,attr"`
				} `xml:"FLocat"`
			} `xml:"file"`
		} `xml:"fileGrp"`
	} `xml:"fileSec"`
}

// MetsRecord contains a GetRequest response.
type MetsRecord struct {
	XMLName xml.Name `xml:"record"`
	Header  struct {
		Identifier string `xml:"identifier"`
		Datestamp  string `xml:"datestamp"`
	} `xml:"header"`
	Metadata struct {
		Mets Mets `xml:"mets"`
	} `xml:"metadata"`
}

// URL returns a list of URLs found in document.
func (r *MetsRecord) URL() (urls []string) {
	if urn := r.URN(); urn != "" {
		urls = append(urls, fmt.Sprintf("http://nbn-resolving.de/%s", urn))
	}
	return
}

func (r *MetsRecord) URN() string {
	mets := r.Metadata.Mets
	mods := mets.DmdSection.Wrap.Data.Mods
	for _, id := range mods.Identifier {
		if strings.ToLower(id.Type) == "urn" {
			return id.Value
		}
	}
	return ""
}

// DOI returns a doi if applicable.
func (r *MetsRecord) DOI() string {
	mets := r.Metadata.Mets
	mods := mets.DmdSection.Wrap.Data.Mods
	for _, id := range mods.Identifier {
		if strings.ToLower(id.Type) == "doi" {
			return id.Value
		}
	}
	return ""
}

// ToIntermediateSchema converts the input to intermediate schema.
func (r *MetsRecord) ToIntermediateSchema() (output *finc.IntermediateSchema, err error) {
	output = finc.NewIntermediateSchema()

	mets := r.Metadata.Mets
	mods := mets.DmdSection.Wrap.Data.Mods
	output.ArticleTitle = mods.TitleInfo.Title
	output.ArticleSubtitle = mods.TitleInfo.SubTitle
	output.DOI = r.DOI()
	output.URL = r.URL()

	// Use earliest date.
	var dates []string
	for _, origin := range mods.Origin {
		output.Publishers = origin.Publisher
		for _, di := range origin.DateIssued {
			dates = append(dates, di.Date)
		}
	}

	sort.Strings(dates)
	if len(dates) > 0 {
		output.RawDate = dates[0]
		output.Date, err = parseDate(output.RawDate)
		if err != nil {
			log.Println(err)
			return output, nil
		}
	}

	for _, lang := range mods.Language {
		output.Languages = append(output.Languages, lang.Value)
	}

	output.SourceID = "93"
	output.RecordID = fmt.Sprintf("ai-%s-%s", output.SourceID,
		base64.RawURLEncoding.EncodeToString([]byte(r.Header.Identifier)))
	output.MegaCollections = []string{"ZVDD"}
	output.Genre = "document"
	return output, nil
}

func parseDate(v string) (time.Time, error) {
	layouts := []string{
		"2006",
		"[2006]",
		"02.01.2006",
		"2006",
		"2006.",
		"um 2006",
		"2006-01-02",
		"1.2006",
		"20XX",
		"19XX",
		"184X",
	}

	for _, l := range layouts {
		t, err := time.Parse(l, v)
		if err == nil {
			return t, nil
		}
	}

	// Second pass.
	vv := simpleDatePattern.FindString(v)

	for _, l := range layouts {
		t, err := time.Parse(l, vv)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", v)
}
