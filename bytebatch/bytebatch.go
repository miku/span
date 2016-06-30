package bytebatch

import (
	"bufio"
	"io"
	"runtime"
	"sync"
)

// BytesBatch is a batch of a byte slices.
type ByteBatch struct {
	b [][]byte
}

// NewByteBatch creates a new BytesBatch with a given capacity.
func NewByteBatch(cap int) *ByteBatch {
	return &ByteBatch{b: make([][]byte, 0, cap)}
}

// Add adds an element to the batch.
func (bb *ByteBatch) Add(b []byte) {
	bb.b = append(bb.b, b)
}

// Reset clear this slice.
func (bb *ByteBatch) Reset() {
	bb.b = bb.b[:0]
}

// Size returns the current size of the batch.
func (bb *ByteBatch) Size() int {
	return len(bb.b)
}

// Slice returns a slice of byte slices.
func (bb *ByteBatch) Slice() [][]byte {
	b := make([][]byte, len(bb.b))
	for i := 0; i < len(bb.b); i++ {
		b[i] = bb.b[i]
	}
	return b
}

// ByteFunc is a func, that transforms a byte slice to another byte slice, and a
// possible error.
type ByteFunc func([]byte) ([]byte, error)

// LineProcessor read bytes from a reader, feed them into a ByteFunc and
// writes the result to a writer. The unit of work is one line.
type LineProcessor struct {
	BatchSize     int
	LineSeparator byte
	NumWorkers    int
	r             io.Reader
	w             io.Writer
	f             ByteFunc
}

// NewLineProcessor creates a new LineProcessor. Default batch size is 10000,
// default line separator is `\n`.
func NewLineProcessor(r io.Reader, w io.Writer, f ByteFunc) *LineProcessor {
	return &LineProcessor{r: r, w: w, f: f, BatchSize: 10000, LineSeparator: '\n', NumWorkers: runtime.NumCPU()}
}

// Run starts workers for executing the given func over the data.
func (p LineProcessor) Run() error {

	// wErr signals a worker or writer error. If an error occurs, the items in
	// the queue are still process, just no items are added to the queue. There
	// is only one way to toggle this, from false to true, so we don't care
	// about synchronisation.
	var wErr error

	// worker takes work from a queue, executes f and sends the result to out.
	worker := func(queue chan [][]byte, out chan []byte, f ByteFunc, wg *sync.WaitGroup) {
		defer wg.Done()
		for batch := range queue {
			for _, b := range batch {
				r, err := f(b)
				if err != nil {
					wErr = err
				}
				out <- r
			}
		}
	}

	// writer buffers writes.
	writer := func(w io.Writer, bc chan []byte, done chan bool) {
		bw := bufio.NewWriter(w)
		for b := range bc {
			if _, err := bw.Write(b); err != nil {
				wErr = err
			}
		}
		bw.Flush()
		done <- true
	}

	queue := make(chan [][]byte)
	out := make(chan []byte)
	done := make(chan bool)

	var wg sync.WaitGroup

	go writer(p.w, out, done)

	for i := 0; i < p.NumWorkers; i++ {
		wg.Add(1)
		go worker(queue, out, p.f, &wg)
	}

	batch := NewByteBatch(p.BatchSize)
	br := bufio.NewReader(p.r)

	for {
		b, err := br.ReadBytes(p.LineSeparator)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		batch.Add(b)
		if batch.Size() == p.BatchSize {
			// to avoid checking on every loop, we only check the for worker or write error here
			if wErr != nil {
				break
			}
			queue <- batch.Slice()
			batch.Reset()
		}
	}

	queue <- batch.Slice()
	batch.Reset()

	close(queue)
	wg.Wait()
	close(out)
	<-done

	return wErr
}
