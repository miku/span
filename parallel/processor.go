// Package parallel implements helpers for fast processing of line oriented
// inputs. Basic usage example:
//
//     r := strings.NewReader("1\n2\n3\n")
//     f := func(ln int, b []byte) ([]byte, error) {
//         result := fmt.Sprintf("#%d %s", ln, string(b))
//         return []byte(result), nil
//     }
//
//     p := parallel.NewProcessor(r, os.Stdout, f)
//     if err := p.Run(); err != nil {
//         log.Fatal(err)
//     }
//
// This would print out:
//
//     #0 1
//     #1 2
//     #2 3
//
// Note that the order of the input is not guaranteed to be preserved. If you
// care about the exact position, utilize the originating line number passed
// into the transforming function.
package parallel

import (
	"bufio"
	"bytes"
	"io"
	"runtime"
	"sync"
)

// Record groups a value and a corresponding line number.
type Record struct {
	lineno int64
	value  []byte
}

// BytesBatch is a slice of byte slices.
type BytesBatch struct {
	b []Record
}

// NewBytesBatch creates a new BytesBatch with a given capacity.
func NewBytesBatch() *BytesBatch {
	return NewBytesBatchCapacity(0)
}

// NewBytesBatchCapacity creates a new BytesBatch with a given capacity.
func NewBytesBatchCapacity(cap int) *BytesBatch {
	return &BytesBatch{b: make([]Record, 0, cap)}
}

// Add adds an element to the batch.
func (bb *BytesBatch) Add(b Record) {
	bb.b = append(bb.b, b)
}

// Reset empties this batch.
func (bb *BytesBatch) Reset() {
	bb.b = nil
}

// Size returns the number of elements in the batch.
func (bb *BytesBatch) Size() int {
	return len(bb.b)
}

// Slice returns a slice of byte slices.
func (bb *BytesBatch) Slice() []Record {
	b := make([]Record, len(bb.b))
	for i := 0; i < len(bb.b); i++ {
		b[i] = bb.b[i]
	}
	return b
}

// TransformerFunc takes a line number and a slice of bytes and returns a slice of bytes and a
// an error. A common denominator of functions that transform data.
type TransformerFunc func(lineno int64, b []byte) ([]byte, error)

// Processor can process lines in parallel.
type Processor struct {
	BatchSize       int
	RecordSeparator byte
	NumWorkers      int
	SkipEmptyLines  bool
	r               io.Reader
	w               io.Writer
	f               TransformerFunc
}

// NewProcessor creates a new line processor, which reads lines from a reader,
// applies a function and writes results back to a writer.
func NewProcessor(r io.Reader, w io.Writer, f TransformerFunc) *Processor {
	return &Processor{
		BatchSize:       10000,
		RecordSeparator: '\n',
		NumWorkers:      runtime.NumCPU(),
		SkipEmptyLines:  true,
		r:               r,
		w:               w,
		f:               f,
	}
}

// RunWorkers allows to quickly set the number of workers.
func (p *Processor) RunWorkers(numWorkers int) error {
	p.NumWorkers = numWorkers
	return p.Run()
}

// Run starts the workers, crunching through the input.
func (p *Processor) Run() error {

	// wErr signals a worker or writer error. If an error occurs, the items in
	// the queue are still process, just no items are added to the queue. There
	// is only one way to toggle this, from false to true, so we don't care
	// about synchronisation.
	var wErr error

	// The worker fetches items from a queue, executes f and sends the result to the out channel.
	worker := func(queue chan []Record, out chan []byte, f TransformerFunc, wg *sync.WaitGroup) {
		defer wg.Done()
		for batch := range queue {
			for _, record := range batch {
				r, err := f(record.lineno, record.value)
				if err != nil {
					wErr = err
				}
				out <- r
			}
		}
	}

	// The writer collects and buffers writes.
	writer := func(w io.Writer, bc chan []byte, done chan bool) {
		bw := bufio.NewWriter(w)
		for b := range bc {
			if _, err := bw.Write(b); err != nil {
				wErr = err
			}
		}
		if err := bw.Flush(); err != nil {
			wErr = err
		}
		done <- true
	}

	queue := make(chan []Record)
	out := make(chan []byte)
	done := make(chan bool)

	var wg sync.WaitGroup

	go writer(p.w, out, done)

	for i := 0; i < p.NumWorkers; i++ {
		wg.Add(1)
		go worker(queue, out, p.f, &wg)
	}

	batch := NewBytesBatchCapacity(p.BatchSize)
	br := bufio.NewReader(p.r)
	var i int64

	for {
		b, err := br.ReadBytes(p.RecordSeparator)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(bytes.TrimSpace(b)) == 0 && p.SkipEmptyLines {
			continue
		}
		batch.Add(Record{lineno: i, value: b})
		if batch.Size() == p.BatchSize {
			// To avoid checking on each loop, we only check for worker or write errors here.
			if wErr != nil {
				break
			}
			queue <- batch.Slice()
			batch.Reset()
		}
		i++
	}

	queue <- batch.Slice()
	batch.Reset()

	close(queue)
	wg.Wait()
	close(out)
	<-done

	return wErr
}
