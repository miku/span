package bytebatch

import (
	"bufio"
	"io"
	"runtime"
	"sync"
)

// ByteBatch is a batch of byte slices.
type ByteBatch struct {
	b [][]byte
}

// NewByteBatch creates a new ByteBatch with a given capacity.
func NewByteBatch(cap int) *ByteBatch {
	return &ByteBatch{b: make([][]byte, 0, cap)}
}

// Add adds an element to the batch.
func (bb *ByteBatch) Add(b []byte) {
	bb.b = append(bb.b, b)
}

// Reset empties this batch.
func (bb *ByteBatch) Reset() {
	bb.b = nil
}

// Size returns the number of elements in the batch.
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

// ByteFunc is a function, that transforms a byte slice into another byte slice (plus error).
type ByteFunc func([]byte) ([]byte, error)

// LineProcessor reads bytes from a reader, feeds them into a ByteFunc and
// writes the result to a writer. The default unit of work is one line.
type LineProcessor struct {
	BatchSize     int
	LineSeparator byte
	NumWorkers    int
	r             io.Reader
	w             io.Writer
	f             ByteFunc
}

// NewLineProcessor creates a new LineProcessor. Default batch size is 10000,
// default line separator is `\n`, default number of workers equals the cpu count.
func NewLineProcessor(r io.Reader, w io.Writer, f ByteFunc) *LineProcessor {
	return &LineProcessor{r: r, w: w, f: f, BatchSize: 10000, LineSeparator: '\n', NumWorkers: runtime.NumCPU()}
}

// Run starts executing workers, crunching trough the input.
func (p LineProcessor) Run() error {

	// wErr signals a worker or writer error. If an error occurs, the items in
	// the queue are still process, just no items are added to the queue. There
	// is only one way to toggle this, from false to true, so we don't care
	// about synchronisation.
	var wErr error

	// worker takes []byte batches from a channel queue, executes f and sends the result to the out channel.
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
			// to avoid checking on each loop, we only check for worker or write errors here
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
