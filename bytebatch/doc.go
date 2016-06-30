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
// Package bytebatch implements a []byte batch and parallel processing
// facilities for line oriented input.
//
// Basic usage:
//
//     p := bytebatch.NewLineProcessor(os.Stdin, os.Stdout, func(b []byte) ([]byte, error) {
//         // implement your business logic here
//         return b, nil
//     })
//     if err := p.Run(); err != nil {
//         // handle error
//     }
//
// More self contained examples (parallel wc -l, grep -F):
//
// * https://gist.github.com/miku/0687690c68e6104893a7c328d2900ca8.
//
package bytebatch
