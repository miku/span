package marc

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// Record for MARC-XML data.
type Record struct {
	XMLName xml.Name `xml:"record"`
	Text    string   `xml:",chardata"`
	Header  struct {
		Text       string `xml:",chardata"`
		Status     string `xml:"status,attr"`
		Identifier struct {
			Text string `xml:",chardata"` // oai:gesis.izsoz.de:docume...
		} `xml:"identifier"`
		Datestamp struct {
			Text string `xml:",chardata"` // 2012-08-29T21:40:31Z, 201...
		} `xml:"datestamp"`
		SetSpec []struct {
			Text string `xml:",chardata"` // com_community_10100, com_...
		} `xml:"setSpec"`
	} `xml:"header"`
	Metadata struct {
		Text   string `xml:",chardata"`
		Record struct {
			Text           string `xml:",chardata"`
			Xmlns          string `xml:"xmlns,attr"`
			Doc            string `xml:"doc,attr"`
			Xalan          string `xml:"xalan,attr"`
			Xsi            string `xml:"xsi,attr"`
			SchemaLocation string `xml:"schemaLocation,attr"`
			Leader         struct {
				Text string `xml:",chardata"` // 00000nam a2200000 u 4500,...
			} `xml:"leader"`
			Controlfield []struct {
				Text string `xml:",chardata"` // 20080514135900.0, cr|||||...
				Tag  string `xml:"tag,attr"`
			} `xml:"controlfield"`
			Datafield []struct {
				Text     string `xml:",chardata"`
				Ind2     string `xml:"ind2,attr"`
				Ind1     string `xml:"ind1,attr"`
				Tag      string `xml:"tag,attr"`
				Subfield []struct {
					Text string `xml:",chardata"` // b, http://www.ssoar.info/...
					Code string `xml:"code,attr"`
				} `xml:"subfield"`
			} `xml:"datafield"`
		} `xml:"record"`
	} `xml:"metadata"`
	About struct {
		Text string `xml:",chardata"`
	} `xml:"about"`
}

func (r Record) MustGetControlField(tag string) string {
	v, err := r.GetControlField(tag)
	if err != nil {
		panic(err)
	}
	return v
}

func (r Record) GetControlField(tag string) (string, error) {
	for _, f := range r.Metadata.Record.Controlfield {
		if f.Tag == tag {
			return f.Text, nil
		}
	}
	return "", fmt.Errorf("undefined tag: %s", tag)
}

func (r Record) MustGetFirstDataField(spec string) string {
	value, err := r.GetFirstDataField(spec)
	if err != nil {
		panic(err)
	}
	return value
}

func (r Record) GetFirstDataField(spec string) (string, error) {
	values, err := r.GetDataFields(spec)
	if err != nil {
		return "", err
	}
	if len(values) == 0 {
		return "", nil
	}
	return values[0], nil
}

func (r Record) MustGetDataFields(spec string) []string {
	result, err := r.GetDataFields(spec)
	if err != nil {
		panic(err)
	}
	return result
}

func (r Record) GetDataFields(spec string) (result []string, err error) {
	parts := strings.Split(spec, ".")
	if len(parts) != 2 {
		return result, fmt.Errorf("spec must be of the form tag.subfield, like 245.a")
	}
	tag, subfield := parts[0], parts[1]
	for _, f := range r.Metadata.Record.Datafield {
		if f.Tag == tag {
			for _, sf := range f.Subfield {
				if sf.Code == subfield {
					result = append(result, sf.Text)
				}
			}
		}
	}
	return result, nil
}
