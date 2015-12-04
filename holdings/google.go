//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                    The Finc Authors, http://finc.info
//                    Martin Czygan, <martin.czygan@uni-leipzig.de>
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
package holdings

// Item is the main google scholar holdings container
type Item struct {
	Title string     `xml:"title"`
	ISSN  string     `xml:"issn"`
	Covs  []Coverage `xml:"coverage"`
}

// Coverage contains coverage information for an item
type Coverage struct {
	FromYear         int    `xml:"from>year"`
	FromVolume       int    `xml:"from>volume"`
	FromIssue        int    `xml:"from>issue"`
	ToYear           int    `xml:"to>year"`
	ToVolume         int    `xml:"to>volume"`
	ToIssue          int    `xml:"to>issue"`
	Comment          string `xml:"comment"`
	DaysNotAvailable int    `xml:"embargo>days_not_available"`
}
