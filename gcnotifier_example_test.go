package gcnotifier

import (
	"io"
	"io/ioutil"
	"sync"
	"time"
)

// BufferingWriter implements a simple time-based buffering io.Writer: data
// written to it is buffered for up to 100ms, then flushed out in a single call
// to out.Write and the buffer is reused. If GC runs, the buffer is flushed and
// then discarded so that it can be collected during the next GC run. The
// example is necessarily simplistic, a real implementation would be more
// refined (e.g. on GC flush or resize the buffer based on a threshold,
// perform asynchronous flushes, properly signal completions and propagate
// errors, adaptively preallocate the buffer based on the previous capacity,
// etc.)
type BufferingWriter struct {
	l sync.Mutex
	out io.Writer
	buf []byte
	flushEvery time.Duration
	closeCh chan struct{}
}

func NewBufferingWriter(out io.Writer, flushEvery time.Duration) *BufferingWriter {
	w := &BufferingWriter{
		closeCh: make(chan struct{}),
		out: out,
		flushEvery: flushEvery,
	}
	go w.run()
	return w
}

func (w *BufferingWriter) run() {
	flushCh := time.Tick(w.flushEvery)

	gcn := New()
	defer gcn.Close()

	for {
		select {
		case <-flushCh:
			// time to flush the buffer (but reuse it for the next writes)
			w.flush(true)
		case <-gcn.AfterGC():
			// GC just ran: flush and then drop the buffer
			w.flush(false)
		case <-w.closeCh:
			return
		}
	}
}

func (w *BufferingWriter) flush(reuse bool) {
	w.l.Lock()
	if len(w.buf) > 0 {
		w.out.Write(w.buf)
	}
	if reuse {
		w.buf = w.buf[:0]
	} else {
		w.buf = nil
	}
	w.l.Unlock()
}

func (w *BufferingWriter) Write(buf []byte) (int, error) {
	w.l.Lock()
	w.buf = append(w.buf, buf...)
	w.l.Unlock()
	return len(buf), nil
}

func (w *BufferingWriter) Close() {
	w.closeCh <- struct{}{}
	w.flush(false)
}

func Example() {
	out := NewBufferingWriter(ioutil.Discard, 100*time.Millisecond)

	for i := 0; i < 1<<20; i++ {
		out.Write(make([]byte, 1024))
	}

	out.Close()
}
