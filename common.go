package span

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"runtime"
	"sync"

	"github.com/miku/span/finc"
)

const (
	// AppVersion of span package. Commandline tools will show this on -v.
	AppVersion = "0.1.52"
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

// FromXML is like FromXMLSize, with a default batch size of 2000 XML documents.
func FromXML(r io.Reader, name string, decoderFunc XMLDecoderFunc) (chan []Importer, error) {
	return FromXMLSize(r, name, decoderFunc, 2000)
}

// FromXMLSize returns a channel of importable document slices given a reader
// over XML, a name of the XML start element, a XMLDecoderFunc callback that
// deserializes an XML snippet and a batch size. TODO(miku): more idiomatic
// error handling, e.g. over error channel
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

// JSONDecoderFunc turns a string into an importable object.
type JSONDecoderFunc func(s string) (Importer, error)

// FromJSON returns a channel of slices of importable objects.
func FromJSON(r io.Reader, decoder JSONDecoderFunc) (chan []Importer, error) {
	return FromJSONSize(r, decoder, 20000)
}

// FromJSONSize returns a channel of slices of importable objects, given a
// reader, decoder and number of documents to batch.
func FromJSONSize(r io.Reader, decoder JSONDecoderFunc, size int) (chan []Importer, error) {

	worker := func(queue chan []string, out chan Importer, wg *sync.WaitGroup) {
		defer wg.Done()
		for batch := range queue {
			for _, s := range batch {
				doc, err := decoder(s)
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

	queue := make(chan []string)
	out := make(chan Importer)
	done := make(chan bool)

	var wg sync.WaitGroup

	go batcher(out, done)

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(queue, out, &wg)
	}

	reader := bufio.NewReader(r)
	var lines []string
	var i int

	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			i++
			lines = append(lines, line)
			if i == size {
				batch := make([]string, size)
				copy(batch, lines)
				queue <- batch
				lines = lines[:0]
				i = 0
			}
		}
		batch := make([]string, len(lines))
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
