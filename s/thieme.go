package s

import (
	"strings"

	"github.com/miku/span/s/oai"
)

// ThiemeRecord flavored OAI record.
type ThiemeRecord struct {
	oai.DublinCollection
}

// Title returns the title of the record.
func (r ThiemeRecord) Title() string {
	ts := r.Metadata.Collection.DC.Title
	if len(ts) == 0 {
		return ""
	}
	return strings.Join(ts, " ")
}

// Authors returns a list of authors.
func (r ThiemeRecord) Authors() (authors []string) {
	for _, field := range r.Metadata.Collection.DC.Creator {
		f := strings.TrimSpace(field)
		if f == "" {
			continue
		}
		authors = append(authors, f)
	}
	return
}

// Subjects returns zero or more subjects.
func (r ThiemeRecord) Subjects() (ss []string) {
	for _, field := range r.Metadata.Collection.DC.Subject {
		for _, value := range strings.Split(field, ",") {
			ss = append(ss, value)
		}
	}
	return
}

// DOI returns the document object identifier.
func (r ThiemeRecord) DOI() string {
	for _, s := range r.Metadata.Collection.DC.Identifier {
		if strings.HasPrefix("10.", s) {
			return s
		}
	}
	return ""
}

// URN returns the Uniform Resource Name.
func (r ThiemeRecord) URN() string {
	for _, s := range r.Metadata.Collection.DC.Identifier {
		value := strings.ToLower(s)
		if strings.HasPrefix(value, "urn:") {
			return value
		}
	}
	return ""
}

// Languages returns three letter language strings.
func (r ThiemeRecord) Languages() []string {
	return r.Metadata.Collection.DC.Language
}

// Date returns the publishing date in ISO format.
func (r ThiemeRecord) Date() string {
	// TODO: implement date parsing.
	return r.Metadata.Collection.DC.Date[0]
}

// Publisher returns the publisher.
func (r ThiemeRecord) Publisher() string {
	parts := strings.Split(":", r.Metadata.Collection.DC.Publisher[0])
	return strings.TrimRight(parts[1], ";")
}

// Location of publisher.
func (r ThiemeRecord) Location() string {
	parts := strings.Split(":", r.Metadata.Collection.DC.Publisher[0])
	return strings.TrimSpace(parts[0])
}

// URL to document.
func (r ThiemeRecord) URL() string {
	return r.Metadata.Collection.DC.Relation[0]
}
