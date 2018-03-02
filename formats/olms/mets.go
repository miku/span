package olms

import (
	"encoding/xml"
	"fmt"
	"log"
	"strings"

	"github.com/miku/span/formats/finc"
)

// MetsRecord was generated 2018-03-02 12:54:13 by tir on hayiti.
type MetsRecord struct {
	XMLName xml.Name `xml:"record"`
	Text    string   `xml:",chardata"`
	Header  struct {
		Text       string `xml:",chardata"`
		Identifier struct {
			Text string `xml:",chardata"` // www.olmsonline.de:PPN5212...
		} `xml:"identifier"`
		Datestamp struct {
			Text string `xml:",chardata"` // 2012-02-01T11:12:51Z, 201...
		} `xml:"datestamp"`
		SetSpec struct {
			Text string `xml:",chardata"` // philosophie/neuzeit_(bis_...
		} `xml:"setSpec"`
	} `xml:"header"`
	Metadata struct {
		Text string `xml:",chardata"`
		Mets struct {
			Text           string `xml:",chardata"`
			Mets           string `xml:"mets,attr"`
			Mods           string `xml:"mods,attr"`
			Dv             string `xml:"dv,attr"`
			Xlink          string `xml:"xlink,attr"`
			Xsi            string `xml:"xsi,attr"`
			SchemaLocation string `xml:"schemaLocation,attr"`
			DmdSec         []struct {
				Text   string `xml:",chardata"`
				ID     string `xml:"ID,attr"`
				MdWrap struct {
					Text    string `xml:",chardata"`
					MDTYPE  string `xml:"MDTYPE,attr"`
					XmlData struct {
						Text string `xml:",chardata"`
						Mods struct {
							Text       string `xml:",chardata"`
							Mods       string `xml:"mods,attr"`
							RecordInfo struct {
								Text             string `xml:",chardata"`
								RecordIdentifier struct {
									Text   string `xml:",chardata"` // PPN521206804, PPN52126692...
									Source string `xml:"source,attr"`
								} `xml:"recordIdentifier"`
							} `xml:"recordInfo"`
							Identifier []struct {
								Text string `xml:",chardata"` // http://www.olmsonline.de/...
								Type string `xml:"type,attr"`
							} `xml:"identifier"`
							Location struct {
								Text string `xml:",chardata"`
								URL  struct {
									Text string `xml:",chardata"` // http://www.olmsonline.de/...
								} `xml:"url"`
								PhysicalLocation struct {
									Text string `xml:",chardata"` // <SUB Göttingen>76 A 1403...
									Type string `xml:"type,attr"`
								} `xml:"physicalLocation"`
							} `xml:"location"`
							TitleInfo struct {
								Text  string `xml:",chardata"`
								Title struct {
									Text string `xml:",chardata"` // Ausgewählte Werke, Ausge...
								} `xml:"title"`
							} `xml:"titleInfo"`
							Language struct {
								Text         string `xml:",chardata"`
								LanguageTerm struct {
									Text      string `xml:",chardata"` // lat, ger, ger, ger, ger, ...
									Type      string `xml:"type,attr"`
									Authority string `xml:"authority,attr"`
								} `xml:"languageTerm"`
							} `xml:"language"`
							OriginInfo []struct {
								Text       string `xml:",chardata"`
								DateIssued struct {
									Text     string `xml:",chardata"` // 1993, 1994, 2006, 1978, 1...
									Keydate  string `xml:"keydate,attr"`
									Encoding string `xml:"encoding,attr"`
									KeyDate  string `xml:"keyDate,attr"`
								} `xml:"dateIssued"`
								Place struct {
									Text      string `xml:",chardata"`
									PlaceTerm struct {
										Text string `xml:",chardata"` // Hildesheim [u.a.], Götti...
										Type string `xml:"type,attr"`
									} `xml:"placeTerm"`
								} `xml:"place"`
								Publisher struct {
									Text string `xml:",chardata"` // Olms, Georg Olms Verlag A...
								} `xml:"publisher"`
								Edition struct {
									Text string `xml:",chardata"` // [Electronic ed.], [Electr...
								} `xml:"edition"`
								DateCaptured struct {
									Text     string `xml:",chardata"` // 2012-02-01, 2007-06-14, 2...
									Encoding string `xml:"encoding,attr"`
								} `xml:"dateCaptured"`
							} `xml:"originInfo"`
							Subject struct {
								Text      string `xml:",chardata"`
								Authority string `xml:"authority,attr"`
								Topic     struct {
									Text string `xml:",chardata"` // thomausg, fouqausg, mess,...
								} `xml:"topic"`
							} `xml:"subject"`
							Classification []struct {
								Text      string `xml:",chardata"` // Philosophie/Neuzeit (bis ...
								Authority string `xml:"authority,attr"`
							} `xml:"classification"`
							PhysicalDescription struct {
								Text          string `xml:",chardata"`
								DigitalOrigin struct {
									Text string `xml:",chardata"` // reformatted digital, refo...
								} `xml:"digitalOrigin"`
								Extent struct {
									Text string `xml:",chardata"` // V69 pages, 1 page, 2 page...
								} `xml:"extent"`
							} `xml:"physicalDescription"`
							Name []struct {
								Text string `xml:",chardata"`
								Type string `xml:"type,attr"`
								Role struct {
									Text     string `xml:",chardata"`
									RoleTerm struct {
										Text      string `xml:",chardata"` // aut, aut, aut, aut, aut, ...
										Type      string `xml:"type,attr"`
										Authority string `xml:"authority,attr"`
									} `xml:"roleTerm"`
								} `xml:"role"`
								NamePart []struct {
									Text string `xml:",chardata"` // Thomasius, Christian, Fou...
									Type string `xml:"type,attr"`
								} `xml:"namePart"`
							} `xml:"name"`
							AccessCondition []struct {
								Text      string `xml:",chardata"` // reprint, reprint, reprint...
								Authority string `xml:"authority,attr"`
							} `xml:"accessCondition"`
							RelatedItem struct {
								Text       string `xml:",chardata"`
								Type       string `xml:"type,attr"`
								RecordInfo struct {
									Text             string `xml:",chardata"`
									RecordIdentifier struct {
										Text   string `xml:",chardata"` // PPN521206804, PPN51995679...
										Source string `xml:"source,attr"`
									} `xml:"recordIdentifier"`
								} `xml:"recordInfo"`
							} `xml:"relatedItem"`
							Part struct {
								Text   string `xml:",chardata"`
								Type   string `xml:"type,attr"`
								Order  string `xml:"order,attr"`
								Detail struct {
									Text   string `xml:",chardata"`
									Number struct {
										Text string `xml:",chardata"` // Bd. 19, Abt. I, Bd. 1, Ab...
									} `xml:"number"`
								} `xml:"detail"`
							} `xml:"part"`
						} `xml:"mods"`
					} `xml:"xmlData"`
				} `xml:"mdWrap"`
			} `xml:"dmdSec"`
			AmdSec struct {
				Text     string `xml:",chardata"`
				ID       string `xml:"ID,attr"`
				RightsMD struct {
					Text   string `xml:",chardata"`
					ID     string `xml:"ID,attr"`
					MdWrap struct {
						Text        string `xml:",chardata"`
						MIMETYPE    string `xml:"MIMETYPE,attr"`
						MDTYPE      string `xml:"MDTYPE,attr"`
						OTHERMDTYPE string `xml:"OTHERMDTYPE,attr"`
						XmlData     struct {
							Text   string `xml:",chardata"`
							Rights struct {
								Text  string `xml:",chardata"`
								Dv    string `xml:"dv,attr"`
								Owner struct {
									Text string `xml:",chardata"` // Georg Olms Verlag AG, Geo...
								} `xml:"owner"`
								OwnerLogo struct {
									Text string `xml:",chardata"` // http://www.olmsonline.de/...
								} `xml:"ownerLogo"`
								OwnerSiteURL struct {
									Text string `xml:",chardata"` // http://www.olmsonline.de/...
								} `xml:"ownerSiteURL"`
								OwnerContact struct {
									Text string `xml:",chardata"`
								} `xml:"ownerContact"`
							} `xml:"rights"`
						} `xml:"xmlData"`
					} `xml:"mdWrap"`
				} `xml:"rightsMD"`
				DigiprovMD struct {
					Text   string `xml:",chardata"`
					ID     string `xml:"ID,attr"`
					MdWrap struct {
						Text        string `xml:",chardata"`
						MIMETYPE    string `xml:"MIMETYPE,attr"`
						MDTYPE      string `xml:"MDTYPE,attr"`
						OTHERMDTYPE string `xml:"OTHERMDTYPE,attr"`
						XmlData     struct {
							Text  string `xml:",chardata"`
							Links struct {
								Text      string `xml:",chardata"`
								Dv        string `xml:"dv,attr"`
								Reference struct {
									Text string `xml:",chardata"` // http://gso.gbv.de/xslt/DB...
								} `xml:"reference"`
								Presentation struct {
									Text string `xml:",chardata"` // http://www.olmsonline.de/...
								} `xml:"presentation"`
							} `xml:"links"`
						} `xml:"xmlData"`
					} `xml:"mdWrap"`
				} `xml:"digiprovMD"`
			} `xml:"amdSec"`
			// StructMap
			FileSec struct {
				Text    string `xml:",chardata"`
				FileGrp []struct {
					Text string `xml:",chardata"`
					USE  string `xml:"USE,attr"`
					File []struct {
						Text     string `xml:",chardata"`
						ID       string `xml:"ID,attr"`
						MIMETYPE string `xml:"MIMETYPE,attr"`
						FLocat   struct {
							Text    string `xml:",chardata"`
							LOCTYPE string `xml:"LOCTYPE,attr"`
							Href    string `xml:"href,attr"`
							Xlink   string `xml:"xlink,attr"`
						} `xml:"FLocat"`
					} `xml:"file"`
				} `xml:"fileGrp"`
			} `xml:"fileSec"`
			StructLink struct {
				Text   string `xml:",chardata"`
				SmLink []struct {
					Text  string `xml:",chardata"`
					From  string `xml:"from,attr"`
					To    string `xml:"to,attr"`
					Xlink string `xml:"xlink,attr"`
				} `xml:"smLink"`
			} `xml:"structLink"`
		} `xml:"mets"`
	} `xml:"metadata"`
}

// ToIntermediateSchema converts a single record. XXX: WIP.
func (record MetsRecord) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()

	output.SourceID = "12502"
	parts := strings.Split(record.Header.Identifier.Text, ":")
	if len(parts) != 2 {
		return output, fmt.Errorf("cannot find identifier: %s", record.Header.Identifier.Text)
	}
	output.RecordID = record.Header.Identifier.Text
	output.ID = fmt.Sprintf("ai-%s-%s", output.SourceID, parts[1])
	output.MegaCollections = append(output.MegaCollections, "Olms")
	output.Genre = "article"
	output.RefType = "EJOUR"

	for _, sec := range record.Metadata.Mets.DmdSec {
		log.Println(sec.MdWrap.XmlData.Mods.TitleInfo.Title.Text)
	}
	return output, nil
}
