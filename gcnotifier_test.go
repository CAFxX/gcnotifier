package gcnotifier

import (
	"io/ioutil"
	"runtime"
	"testing"
	"time"
)

func TestAfterGC(t *testing.T) {
	doneCh := make(chan struct{})

	go func() {
		M := &runtime.MemStats{}
		NumGC := uint32(0)
		gcn := New()
		for range gcn.AfterGC() {
			runtime.ReadMemStats(M)
			NumGC += 1
			if NumGC != M.NumGC {
				t.Fatal("Skipped a GC notification")
			}
			if NumGC > 500 {
				gcn.Close()
				gcn.Close() // harmless, just for testing
			}
		}
		doneCh <- struct{}{}
	}()

	for {
		select {
		case <-time.After(1 * time.Millisecond):
			b := make([]byte, 1<<20)
			b[0] = 1
		case <-doneCh:
			return
		}
	}
}

// Example implements a simple time-based buffering io.Writer: data sent over
// dataCh is buffered for up to 100ms, then flushed out in a single call to
// out.Write and the buffer is reused. If GC runs, the buffer is flushed and
// then discarded so that it can be collected during the next GC run. The
// example is necessarily simplistic, a real implementation would be more
// refined (e.g. on GC flush or resize the buffer based on a threshold,
// perform asynchronous flushes, properly signal completions and propagate
// errors, adaptively preallocate the buffer based on the previous capacity,
// etc.)
func Example() {
	dataCh := make(chan []byte)
	flushCh := time.Tick(100 * time.Millisecond)
	doneCh := make(chan struct{})

	out := ioutil.Discard

	go func() {
		var buf []byte

		gcn := New()
		defer gcn.Close()

		for {
			select {
			case data := <-dataCh:
				// received data to write to the buffer
				buf = append(buf, data...)
			case <-flushCh:
				// time to flush the buffer (but reuse it for the next writes)
				out.Write(buf)
				buf = buf[:0]
			case <-gcn.AfterGC():
				// GC just ran: flush and then drop the buffer
				out.Write(buf)
				buf = nil
			case <-doneCh:
				// close the writer: flush the buffer and return
				out.Write(buf)
				return
			}
		}
	}()

	for i := 0; i < 1<<20; i++ {
		dataCh <- make([]byte, 1024)
	}
	doneCh <- struct{}{}
}
