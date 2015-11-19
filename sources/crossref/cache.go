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
package crossref

import "sync"

// IntStringCache for int keys and string values with a thread-safe setter.
// TODO(miku): move to something more generic
type IntStringCache struct {
	mu      *sync.RWMutex
	Entries map[int]string
}

// NewIntStringCache creates a new in memory members cache.
func NewIntStringCache() IntStringCache {
	return IntStringCache{Entries: make(map[int]string), mu: new(sync.RWMutex)}
}

// Set sets the string value for an int key, threadsafe.
func (c *IntStringCache) Set(k int, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Entries[k] = v
}
