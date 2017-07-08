package genios

// Collections maps shortcut to collection name.
var Collections = map[string]string{
	"fzs":     "Fachzeitschriften",
	"recht":   "Recht",
	"sowi":    "Sozialwissenschaften",
	"wiwi":    "Wirtschaftswissenschaften",
	"psyn":    "Psychologie",
	"technik": "Technik",
}

// DatabasePackageMap maps database names to package names.
var DatabasePackageMap = map[string][]string{
	"CI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CME": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CGMW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BURE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BUBH": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BTME": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BME": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BGWP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BAA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ARAB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AMW": []string{
		"Wirtschaftswissenschaften",
		"Technik",
		"Fachzeitschriften",
	},
	"AMP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AKA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AGRI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VD": []string{
		"Technik",
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SPI": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"SCMW": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"PHAP": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"PC": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"OGR": []string{
		"Technik",
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MUTE": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"MABL": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"LB": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"KSM": []string{
		"Recht",
		"Fachzeitschriften",
	},
	"JRW": []string{
		"Technik",
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"IGMW": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"ICT": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"HTH": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"HREC": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"HBUS": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"GAGI": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"FK": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"ERIP": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"EPPE": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"ENRE": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"EMW": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"EIRB": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"EI": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"DQW": []string{
		"Technik",
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DQ": []string{
		"Technik",
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DEBA": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"DATA": []string{
		"Technik",
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CMW": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"CDM": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"CAVE": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"BRSC": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"BIWI": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"AUPR": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"AUKO": []string{
		"Technik",
		"Fachzeitschriften",
	},
	"LAW": []string{
		"Recht",
		"Fachzeitschriften",
	},
	"FUS": []string{
		"Recht",
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"EAWO": []string{
		"Recht",
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PJP": []string{
		"Psychologie",
		"Fachzeitschriften",
	},
	"PJPR": []string{
		"Psychologie",
		"Fachzeitschriften",
	},
	"PJCP": []string{
		"Psychologie",
		"Fachzeitschriften",
	},
	"JNC": []string{
		"Psychologie",
		"Fachzeitschriften",
	},
	"JBS": []string{
		"Psychologie",
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"IJAR": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"ZAAA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
		"Sozialwissenschaften",
	},
	"POPS": []string{
		"Fachzeitschriften",
		"Psychologie",
		"Sozialwissenschaften",
	},
	"WLB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"TAWI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DZI": []string{
		"Wirtschaftswissenschaften",
		"Sozialwissenschaften",
	},
	"BIKE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BWI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"NDV": []string{
		"Fachzeitschriften",
		"Recht",
		"Sozialwissenschaften",
	},
	"TAI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"HANA": []string{
		"Fachzeitschriften",
	},
	"HAND": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GG": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WRP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"FJSB": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"DLGM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FIM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MAR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WUW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WUV": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"LAK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"NPE": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"ASSC": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WUB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"MAC": []string{
		"Fachzeitschriften",
	},
	"PACK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GITL": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"PCW": []string{
		"Fachzeitschriften",
	},
	"AUON": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DTZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PCH": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"GITS": []string{
		"Fachzeitschriften",
	},
	"AHOR": []string{
		"Fachzeitschriften",
	},
	"SBAN": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KOCA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"THEX": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SOBE": []string{
		"Fachzeitschriften",
		"Psychologie",
		"Sozialwissenschaften",
	},
	"FLWA": []string{
		"Fachzeitschriften",
	},
	"FLWI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WUWE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"JEEM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DSZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SOLI": []string{
		"LIT",
		"Sozialwissenschaften",
	},
	"TRE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DST": []string{
		"Wirtschaftswissenschaften",
	},
	"STBG": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"VDIA": []string{
		"Fachzeitschriften",
	},
	"ZVS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
	},
	"VDIN": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"DSB": []string{
		"Fachzeitschriften",
		"Recht",
		"Technik",
	},
	"WTRE": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"MF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"DBWV": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"VCNE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MT": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"IFOL": []string{
		"Wirtschaftswissenschaften",
	},
	"FS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AFZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WEBE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"CIOD": []string{
		"Fachzeitschriften",
	},
	"WEBR": []string{
		"Fachzeitschriften",
	},
	"FB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FH": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"ZNER": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"FM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SMA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MDA": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"PHAM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BONL": []string{
		"Fachzeitschriften",
	},
	"MDR": []string{
		"Fachzeitschriften",
		"Recht",
	},
	"BOND": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AUA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"AUB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DFI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VEC": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AUW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"REFA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VRAG": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"LMZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"RLH": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MARC": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PUA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
	},
	"IMWI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"SB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ERP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"OMNI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BBLO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CAV": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"IBF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ANKR": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"FERT": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"ABAU": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KOR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"CAD": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"SOFI": []string{
		"Wirtschaftswissenschaften",
		"Sozialwissenschaften",
	},
	"SOFO": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"LI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"HIF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ONR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PROJ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"STB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"ZRP": []string{
		"Fachzeitschriften",
		"Recht",
	},
	"IWPR": []string{
		"Wirtschaftswissenschaften",
	},
	"SOC": []string{
		"Fachzeitschriften",
		"Psychologie",
		"Sozialwissenschaften",
	},
	"EUIB": []string{
		"Fachzeitschriften",
	},
	"UNTE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"STS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ABZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MREV": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PMGI": []string{
		"Fachzeitschriften",
	},
	"BANK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CFO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"QJIA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MAKR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"TAXI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"NEUL": []string{
		"Fachzeitschriften",
	},
	"DKT": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"GA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GHR": []string{
		"Fachzeitschriften",
		"Recht",
	},
	"JFAV": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ECON": []string{
		"LIT",
		"Wirtschaftswissenschaften",
	},
	"KMAG": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MID": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WEFO": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"TRUC": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"SBIL": []string{
		"Fachzeitschriften",
	},
	"JOR": []string{
		"Fachzeitschriften",
		"Recht",
	},
	"AOB": []string{
		"Fachzeitschriften",
		"Recht",
	},
	"FOGR": []string{
		"Technik",
	},
	"PRIN": []string{
		"Technik",
	},
	"AWCH": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"RAU": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"WIME": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
		"Recht",
	},
	"OEWI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"LP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"HBNI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZFGK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"CHX": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BUIM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZFGG": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DBM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DBL": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MAV": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"HORS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VIM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KUK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DBW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PSYT": []string{
		"Psychologie",
	},
	"GMST": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"EXFO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ABES": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DINM": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"BAMA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZWF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CLIN": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PSYA": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"STAA": []string{
		"Fachzeitschriften",
		"Recht",
		"Sozialwissenschaften",
	},
	"WFP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"GAST": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"NGZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CAIT": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"PKV": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"SBRE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FLOW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"RTH": []string{
		"Fachzeitschriften",
		"Recht",
		"Sozialwissenschaften",
	},
	"RTJ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"IP": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"HMD": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"EDBW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZKF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"EM": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"ED": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"EA": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"EB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"RAAK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SHF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"DLT": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"AFP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"BEFO": []string{
		"LIT",
		"Wirtschaftswissenschaften",
	},
	"AMIF": []string{
		"Fachzeitschriften",
	},
	"SSP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"CIPA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"USTB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"OGEM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"IM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PROD": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"PROC": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"HORA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ISR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"PMGC": []string{
		"Fachzeitschriften",
	},
	"CHAN": []string{
		"Fachzeitschriften",
	},
	"IA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"ISI": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"VJH": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"EJOP": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"MUM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"JBM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"EUSO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ASP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"ASW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MCKW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PAFO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"TETE": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"IMC": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"BBL": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"OEKZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BUMT": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"BBA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"IMP": []string{
		"Fachzeitschriften",
	},
	"OEKO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CARD": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"INFO": []string{
		"Technik",
	},
	"BVPE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
	},
	"KEM": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"HVM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SPAO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KE": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"ITG": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"BVPN": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
	},
	"KH": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ITB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"KM": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"AATG": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KU": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SPAR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ITT": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KEP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AAA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZSR": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"SINV": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DD": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AUBI": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"KRES": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"EUFI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ABIL": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SWW": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"BIM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ERBS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"AECO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"UNED": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"JFNS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VERW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"RBS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CWT": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"VERK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GOV": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"GOP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"LGK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"IMMO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"IMMA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"QZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"WJ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"QE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"IHSL": []string{
		"Wirtschaftswissenschaften",
		"Sozialwissenschaften",
	},
	"ZFS": []string{
		"Fachzeitschriften",
		"Psychologie",
		"Sozialwissenschaften",
	},
	"ZFR": []string{
		"Fachzeitschriften",
		"Recht",
		"Sozialwissenschaften",
	},
	"WM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"ZFP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
	},
	"GJHR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
	},
	"COWI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZFT": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"WW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZFO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
	},
	"BUHA": []string{
		"Fachzeitschriften",
	},
	"WP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
	},
	"DTRS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GFKM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"IWB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"BUSR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"HSR": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"IWT": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"JUKA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"RISK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"IAA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"WEN": []string{
		"Fachzeitschriften",
	},
	"IAC": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"PASS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FVW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VALU": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AAS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CEAB": []string{
		"Technik",
	},
	"DVZB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ETEC": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"AEQ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SVER": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PORT": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZOE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
	},
	"PP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"GEW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"HMDS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"GET": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"STIF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BUS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BUW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BUM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MALE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ABMS": []string{
		"Fachzeitschriften",
	},
	"PM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"UMPS": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"DDS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZFAA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"DDH": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FSEA": []string{
		"Fachzeitschriften",
	},
	"WPSY": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
	},
	"EXTR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ORDO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WWT": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"KAFU": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"LMZA": []string{
		"Fachzeitschriften",
	},
	"AUIN": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"ANT": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"ANP": []string{
		"Fachzeitschriften",
	},
	"PSS": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"KURS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BSPI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FSA": []string{
		"Fachzeitschriften",
	},
	"EMAR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FSE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BULK": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"STGU": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GWW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GWP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DVZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"OELA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DVP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FLUI": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"VRR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"THBE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"JCSM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CASH": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VRA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"WAO": []string{
		"LIT",
		"Sozialwissenschaften",
	},
	"CO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"WWGR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ADIW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"LEMO": []string{
		"Fachzeitschriften",
	},
	"CZ": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"MUAR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WAX": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"CS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"CW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"CT": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"BILA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BILN": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AVFB": []string{
		"Fachzeitschriften",
	},
	"GELD": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PLM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ENEV": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"DIM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"LOG": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"IFOK": []string{
		"Wirtschaftswissenschaften",
	},
	"IFOD": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PDKK": []string{
		"Fachzeitschriften",
	},
	"HDLG": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"XPSY": []string{
		"Psychologie",
	},
	"INSP": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"INST": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"IHKK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"IFOS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CBIL": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"DIBA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CBS": []string{
		"Fachzeitschriften",
	},
	"OWG": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"PPS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GEPR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"LEDI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZFWP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZFWU": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"IX": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"AUTO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"IS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VM": []string{
		"Fachzeitschriften",
		"Recht",
	},
	"AUTH": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AUTI": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"VR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VT": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"VV": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KONP": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"MTA": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"KONT": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"METE": []string{
		"Fachzeitschriften",
	},
	"EPRA": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"TW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"HOLZ": []string{
		"Technik",
	},
	"MCDB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"PWI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"BE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"BM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"BI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"EPP": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"BLIS": []string{
		"LIT",
		"Wirtschaftswissenschaften",
		"Sozialwissenschaften",
	},
	"BW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FIWI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KOEL": []string{
		"Wirtschaftswissenschaften",
	},
	"GRER": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MFZB": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"EJCO": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"LMZS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PMEM": []string{
		"Fachzeitschriften",
	},
	"MTME": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"IJPR": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"FAMR": []string{
		"Fachzeitschriften",
		"Recht",
	},
	"LOGI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"VUV": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"BFUP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AKTS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"PUGK": []string{
		"Fachzeitschriften",
		"Psychologie",
		"Sozialwissenschaften",
	},
	"OP": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"IFKP": []string{
		"Fachzeitschriften",
	},
	"HWWA": []string{
		"Wirtschaftswissenschaften",
	},
	"LKH": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KFZB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZFBF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VERS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FLW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DEI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"EEK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"SMF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DIWW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DIWR": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"FLF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KUSI": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"CMAG": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"PUUM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"KUSE": []string{
		"Recht",
	},
	"EUZW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"ZFKO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ABAL": []string{
		"Fachzeitschriften",
	},
	"ZFKE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KUST": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"EWK": []string{
		"Fachzeitschriften",
		"Psychologie",
		"Sozialwissenschaften",
	},
	"DSTZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"SCOP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"LABO": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"CAPI": []string{
		"Fachzeitschriften",
	},
	"OEIM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KONS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GEBM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VOB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
		"Technik",
	},
	"FBIE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AHGZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"UR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"HTT": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AHGA": []string{
		"Fachzeitschriften",
	},
	"FPP": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"SUG": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"ESTE": []string{
		"Technik",
	},
	"ESTB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"SOSY": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"ABKL": []string{
		"Fachzeitschriften",
	},
	"ACQ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ACAS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ZUG": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BLEC": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"AWOE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SOSI": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"CYBI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"TEC": []string{
		"Fachzeitschriften",
	},
	"AUEL": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"JSPP": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"GFGH": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PMJ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FINR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"REND": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ABSC": []string{
		"Fachzeitschriften",
	},
	"MIND": []string{
		"Wirtschaftswissenschaften",
	},
	"DNK": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"QUAL": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"NEEN": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"SJB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"RCUA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"IT": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"ZPMP": []string{
		"Fachzeitschriften",
	},
	"WIM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"LAPR": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"MAMI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"VUNM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WIR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MAMA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"GDI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"TWA": []string{
		"Fachzeitschriften",
	},
	"DIST": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CID": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"PSYC": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"MSR": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"GVPA": []string{
		"Fachzeitschriften",
	},
	"FEMS": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"TWP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Psychologie",
	},
	"JBWW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GVPR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DAE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CFLA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"THPH": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"WIRO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"TT": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"TR": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"HDJ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"TJ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"KFZW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DAR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Recht",
	},
	"CITL": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"REGA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MIET": []string{
		"Fachzeitschriften",
		"Recht",
	},
	"AWPA": []string{
		"Fachzeitschriften",
		"Psychologie",
		"Sozialwissenschaften",
	},
	"HOE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AE": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"DWW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AEZT": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ADCO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"NOTB": []string{
		"Fachzeitschriften",
		"Recht",
	},
	"AO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GLIP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Technik",
	},
	"HOR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AT": []string{
		"Fachzeitschriften",
		"Technik",
	},
	"VST": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"BRAU": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"CATI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PEMA": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"INDB": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"DI": []string{
		"Wirtschaftswissenschaften",
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"DMS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"EAW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"EMRE": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"EU": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"EVIN": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FIME": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"FSM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GDP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GUBU": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"GUIN": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"HILA": []string{
		"Wirtschaftswissenschaften",
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"HLOG": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"IFAM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"IFMW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"INJO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"ISBF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"JEF": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"JPV": []string{
		"Wirtschaftswissenschaften",
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"LHMW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MBP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MBW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"MEMW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"OER": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PHM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PJIM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PJLS": []string{
		"Wirtschaftswissenschaften",
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"PSJ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PVD": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"REGI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"REMW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SAFR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SBP": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SLO": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"TMW": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"UTIL": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WSIR": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"WWON": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"DISK": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"JGSI": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"NEHO": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"PAPE": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"PJSS": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"POPE": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"TDIA": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"ZFF": []string{
		"Fachzeitschriften",
		"Sozialwissenschaften",
	},
	"KTX": []string{
		"Fachzeitschriften",
		"Psychologie",
	},
	"MAON": []string{
		"Technik",
	},
	"HTEC": []string{
		"Technik",
	},
	"TWNE": []string{
		"Wirtschaftswissenschaften",
	},
	"GUG": []string{
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"ELEO": []string{
		"Technik",
	},
	"AKS": []string{
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"ASPA": []string{
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"JER": []string{
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"PGE": []string{
		"Wirtschaftswissenschaften",
	},
	"FINE": []string{
		"Wirtschaftswissenschaften",
	},
	"MUB": []string{
		"Wirtschaftswissenschaften",
	},
	"FMAI": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"TUD": []string{
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"KUCH": []string{
		"Wirtschaftswissenschaften",
	},
	"FOWE": []string{
		"Technik",
	},
	"ENER": []string{
		"Wirtschaftswissenschaften",
	},
	"TIAM": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"SOA": []string{
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"KONZ": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"AUMO": []string{
		"Technik",
	},
	"KUSO": []string{
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"HPRW": []string{
		"Technik",
	},
	"JOBS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"JOS": []string{
		"Wirtschaftswissenschaften",
		"Fachzeitschriften",
	},
	"PERI": []string{
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
	"SOAS": []string{
		"Sozialwissenschaften",
		"Fachzeitschriften",
	},
}
