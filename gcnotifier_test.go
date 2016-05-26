package gcnotifier

import (
	"runtime"
	"time"
	"testing"
)

func TestAfterGC(t *testing.T) {
	doneCh := make(chan struct{})

	go func() {
		M := &runtime.MemStats{}
		NumGC := uint32(0)
		gcCh := AfterGC()
		for range gcCh {
			runtime.ReadMemStats(M)
			NumGC += 1
			if NumGC != M.NumGC {
				t.Fatal("Skipped a GC notification")
			}
			if NumGC > 500 {
				doneCh <- struct{}{}
				close(gcCh)
				return
			}
		}
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
