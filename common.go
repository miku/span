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
package span

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"
	"sync"

	"github.com/cloudfoundry/gosigar"
	"github.com/miku/span/finc"
)

const (
	// AppVersion of span package. Commandline tools will show this on -v.
	AppVersion = "0.1.110"
	// KeyLengthLimit is a limit imposed by memcached protocol, which is used
	// for blob storage as of June 2015. If we change the key value store,
	// this limit might become obsolete.
	KeyLengthLimit = 250
)

// Skip marks records to skip.
type Skip struct {
	Reason string
}

// Error returns the reason for skipping.
func (s Skip) Error() string {
	return fmt.Sprintf("SKIP %s", s.Reason)
}

// Importer objects can be converted into an intermediate schema.
type Importer interface {
	ToIntermediateSchema() (*finc.IntermediateSchema, error)
}

// Source can emit records given a reader. The channel is of type []Importer,
// to allow the source to send objects over the channel in batches for
// performance (1000 x 1000 docs vs 1000000 x 1 doc).
type Source interface {
	Iterate(io.Reader) (<-chan []Importer, error)
}

// XMLDecoderFunc returns an importable document, given an XML decoder and a
// start element.
type XMLDecoderFunc func(*xml.Decoder, xml.StartElement) (Importer, error)

// FromXML is like FromXMLSize, with a default batch size of 2000 XML
// documents.
func FromXML(r io.Reader, name string, decoderFunc XMLDecoderFunc) (chan []Importer, error) {
	mem := sigar.Mem{}
	if err := mem.Get(); err != nil {
		return FromXMLSize(r, name, decoderFunc, 500)
	}
	switch {
	default:
		return FromXMLSize(r, name, decoderFunc, 2000)
	case mem.Free < 1048576:
		return FromXMLSize(r, name, decoderFunc, 500)
	case mem.Free < 2097152:
		return FromXMLSize(r, name, decoderFunc, 1000)
	}
}

// FromXMLSize returns a channel of importable document slices given a reader
// over XML, a name of the XML start element, a XMLDecoderFunc callback that
// deserializes an XML snippet and a batch size. TODO(miku): more idiomatic
// error handling, e.g. over error channel.
func FromXMLSize(r io.Reader, name string, decoderFunc XMLDecoderFunc, size int) (chan []Importer, error) {
	ch := make(chan []Importer)

	var i int
	var docs []Importer

	go func() {
		decoder := xml.NewDecoder(bufio.NewReader(r))
		for {
			t, _ := decoder.Token()
			if t == nil {
				break
			}
			switch se := t.(type) {
			case xml.StartElement:
				if se.Name.Local == name {
					doc, err := decoderFunc(decoder, se)
					if err != nil {
						log.Fatal(err)
					}
					i++
					docs = append(docs, doc)
					if i == size {
						batch := make([]Importer, len(docs))
						copy(batch, docs)
						ch <- batch
						docs = docs[:0]
						i = 0
					}
				}
			}
		}
		batch := make([]Importer, len(docs))
		copy(batch, docs)
		ch <- batch
		close(ch)
	}()
	return ch, nil
}

// ImporterFunc turns a byte slice into a single importable object.
type ImporterFunc func(b []byte) (Importer, error)

// FromLines returns a channel of slices of importable objects with a default
// batch size of 20000 docs.
func FromLines(r io.Reader, f ImporterFunc) (chan []Importer, error) {
	mem := sigar.Mem{}
	if err := mem.Get(); err != nil {
		return FromLinesSize(r, f, 2000)
	}
	switch {
	default:
		return FromLinesSize(r, f, 20000)
	case mem.Free < 1048576:
		return FromLinesSize(r, f, 2000)
	case mem.Free < 2097152:
		return FromLinesSize(r, f, 5000)
	}
}

// FromLinesSize returns a channel of slices of importable values, given a
// reader, f (for a single value) and number of documents to batch.
// Important: Due to fan-out input and output order will not be preserved.
func FromLinesSize(r io.Reader, f ImporterFunc, size int) (chan []Importer, error) {

	worker := func(queue chan [][]byte, out chan Importer, wg *sync.WaitGroup) {
		defer wg.Done()
		for batch := range queue {
			for _, s := range batch {
				doc, err := f(s)
				if err != nil {
					log.Fatal(err)
				}
				out <- doc
			}
		}
	}

	ch := make(chan []Importer)

	batcher := func(in chan Importer, done chan bool) {
		var docs []Importer
		var i int
		for doc := range in {
			docs = append(docs, doc)
			i++
			if i == size {
				batch := make([]Importer, size)
				copy(batch, docs)
				ch <- batch
				docs = docs[:0]
				i = 0
			}
		}
		batch := make([]Importer, len(docs))
		copy(batch, docs)
		ch <- batch
		done <- true
	}

	queue := make(chan [][]byte)
	out := make(chan Importer)
	done := make(chan bool)

	var wg sync.WaitGroup

	go batcher(out, done)

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(queue, out, &wg)
	}

	reader := bufio.NewReader(r)
	var lines [][]byte
	var i int

	go func() {
		for {
			line, err := reader.ReadBytes('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			if strings.TrimSpace(string(line)) == "" {
				continue
			}
			i++
			lines = append(lines, line)
			if i == size {
				batch := make([][]byte, size)
				copy(batch, lines)
				queue <- batch
				lines = lines[:0]
				i = 0
			}
		}
		batch := make([][]byte, len(lines))
		copy(batch, lines)
		queue <- batch

		close(queue)
		wg.Wait()
		close(out)
		<-done
		close(ch)
	}()

	return ch, nil
}
