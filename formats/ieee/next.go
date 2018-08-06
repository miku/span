package ieee

import "encoding/xml"

// XPublication was generated 2018-07-27 13:48:36 by tir on hayiti. XXX: Transistion to this schema.
type XPublication struct {
	XMLName xml.Name `xml:"publication"`
	Text    string   `xml:",chardata"`
	Title   struct {
		Text string `xml:",chardata"` // IET Generation, Transmiss...
	} `xml:"title"`
	Normtitle struct {
		Text string `xml:",chardata"` // Generation, Transmission ...
	} `xml:"normtitle"`
	Publicationinfo struct {
		Text    string `xml:",chardata"`
		Idamsid struct {
			Text string `xml:",chardata"` // 0b000064806f4938, 0b00006...
		} `xml:"idamsid"`
		Publicationtype struct {
			Text string `xml:",chardata"` // Periodical, Periodical, P...
		} `xml:"publicationtype"`
		Publicationsubtype struct {
			Text string `xml:",chardata"` // IEE Periodical, IEE Perio...
		} `xml:"publicationsubtype"`
		Pubstatus struct {
			Text string `xml:",chardata"` // Active, Active, Active, A...
		} `xml:"pubstatus"`
		Publicationopenaccess struct {
			Text string `xml:",chardata"` // F, F, F, F, F, F, F, F, F...
		} `xml:"publicationopenaccess"`
		StandardID struct {
			Text string `xml:",chardata"` // 0, 0, 0, 0, 0, 0, 0, 0, 0...
		} `xml:"standard_id"`
		ISSN []struct {
			Text      string `xml:",chardata"` // 1751-8687, 1751-8695, 175...
			Mediatype string `xml:"mediatype,attr"`
		} `xml:"issn"`
		Keywordset struct {
			Text        string `xml:",chardata"`
			Keywordtype string `xml:"keywordtype,attr"`
			Keyword     []struct {
				Text        string `xml:",chardata"`
				Keywordterm struct {
					Text string `xml:",chardata"` // Changed "Copyright" from ...
				} `xml:"keywordterm"`
			} `xml:"keyword"`
		} `xml:"keywordset"`
		Pubtopicalbrowseset struct {
			Text             string `xml:",chardata"`
			Pubtopicalbrowse []struct {
				Text string `xml:",chardata"` // Power, Energy and Industr...
			} `xml:"pubtopicalbrowse"`
		} `xml:"pubtopicalbrowseset"`
		Copyrightgroup struct {
			Text      string `xml:",chardata"`
			Copyright struct {
				Text string `xml:",chardata"`
				Year struct {
					Text string `xml:",chardata"` // 2007-2012, 2007-2012, 200...
				} `xml:"year"`
				Holder struct {
					Text string `xml:",chardata"` // IET, IET, IET, IET, IET, ...
				} `xml:"holder"`
			} `xml:"copyright"`
		} `xml:"copyrightgroup"`
		Publisher struct {
			Text          string `xml:",chardata"`
			Publishername struct {
				Text string `xml:",chardata"` // IET, IET, IET, IET, IET, ...
			} `xml:"publishername"`
			Address struct {
				Text    string `xml:",chardata"`
				Country struct {
					Text string `xml:",chardata"` // UK, UK, UK, UK, UK, UK, U...
				} `xml:"country"`
				Street struct {
					Text string `xml:",chardata"` // Michael Faraday House, Mi...
				} `xml:"street"`
				Otheraddr struct {
					Text string `xml:",chardata"` // Six Hills Way, Six Hills ...
				} `xml:"otheraddr"`
				City struct {
					Text string `xml:",chardata"` // Stevenage, Stevenage, Ste...
				} `xml:"city"`
				State struct {
					Text string `xml:",chardata"` // Hertfordshire, Hertfordsh...
				} `xml:"state"`
				Postcode struct {
					Text string `xml:",chardata"` // SG1 2AY, SG1 2AY, SG1 2AY...
				} `xml:"postcode"`
			} `xml:"address"`
			Publisherloc struct {
				Text string `xml:",chardata"` // Michael Faraday House, Mi...
			} `xml:"publisherloc"`
		} `xml:"publisher"`
		Holdstatus struct {
			Text string `xml:",chardata"` // Publish, Publish, Publish...
		} `xml:"holdstatus"`
		Confgroup struct {
			Text          string `xml:",chardata"`
			DoiPermission struct {
				Text string `xml:",chardata"` // F, F, F, F, F, F, F, F, F...
			} `xml:"doi_permission"`
			Confdate []struct {
				Text         string `xml:",chardata"`
				Confdatetype string `xml:"confdatetype,attr"`
				Year         struct {
					Text string `xml:",chardata"` // 1994, 1994, 1994, 1994, 1...
				} `xml:"year"`
				Month struct {
					Text string `xml:",chardata"` // Aug., Aug., Aug., Aug., A...
				} `xml:"month"`
				Day struct {
					Text string `xml:",chardata"` // 24, 26, 24, 26, 24, 26, 2...
				} `xml:"day"`
			} `xml:"confdate"`
			Confcountry struct {
				Text string `xml:",chardata"` // UK, UK, UK, UK, UK, UK, U...
			} `xml:"confcountry"`
			Conftitle struct {
				Text string `xml:",chardata"` // 13th International Confer...
			} `xml:"conftitle"`
			Conflocation struct {
				Text string `xml:",chardata"` // Edinburgh, Edinburgh, Edi...
			} `xml:"conflocation"`
			ConferenceType struct {
				Text string `xml:",chardata"` // E, E, E, E, E, E, E, E, E...
			} `xml:"conference_type"`
			ConferenceStage struct {
				Text string `xml:",chardata"`
			} `xml:"conference_stage"`
		} `xml:"confgroup"`
		Amsid struct {
			Text string `xml:",chardata"` // 4082359, 4082359, 4082359...
		} `xml:"amsid"`
		Coden struct {
			Text string `xml:",chardata"` // IGTDAW, IGTDAW, IGTDAW, I...
		} `xml:"coden"`
		Packagememberset struct {
			Text          string `xml:",chardata"`
			Packagemember []struct {
				Text string `xml:",chardata"` // None, None, None, None, N...
			} `xml:"packagemember"`
		} `xml:"packagememberset"`
		Ieeeabbrev struct {
			Text string `xml:",chardata"` // IEE, IEE, IEE, IEE, IEE, ...
		} `xml:"ieeeabbrev"`
		Acronym []struct {
			Text        string `xml:",chardata"` // IET, IET, IET, IET, IET, ...
			Acronymtype string `xml:"acronymtype,attr"`
		} `xml:"acronym"`
		Pubsponsor []struct {
			Text    string `xml:",chardata"`
			Society struct {
				Text string `xml:",chardata"` // Univ. Maryland Inst. Adv....
			} `xml:"society"`
		} `xml:"pubsponsor"`
		ISBN []struct {
			Text      string `xml:",chardata"` // 978-1-78561-139-1, 978-1-...
			Isbntype  string `xml:"isbntype,attr"`
			Mediatype string `xml:"mediatype,attr"`
		} `xml:"isbn"`
		Invpartnumber struct {
			Text string `xml:",chardata"` // PEP0554Z, PES07947, PEZ08...
		} `xml:"invpartnumber"`
		Productnumber []struct {
			Text    string `xml:",chardata"` // PEP0554Z, PES07947, PEZ08...
			Pubtype string `xml:"pubtype,attr"`
		} `xml:"productnumber"`
		LCCN struct {
			Text string `xml:",chardata"` // 2006468679, 2006468679, 2...
		} `xml:"lccn"`
		BmsProductNumber []struct {
			Text      string `xml:",chardata"` // CFP17O22-ART, CFP17O22-AR...
			Mediatype string `xml:"mediatype,attr"`
		} `xml:"bms_product_number"`
		Tcn struct {
			Text string `xml:",chardata"` // 43375, 43375, 43375, 4337...
		} `xml:"tcn"`
		PublicationIeeeSponsored struct {
			Text string `xml:",chardata"` // No, No, No, No, No, No, N...
		} `xml:"publication_ieee_sponsored"`
		T13comments struct {
			Text string `xml:",chardata"` // Quality Review 7.13.2017,...
		} `xml:"t13comments"`
		Publicationdoi struct {
			Text string `xml:",chardata"` // 10.1109/INTMAG.5091, 10.1...
		} `xml:"publicationdoi"`
		ConfURL struct {
			Text string `xml:",chardata"` // wirelesshealth.org/2016, ...
		} `xml:"conf_url"`
		PubApprovalDate struct {
			Text string `xml:",chardata"` // 11, 11, 11, 11, 11, 11, 1...
		} `xml:"PubApprovalDate"`
		Stdnumber struct {
			Text string `xml:",chardata"` // 8802-1Q:2016/Cor 1-2017, ...
		} `xml:"stdnumber"`
		StandardSubtype struct {
			Text string `xml:",chardata"` // IEEE Standard, IEEE Stand...
		} `xml:"standard_subtype"`
		StandardStatus struct {
			Text string `xml:",chardata"` // Active, Active, Active, A...
		} `xml:"standard_status"`
		Standardmodifierset struct {
			Text             string `xml:",chardata"`
			StandardModifier struct {
				Text string `xml:",chardata"` // Approved, Approved, Appro...
			} `xml:"standard_modifier"`
		} `xml:"standardmodifierset"`
		StandardFamily struct {
			Text string `xml:",chardata"` // 802.1Q-2014/Cor 1, 802.3b...
		} `xml:"standard_family"`
		Standardpackageset struct {
			Text            string `xml:",chardata"`
			StandardPackage []struct {
				Text string `xml:",chardata"` // Local and Metropolitan Ar...
			} `xml:"standard_package"`
		} `xml:"standardpackageset"`
		Icscodes struct {
			Text     string `xml:",chardata"`
			CodeTerm []struct {
				Text    string `xml:",chardata"` // Networking, Networking, P...
				Codenum string `xml:"codenum,attr"`
			} `xml:"code_term"`
		} `xml:"icscodes"`
		Pubsponsoringcommitteeset struct {
			Text                   string `xml:",chardata"`
			Pubsponsoringcommittee []struct {
				Text string `xml:",chardata"` // LAN/MAN Standards Committ...
			} `xml:"pubsponsoringcommittee"`
		} `xml:"pubsponsoringcommitteeset"`
		StandardRelationship []struct {
			Text             string `xml:",chardata"` // 7394900, 8082796, 6086, 6...
			Prodnum          string `xml:"prodnum,attr"`
			RelationshipDate string `xml:"relationship_date,attr"`
			Type             string `xml:"type,attr"`
		} `xml:"standard_relationship"`
		AssociatedPunumber struct {
			Text string `xml:",chardata"` // 7445795, 7425111, 7995159...
		} `xml:"associated_punumber"`
		StandardBundle struct {
			Text       string `xml:",chardata"`
			BundleName struct {
				Text string `xml:",chardata"` // IEEE Standard for Design ...
			} `xml:"bundle_name"`
			BundleType struct {
				Text string `xml:",chardata"` // Mandatory, Mandatory, Man...
			} `xml:"bundle_type"`
			BundleProductNumber struct {
				Text string `xml:",chardata"` // STDRL20513, STDRL20779, S...
			} `xml:"bundle_product_number"`
			BaseStandardProductNumber struct {
				Text string `xml:",chardata"` // STD20779, STD22475, STD22...
			} `xml:"base_standard_product_number"`
		} `xml:"standard_bundle"`
	} `xml:"publicationinfo"`
	Volume struct {
		Text       string `xml:",chardata"`
		Volumeinfo struct {
			Text string `xml:",chardata"`
			Year struct {
				Text string `xml:",chardata"` // 2017, 2017, 2017, 2017, 2...
			} `xml:"year"`
			Volumenum struct {
				Text string `xml:",chardata"` // 11, 11, 11, 11, 11, 11, 1...
			} `xml:"volumenum"`
			Idamsid struct {
				Text string `xml:",chardata"` // 0b00006485b9045f, 0b00006...
			} `xml:"idamsid"`
			Issue struct {
				Text     string `xml:",chardata"`
				Issuenum struct {
					Text string `xml:",chardata"` // 2, 2, 2, 2, 2, 2, 2, 2, 2...
				} `xml:"issuenum"`
				Amsid struct {
					Text string `xml:",chardata"` // 7834146, 7834146, 7834146...
				} `xml:"amsid"`
				Issuestatus struct {
					Text string `xml:",chardata"` // Complete, Complete, Compl...
				} `xml:"issuestatus"`
				Issuepart struct {
					Text string `xml:",chardata"` // Regular Papers, Regular P...
				} `xml:"issuepart"`
				Amscreatedate struct {
					Text string `xml:",chardata"` // 1/19/2017 12:00:00 AM, 1/...
				} `xml:"amscreatedate"`
				IssueCompleteDate struct {
					Text string `xml:",chardata"`
					Year struct {
						Text string `xml:",chardata"` // 2017, 2016, 2017, 2016, 2...
					} `xml:"year"`
					Month struct {
						Text string `xml:",chardata"` // 1, 12, 1, 12, 12, 12, 12,...
					} `xml:"month"`
					Day struct {
						Text string `xml:",chardata"` // 9, 1, 12, 20, 6, 5, 5, 5,...
					} `xml:"day"`
				} `xml:"issue_complete_date"`
				Filename struct {
					Text         string `xml:",chardata"` // 8272040.pdf, 8272040.pdf,...
					Docpartition string `xml:"docpartition,attr"`
					Filetype     string `xml:"filetype,attr"`
				} `xml:"filename"`
			} `xml:"issue"`
			Notegroup struct {
				Text string `xml:",chardata"`
				Note struct {
					Text string `xml:",chardata"` // Completed and re-exported...
				} `xml:"note"`
			} `xml:"notegroup"`
		} `xml:"volumeinfo"`
		Article struct {
			Text  string `xml:",chardata"`
			Title struct {
				Text string `xml:",chardata"` // Power-dependent droop-bas...
			} `xml:"title"`
			Articleinfo struct {
				Text          string `xml:",chardata"`
				Articleseqnum struct {
					Text string `xml:",chardata"` // 3831, 3391, 3301, 5401, 4...
				} `xml:"articleseqnum"`
				Csarticlesortorder struct {
					Text string `xml:",chardata"` // 0, 0, 0, 0, 0, 0, 0, 0, 0...
				} `xml:"csarticlesortorder"`
				Articledoi struct {
					Text string `xml:",chardata"` // 10.1049/iet-gtd.2016.0764...
				} `xml:"articledoi"`
				Idamsid struct {
					Text string `xml:",chardata"` // 0b00006485b94e5b, 0b00006...
				} `xml:"idamsid"`
				Articlestatus struct {
					Text string `xml:",chardata"` // Active, Active, Active, A...
				} `xml:"articlestatus"`
				Contenttype struct {
					Text string `xml:",chardata"` // orig-research, orig-resea...
				} `xml:"contenttype"`
				Othercontenttype struct {
					Text string `xml:",chardata"` // research-article, researc...
				} `xml:"othercontenttype"`
				Articleopenaccess struct {
					Text string `xml:",chardata"` // F, F, F, F, F, F, F, F, F...
				} `xml:"articleopenaccess"`
				Articleshowflag struct {
					Text string `xml:",chardata"` // T, T, T, T, T, T, T, T, T...
				} `xml:"articleshowflag"`
				Articleplagiarizedflag struct {
					Text string `xml:",chardata"` // F, F, F, F, F, F, F, F, F...
				} `xml:"articleplagiarizedflag"`
				Articlenodoiflag struct {
					Text string `xml:",chardata"` // F, F, F, F, F, F, F, F, F...
				} `xml:"articlenodoiflag"`
				Articlecoverimageflag struct {
					Text string `xml:",chardata"` // F, F, F, F, F, F, F, F, F...
				} `xml:"articlecoverimageflag"`
				Csarticlehtmlflag struct {
					Text string `xml:",chardata"` // F, F, F, F, F, F, F, F, F...
				} `xml:"csarticlehtmlflag"`
				Articlereferenceflag struct {
					Text string `xml:",chardata"` // F, F, F, F, F, F, F, F, F...
				} `xml:"articlereferenceflag"`
				Articlepeerreviewflag struct {
					Text string `xml:",chardata"` // F, F, F, F, F, F, F, F, F...
				} `xml:"articlepeerreviewflag"`
				Holdstatus struct {
					Text string `xml:",chardata"` // Publish, Publish, Publish...
				} `xml:"holdstatus"`
				Issuenum struct {
					Text string `xml:",chardata"` // 2, 2, 2, 2, 2, 2, 2, 2, 2...
				} `xml:"issuenum"`
				Articlecopyright struct {
					Text         string `xml:",chardata"` // © The Institution of Eng...
					Holderisieee string `xml:"holderisieee,attr"`
					Year         string `xml:"year,attr"`
				} `xml:"articlecopyright"`
				Abstract []struct {
					Text         string `xml:",chardata"` // The concept of voltage so...
					Abstracttype string `xml:"abstracttype,attr"`
				} `xml:"abstract"`
				Authorgroup struct {
					Text   string `xml:",chardata"`
					Author []struct {
						Text        string `xml:",chardata"`
						Role        string `xml:"role,attr"`
						Authororder struct {
							Text string `xml:",chardata"` // 0, 0, 0, 0, 0, 0, 0, 0, 0...
						} `xml:"authororder"`
						Normname struct {
							Text string `xml:",chardata"` // Stamatiou, G., Bongiorno,...
						} `xml:"normname"`
						Nonnormname struct {
							Text string `xml:",chardata"` // Georgios Stamatiou, Massi...
						} `xml:"nonnormname"`
						Authorrefid struct {
							Text string `xml:",chardata"` // A16585073.1, A16585073.1,...
						} `xml:"authorrefid"`
						Firstname struct {
							Text string `xml:",chardata"` // Georgios, Massimo, Mikel,...
						} `xml:"firstname"`
						Surname struct {
							Text string `xml:",chardata"` // Stamatiou, Bongiorno, Arm...
						} `xml:"surname"`
						Affiliation struct {
							Text string `xml:",chardata"` // Dept. of Energy &amp; Env...
						} `xml:"affiliation"`
						Email struct {
							Text string `xml:",chardata"` // georgios.stamatiou@chalme...
						} `xml:"email"`
						Unicodefirstname struct {
							Text string `xml:",chardata"` // Abdulatif, Heshan, Ibrahi...
						} `xml:"unicodefirstname"`
						Unicodesurname struct {
							Text string `xml:",chardata"` // Alabdulatif, Kumarage, Kh...
						} `xml:"unicodesurname"`
						Authortype struct {
							Text string `xml:",chardata"` // Author, Author, Author, A...
						} `xml:"authortype"`
						Lineage struct {
							Text string `xml:",chardata"` // Jr, Jr., Junior, Junior, ...
						} `xml:"lineage"`
						Orcid struct {
							Text string `xml:",chardata"` // 0000-0002-2201-7327, 0000...
						} `xml:"orcid"`
						Unicodemiddlename struct {
							Text string `xml:",chardata"` // Alay&#xF3;n
						} `xml:"unicodemiddlename"`
						Authorbio struct {
							Text string `xml:",chardata"` // Dong Xu received the PhD ...
						} `xml:"authorbio"`
						Nickname struct {
							Text string `xml:",chardata"` // Gary, Gary, Ted
						} `xml:"nickname"`
					} `xml:"author"`
				} `xml:"authorgroup"`
				Date []struct {
					Text     string `xml:",chardata"`
					Datetype string `xml:"datetype,attr"`
					Year     struct {
						Text string `xml:",chardata"` // 2017, 2017, 2017, 2017, 2...
					} `xml:"year"`
					Month struct {
						Text string `xml:",chardata"` // 1, 1, 1, 1, 1, 1, 1, 1, 1...
					} `xml:"month"`
					Day struct {
						Text string `xml:",chardata"` // 26, 26, 27, 26, 26, 27, 2...
					} `xml:"day"`
				} `xml:"date"`
				Numpages struct {
					Text string `xml:",chardata"` // 0, 0, 0, 0, 0, 0, 0, 0, 0...
				} `xml:"numpages"`
				Size struct {
					Text string `xml:",chardata"` // 872075, 716350, 362541, 1...
				} `xml:"size"`
				Filename []struct {
					Text         string `xml:",chardata"` // 07835087.pdf, 07834938.pd...
					Docpartition string `xml:"docpartition,attr"`
					Filetype     string `xml:"filetype,attr"`
				} `xml:"filename"`
				Artpagenums struct {
					Text      string `xml:",chardata"` // 13-13, 30-31, 26-29, 4-4,...
					Endpage   string `xml:"endpage,attr"`
					Startpage string `xml:"startpage,attr"`
				} `xml:"artpagenums"`
				Pubsnumber []struct {
					Text      string `xml:",chardata"` // 16585073, 16585068, 16585...
					Pubidtype string `xml:"pubidtype,attr"`
				} `xml:"pubsnumber"`
				Numreferences struct {
					Text string `xml:",chardata"` // 24, 33, 34, 22, 25, 30, 2...
				} `xml:"numreferences"`
				Amsid struct {
					Text string `xml:",chardata"` // 7835087, 7834938, 7834955...
				} `xml:"amsid"`
				Keywordset []struct {
					Text        string `xml:",chardata"`
					Keywordtype string `xml:"keywordtype,attr"`
					Keyword     []struct {
						Text        string `xml:",chardata"`
						Keywordterm struct {
							Text string `xml:",chardata"` // voltage control, HVDC pow...
						} `xml:"keywordterm"`
						Keywordmodifier struct {
							Text string `xml:",chardata"` // information mapping the P...
						} `xml:"keywordmodifier"`
					} `xml:"keyword"`
				} `xml:"keywordset"`
				Indexclassificationset struct {
					Text                string `xml:",chardata"`
					Indexclassification []struct {
						Text               string `xml:",chardata"` // B8120G, B8110C, C3340H, C...
						Classificationcode string `xml:"classificationcode,attr"`
					} `xml:"indexclassification"`
				} `xml:"indexclassificationset"`
				Treatmentcodeset struct {
					Text          string `xml:",chardata"`
					Treatmentcode []struct {
						Text string `xml:",chardata"` // P, T, P, P, T, P, T, P, P...
					} `xml:"treatmentcode"`
				} `xml:"treatmentcodeset"`
				Numericalindexset struct {
					Text           string `xml:",chardata"`
					Numericalindex []struct {
						Text             string `xml:",chardata"`
						Physicalquantity struct {
							Text string `xml:",chardata"` // size, power, voltage, siz...
						} `xml:"physicalquantity"`
						Numericvalue []struct {
							Text string `xml:",chardata"` // 1.8E-07, 1.0E+04, 1.2E+03...
						} `xml:"numericvalue"`
						Standardunit struct {
							Text string `xml:",chardata"` // m, W, V, m, W, bit/s, Hz,...
						} `xml:"standardunit"`
					} `xml:"numericalindex"`
				} `xml:"numericalindexset"`
				Chemicalindexset struct {
					Text          string `xml:",chardata"`
					Chemicalindex []struct {
						Text     string `xml:",chardata"` // C, C, GaAs, As, Ga, GaAs,...
						Chemrole string `xml:"chemrole,attr"`
					} `xml:"chemicalindex"`
				} `xml:"chemicalindexset"`
				Issuepart struct {
					Text string `xml:",chardata"` // Regular Papers, Regular P...
				} `xml:"issuepart"`
				ArticleQuality struct {
					Text string `xml:",chardata"` // X, X, X, Y, X, X, X, X, X...
				} `xml:"article_quality"`
				Recordsrc struct {
					Text string `xml:",chardata"` // INSPEC, INSPEC, INSPEC, I...
				} `xml:"recordsrc"`
				Csarticleid struct {
					Text string `xml:",chardata"` // 2628a050, 2628a156, 2628a...
				} `xml:"csarticleid"`
				Multimedia struct {
					Text    string `xml:",chardata"`
					Summary struct {
						Text string `xml:",chardata"` // Corner Detection Based Re...
					} `xml:"summary"`
					Compressed struct {
						Text               string `xml:",chardata"`
						Compressedfilename struct {
							Text string `xml:",chardata"` // robio17-359_mm.zip, robio...
						} `xml:"compressedfilename"`
						Compressedfilesize struct {
							Text string `xml:",chardata"` // 7371067, 113891, 6327734,...
						} `xml:"compressedfilesize"`
						Compressiontype struct {
							Text string `xml:",chardata"` // Zip, Zip, Zip, Zip, Zip, ...
						} `xml:"compressiontype"`
						Environmenttype struct {
							Text string `xml:",chardata"` // Windows, Windows, Windows...
						} `xml:"environmenttype"`
						Readmefile struct {
							Text string `xml:",chardata"` // robio17-v359_Readme.txt, ...
						} `xml:"readmefile"`
					} `xml:"compressed"`
					Component []struct {
						Text              string `xml:",chardata"`
						Componentfilename struct {
							Text string `xml:",chardata"` // V02_01nn.pdf, tuffc02fron...
						} `xml:"componentfilename"`
						Componentfilesize struct {
							Text string `xml:",chardata"` // 5069652, 47784788, 658901...
						} `xml:"componentfilesize"`
						Componenttype struct {
							Text string `xml:",chardata"` // PDF, .tif, .mp4, .mp4, .m...
						} `xml:"componenttype"`
						Componentplatform struct {
							Text string `xml:",chardata"` // Windows/MAC, All, All, Al...
						} `xml:"componentplatform"`
						Componentdoi struct {
							Text string `xml:",chardata"` // 10.1109/ISSCC.2008.452304...
						} `xml:"componentdoi"`
						Componenttitle struct {
							Text string `xml:",chardata"` // A 128&[mult]128 Single-Ph...
						} `xml:"componenttitle"`
						Componentdescription struct {
							Text string `xml:",chardata"` // The accompanying paper pr...
						} `xml:"componentdescription"`
						Componentperson []struct {
							Text                     string `xml:",chardata"`
							Componentpersonfirstname struct {
								Text string `xml:",chardata"` // Pengfei, Pengfei, Tomasz ...
							} `xml:"componentpersonfirstname"`
							Componentpersonsurname struct {
								Text string `xml:",chardata"` // Song, Song, Czernuszewicz...
							} `xml:"componentpersonsurname"`
						} `xml:"componentperson"`
					} `xml:"component"`
				} `xml:"multimedia"`
				Articlepubmedid struct {
					Text string `xml:",chardata"` // 17946858, 17946779, 17947...
				} `xml:"articlepubmedid"`
				MeshHeadingList struct {
					Text        string `xml:",chardata"`
					MeshFlag    string `xml:"MeshFlag,attr"`
					MeshHeading []struct {
						Text           string `xml:",chardata"`
						DescriptorName struct {
							Text string `xml:",chardata"` // Animals, Cartilage, Artic...
						} `xml:"DescriptorName"`
						QualifierName struct {
							Text string `xml:",chardata"` // pharmacology, methods, dr...
						} `xml:"QualifierName"`
					} `xml:"MeshHeading"`
				} `xml:"MeshHeadingList"`
				URI struct {
					Text string `xml:",chardata"` // 591-8411 jim.taylor@L-3co...
				} `xml:"uri"`
				Articlestdscope struct {
					Text string `xml:",chardata"` // IEEE Std 1003.1-201x defi...
				} `xml:"articlestdscope"`
				Articlestdpurpose struct {
					Text string `xml:",chardata"` // Several principles guided...
				} `xml:"articlestdpurpose"`
				Fundrefgrp struct {
					Text    string `xml:",chardata"`
					Fundref []struct {
						Text       string `xml:",chardata"`
						FunderName struct {
							Text string `xml:",chardata"` // Google Inc., Eleanor Mile...
						} `xml:"funder_name"`
						FunderID struct {
							Text string `xml:",chardata"` // 10.13039/100006754, 10.13...
						} `xml:"funder_id"`
						GrantNumber []struct {
							Text string `xml:",chardata"` // W911NF-10-2-0022, EFRI-M3...
						} `xml:"grant_number"`
						AgencyName struct {
							Text string `xml:",chardata"` // National Natural Science ...
						} `xml:"agency_name"`
					} `xml:"fundref"`
				} `xml:"fundrefgrp"`
				Articleid struct {
					Text string `xml:",chardata"` // 9800101, 0001901, 0000101...
				} `xml:"articleid"`
				ArticleCopyrightStatement struct {
					Text string `xml:",chardata"` // 2329-9290 © 2018 IEEE. P...
				} `xml:"article_copyright_statement"`
				ArticleGraphicalAbstract struct {
					Text                  string `xml:",chardata"`
					GraphicalAbstractType struct {
						Text string `xml:",chardata"` // graphic, graphic, graphic...
					} `xml:"graphical-abstract-type"`
					GraphicalAbstractFileSize struct {
						Text string `xml:",chardata"` // 58KB, 34KB, 89KB, 84KB, 6...
					} `xml:"graphical-abstract-file-size"`
					GraphicalAbstractSummary struct {
						Text string `xml:",chardata"` // Proposal of a proof-of-co...
					} `xml:"graphical-abstract-summary"`
				} `xml:"article-graphical-abstract"`
				Articlejournaltopicset struct {
					Text                string `xml:",chardata"`
					Articlejournaltopic []struct {
						Text string `xml:",chardata"` // Analog and Mixed Mode Cir...
					} `xml:"articlejournaltopic"`
				} `xml:"articlejournaltopicset"`
				Articleimpactstatement struct {
					Text string `xml:",chardata"` // The proposed PCF can be u...
				} `xml:"articleimpactstatement"`
				Articlelicense struct {
					Text string `xml:",chardata"` // Traditional, Traditional,...
				} `xml:"articlelicense"`
				ArticleLicenseURI struct {
					Text string `xml:",chardata"` // http://www.ieee.org/publi...
				} `xml:"article_license_uri"`
				Articlemancentralid struct {
					Text string `xml:",chardata"` // TSTE-00557-2015, TSTE-007...
				} `xml:"articlemancentralid"`
				FundingDepositGrp struct {
					Text          string `xml:",chardata"`
					DepositAgency struct {
						Text string `xml:",chardata"` // CHORUS, CHORUS, CHORUS, C...
					} `xml:"deposit_agency"`
				} `xml:"funding_deposit_grp"`
				Articleissuetitle struct {
					Text string `xml:",chardata"` // Breakthroughs in Photonic...
				} `xml:"articleissuetitle"`
				Articleissuesubtitle struct {
					Text string `xml:",chardata"` // Z. Niu, K.-C. Chen, S. M....
				} `xml:"articleissuesubtitle"`
			} `xml:"articleinfo"`
			Majortopic struct {
				Text string `xml:",chardata"` // NGNI: Next-Generation Net...
			} `xml:"majortopic"`
		} `xml:"article"`
	} `xml:"volume"`
	Titleabbrev struct {
		Text string `xml:",chardata"` // IET Image Process., IET I...
	} `xml:"titleabbrev"`
	Shorttitle struct {
		Text string `xml:",chardata"` // Electricity distribution ...
	} `xml:"shorttitle"`
	Standardsfamilytitle struct {
		Text string `xml:",chardata"` // IEEE Standard for Local a...
	} `xml:"standardsfamilytitle"`
}
