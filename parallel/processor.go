// Package parallel implements helpers for fast processing of line oriented
// inputs. Basic usage example:
//
//	r := strings.NewReader("1\n2\n3\n")
//	f := func(ln int, b []byte) ([]byte, error) {
//	    result := fmt.Sprintf("#%d %s", ln, string(b))
//	    return []byte(result), nil
//	}
//
//	p := parallel.NewProcessor(r, os.Stdout, f)
//	if err := p.Run(); err != nil {
//	    log.Fatal(err)
//	}
//
// This would print out:
//
//	#0 1
//	#1 2
//	#2 3
//
// Note that the order of the input is not guaranteed to be preserved. If you
// care about the exact position, utilize the originating line number passed
// into the transforming function.
package parallel

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"runtime"
	"sync"
)

// outputFlushSize is how much converted output a worker accumulates before
// handing it to the writer. It trades channel handoffs against memory in
// flight, which is bounded by roughly 2*NumWorkers times this value: one
// buffer per worker being filled, plus the buffered out channel. Records are
// never split across a flush.
const outputFlushSize = 1 << 20

// Record groups a value and a corresponding line number.
type Record struct {
	lineno int64
	value  []byte
}

// BytesBatch is a slice of byte slices.
type BytesBatch struct {
	b []Record
}

// NewBytesBatchCapacity creates a new BytesBatch with a given capacity.
func NewBytesBatchCapacity(cap int) *BytesBatch {
	return &BytesBatch{b: make([]Record, 0, cap)}
}

// Add adds an element to the batch.
func (bb *BytesBatch) Add(b Record) {
	bb.b = append(bb.b, b)
}

// Reset empties this batch, keeping the underlying array for reuse. This is
// safe, because Slice returns a copy of the records.
func (bb *BytesBatch) Reset() {
	bb.b = bb.b[:0]
}

// Size returns the number of elements in the batch.
func (bb *BytesBatch) Size() int {
	return len(bb.b)
}

// Slice returns a slice of byte slices.
func (bb *BytesBatch) Slice() []Record {
	b := make([]Record, len(bb.b))
	copy(b, bb.b)
	return b
}

// TransformerFunc takes a line number and a slice of bytes and returns a slice of bytes and a
// an error. A common denominator of functions that transform data.
type TransformerFunc func(lineno int64, b []byte) ([]byte, error)

// Processor can process lines in parallel.
type Processor struct {
	BatchSize        int
	RecordSeparator  byte
	NumWorkers       int
	SkipEmptyLines   bool
	BatchMemoryLimit int64
	r                io.Reader
	w                io.Writer
	f                TransformerFunc
}

// NewProcessor creates a new line processor, which reads lines from a reader,
// applies a function and writes results back to a writer.
func NewProcessor(r io.Reader, w io.Writer, f TransformerFunc) *Processor {
	return &Processor{
		BatchSize:        10000,
		RecordSeparator:  '\n',
		NumWorkers:       runtime.NumCPU(),
		SkipEmptyLines:   true,
		BatchMemoryLimit: 34359738368, // 32GB
		r:                r,
		w:                w,
		f:                f,
	}
}

// RunWorkers allows to quickly set the number of workers.
func (p *Processor) RunWorkers(numWorkers int) error {
	p.NumWorkers = numWorkers
	return p.Run()
}

// Run starts the workers, crunching through the input.
func (p *Processor) Run() error {
	// procErr records the first worker or writer error. If an error occurs,
	// items already in the queue are still processed, just no new batches are
	// enqueued. Access is guarded by a mutex, the producer only checks it
	// between batches.
	var (
		mu      sync.Mutex
		procErr error
	)
	setErr := func(err error) {
		mu.Lock()
		if procErr == nil {
			procErr = err
		}
		mu.Unlock()
	}
	getErr := func() error {
		mu.Lock()
		defer mu.Unlock()
		return procErr
	}
	// The worker fetches items from a queue, executes f and sends the results
	// to the out channel. Results are accumulated into a buffer and sent in
	// one piece: roughly one handoff per outputFlushSize of output rather
	// than one per record, which is what input batching is for. On error,
	// that one result is dropped.
	worker := func(queue chan []Record, out chan []byte, f TransformerFunc, wg *sync.WaitGroup) {
		defer wg.Done()
		// The buffer is handed to the writer, so a new one is allocated after
		// each flush rather than reused.
		newBuf := func() *bytes.Buffer {
			return bytes.NewBuffer(make([]byte, 0, outputFlushSize+outputFlushSize/8))
		}
		buf := newBuf()
		for batch := range queue {
			for _, record := range batch {
				r, err := f(record.lineno, record.value)
				if err != nil {
					setErr(err)
					continue
				}
				buf.Write(r)
				// Flush on size, not on batch boundaries: a batch is a count
				// of records, so buffering a whole one would make memory in
				// flight scale with BatchSize times the size of a converted
				// record, which for SOLR documents is a lot.
				if buf.Len() >= outputFlushSize {
					out <- buf.Bytes()
					buf = newBuf()
				}
			}
		}
		if buf.Len() > 0 {
			out <- buf.Bytes()
		}
	}
	// The writer collects and buffers writes.
	writer := func(w io.Writer, bc chan []byte, done chan bool) {
		bw := bufio.NewWriter(w)
		for b := range bc {
			if _, err := bw.Write(b); err != nil {
				setErr(err)
			}
		}
		if err := bw.Flush(); err != nil {
			setErr(err)
		}
		done <- true
	}
	// out is buffered so that a worker finishing a batch does not block on the
	// single writer goroutine. queue stays unbuffered on purpose: buffering it
	// would let the producer enqueue further batches before it notices a
	// transform error, widening the partial output a failed run leaves behind.
	var (
		queue = make(chan []Record)
		out   = make(chan []byte, p.NumWorkers)
		done  = make(chan bool)
		wg    sync.WaitGroup
	)
	go writer(p.w, out, done)
	for i := 0; i < p.NumWorkers; i++ {
		wg.Add(1)
		go worker(queue, out, p.f, &wg)
	}
	var (
		batch      = NewBytesBatchCapacity(p.BatchSize)
		br         = bufio.NewReader(p.r)
		i          int64
		batchBytes int64
	)
	for {
		b, readErr := br.ReadBytes(p.RecordSeparator)
		if readErr != nil && readErr != io.EOF {
			return readErr
		}
		// A final line without a trailing separator arrives together with
		// io.EOF and is processed like any other line.
		if !(p.SkipEmptyLines && len(bytes.TrimSpace(b)) == 0) {
			batch.Add(Record{lineno: i, value: b})
			batchBytes += int64(len(b))
			i++
			if batch.Size() == p.BatchSize || batchBytes > p.BatchMemoryLimit {
				if batchBytes > p.BatchMemoryLimit {
					log.Printf("trim batch to %d, exceeding memory limit %d", batch.Size(), batchBytes)
				}
				// To avoid checking on each line, we only check for worker or
				// write errors here.
				if getErr() != nil {
					break
				}
				queue <- batch.Slice()
				batch.Reset()
				batchBytes = 0
			}
		}
		if readErr == io.EOF {
			break
		}
	}
	if batch.Size() > 0 && getErr() == nil {
		queue <- batch.Slice()
		batch.Reset()
	}
	close(queue)
	wg.Wait()
	close(out)
	<-done
	return getErr()
}
