package genios

import (
	"encoding/xml"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestDocument(t *testing.T) {
	var cases = []struct {
		s   string
		doc Document
		err error
	}{
		{
			`
<Document ID="200101002" IDNAME="NO" DB="XZWF">
<Abstract>n.n.</Abstract>
<Authors><Author>n.n.</Author></Authors>
<Descriptors><Descriptor>n.n.</Descriptor></Descriptors>
<Date>
20010101
</Date>
<Issue>
1-2
</Issue>
<ISSN>
0932-0482
</ISSN>
<Language>n.n.</Language>
<Page>
3
</Page>
<Publication-Title>
ZWF - XXXXXXXXX für wirtschaftlichen XXXXXXXXXXX
</Publication-Title>
<Source>
ZWF
</Source>
<Title>
Multinationale XXXXXXXXX
</Title>
<Text>
Lorem ipsum dolor sit amet, consectetur adipisicing elit. Iusto sit esse tempore repellendus nemo, expedita vitae praesentium, voluptatibus. Illum error distinctio incidunt, magnam autem quisquam cum odio omnis culpa ipsum.
</Text>
<Volume>n.n.</Volume>
<Year>
2001
</Year>
<Copyright>
Alle xxxxxxxxxxxxxxxxxxxxxxxxxxx xxxxxxx.<BR/>www.xxxxxxxx.de
</Copyright>
</Document>
`, Document{
				ID:               "200101002",
				DB:               "XZWF",
				IDNAME:           "NO",
				ISSN:             "\n0932-0482\n",
				Source:           "\nZWF\n",
				PublicationTitle: "\nZWF - XXXXXXXXX für wirtschaftlichen XXXXXXXXXXX\n",
				Title:            "\nMultinationale XXXXXXXXX\n",
				Year:             "\n2001\n",
				RawDate:          "\n20010101\n",
				Volume:           "n.n.",
				Issue:            "\n1-2\n",
				RawAuthors:       []string{"n.n."},
				Language:         "n.n.",
				Abstract:         "n.n.",
				Group:            "",
				Descriptors:      "n.n.",
				Text:             "\nLorem ipsum dolor sit amet, consectetur adipisicing elit. Iusto sit esse tempore repellendus nemo, expedita vitae praesentium, voluptatibus. Illum error distinctio incidunt, magnam autem quisquam cum odio omnis culpa ipsum.\n",
			},
			nil},
	}

	for _, c := range cases {
		var doc Document
		err := xml.Unmarshal([]byte(c.s), &doc)
		if err != c.err {
			t.Errorf("got: %v, want: %v", err, c.err)
		}

		if diff := pretty.Compare(doc, c.doc); diff != "" {
			t.Log(diff)
			t.Errorf("got %+v, want %+v", doc, c.doc)
		}

	}
}
