//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                 by The Finc Authors, http://finc.info
//                 by Martin Czygan, <martin.czygan@uni-leipzig.de>
//
// This file is part of some open source application.
//
// Some open source application is free software: you can redistribute
// it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation, either
// version 3 of the License, or (at your option) any later version.
//
// Some open source application is distributed in the hope that it will
// be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
//
// @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
//
package exporter

import (
	"github.com/kennygrant/sanitize"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
)

// Solr4vu13v3 is like Solr4vu13v2, only the HierarchyParentTitle won't be
// filled with the journal title.
type Solr4Vufind13v3 struct {
	Solr4Vufind13v2
}

// Attach attaches the ISILs to a record.
func (s *Solr4Vufind13v3) Attach(isils []string) {
	s.Institutions = isils
}

// Export method from intermediate schema to solr 4/13 schema.
func (s *Solr4Vufind13v3) Convert(is finc.IntermediateSchema) error {
	s.Allfields = is.Allfields()
	s.Formats = append(s.Formats, is.Format)
	s.Fullrecord = "blob:" + is.RecordID
	s.Fulltext = is.Fulltext
	s.ID = is.RecordID
	s.Imprint = is.Imprint()
	s.ISSN = is.ISSNList()
	s.MegaCollections = append(s.MegaCollections, is.MegaCollection)
	s.PublishDateSort = is.Date.Year()
	s.Publishers = is.Publishers
	s.RecordType = finc.AIRecordType
	s.Series = append(s.Series, is.JournalTitle)
	s.SourceID = is.SourceID
	s.Subtitle = is.ArticleSubtitle
	s.TitleSort = is.SortableTitle()
	s.Topics = is.Subjects
	s.URL = is.URL

	classes := container.NewStringSet()
	for _, s := range is.Subjects {
		for _, class := range SubjectMapping.LookupDefault(s, []string{}) {
			classes.Add(class)
		}
	}
	s.FincClassFacet = classes.Values()

	sanitized := sanitize.HTML(is.ArticleTitle)
	s.Title, s.TitleFull, s.TitleShort = sanitized, sanitized, sanitized

	for _, lang := range is.Languages {
		s.Languages = append(s.Languages, LanguageMap.LookupDefault(lang, lang))
	}

	for _, author := range is.Authors {
		s.SecondaryAuthors = append(s.SecondaryAuthors, author.String())
		s.AuthorFacet = append(s.AuthorFacet, author.String())
	}

	if len(s.SecondaryAuthors) > 0 {
		s.Author = s.SecondaryAuthors[0]
	}

	s.AccessFacet = AIAccessFacet

	// site specific formats
	s.FormatDe105 = []string{FormatDe105.LookupDefault(is.Format, "")}
	s.FormatDe14 = []string{FormatDe14.LookupDefault(is.Format, "")}
	s.FormatDe15 = []string{FormatDe15.LookupDefault(is.Format, "")}
	s.FormatDe520 = []string{FormatDe520.LookupDefault(is.Format, "")}
	s.FormatDe540 = []string{FormatDe540.LookupDefault(is.Format, "")}
	s.FormatDeCh1 = []string{FormatDeCh1.LookupDefault(is.Format, "")}
	s.FormatDed117 = []string{FormatDed117.LookupDefault(is.Format, "")}
	s.FormatDeGla1 = []string{FormatDeGla1.LookupDefault(is.Format, "")}
	s.FormatDel152 = []string{FormatDel152.LookupDefault(is.Format, "")}
	s.FormatDel189 = []string{FormatDel189.LookupDefault(is.Format, "")}
	s.FormatDeZi4 = []string{FormatDeZi4.LookupDefault(is.Format, "")}
	s.FormatDeZwi2 = []string{FormatDeZwi2.LookupDefault(is.Format, "")}

	s.ContainerVolume = is.Volume
	s.ContainerIssue = is.Issue
	s.ContainerStartPage = is.StartPage
	s.ContainerTitle = is.JournalTitle

	return nil
}
